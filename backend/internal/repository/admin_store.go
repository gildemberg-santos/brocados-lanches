package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// AdminUpdateStoreSettings grava o documento único de configuração e a auditoria.
// O upsert mantém a API utilizável mesmo se o seed não tiver criado o documento;
// createdAt só é definido na primeira gravação.
func (r *CatalogRepository) AdminUpdateStoreSettings(ctx context.Context, settings domain.StoreSettings, audit domain.AuditLog) error {
	document := newStoreSettingsDocument(settings)

	write := func(writeCtx context.Context) error {
		update := bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "name", Value: document.Name},
				{Key: "isOpen", Value: document.IsOpen},
				{Key: "openingHours", Value: document.OpeningHours},
				{Key: "deliveryTime", Value: document.DeliveryTime},
				{Key: "whatsapp", Value: document.WhatsApp},
				{Key: "instagram", Value: document.Instagram},
				{Key: "address", Value: document.Address},
				{Key: "updatedAt", Value: document.UpdatedAt.UTC()},
			}},
			{Key: "$setOnInsert", Value: bson.D{{Key: "createdAt", Value: document.UpdatedAt.UTC()}}},
		}
		opts := options.UpdateOne().SetUpsert(true)
		if _, err := r.storeSettings.UpdateOne(writeCtx, bson.D{{Key: "_id", Value: domain.StoreSettingsID}}, update, opts); err != nil {
			return fmt.Errorf("update store settings: %w", err)
		}
		return nil
	}

	return r.withStoreSettingsTransaction(ctx, write, audit)
}

// AdminUpdateWebhookSettings grava somente o subdocumento webhook, criando o
// documento único quando ele ainda não existe. Assim um PATCH da loja não
// sobrescreve a integração e um PATCH da integração não restaura campos
// operacionais lidos antes de uma mutação concorrente.
func (r *CatalogRepository) AdminUpdateWebhookSettings(ctx context.Context, settings domain.StoreSettings, audit domain.AuditLog) error {
	document := newStoreSettingsDocument(settings)

	write := func(writeCtx context.Context) error {
		update := bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "webhook", Value: document.Webhook},
				{Key: "updatedAt", Value: document.UpdatedAt.UTC()},
			}},
			{Key: "$setOnInsert", Value: bson.D{{Key: "createdAt", Value: document.UpdatedAt.UTC()}}},
		}
		opts := options.UpdateOne().SetUpsert(true)
		if _, err := r.storeSettings.UpdateOne(writeCtx, bson.D{{Key: "_id", Value: domain.StoreSettingsID}}, update, opts); err != nil {
			return fmt.Errorf("update webhook settings: %w", err)
		}
		return nil
	}

	return r.withStoreSettingsTransaction(ctx, write, audit)
}

// AdminUpdatePixSettings grava somente o subdocumento pix, criando o documento
// único quando ele ainda não existe. Um PATCH da loja ou do webhook não pode
// sobrescrever a chave configurada por outra requisição.
func (r *CatalogRepository) AdminUpdatePixSettings(ctx context.Context, settings domain.StoreSettings, audit domain.AuditLog) error {
	document := newStoreSettingsDocument(settings)

	write := func(writeCtx context.Context) error {
		update := bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "pix", Value: document.Pix},
				{Key: "updatedAt", Value: document.UpdatedAt.UTC()},
			}},
			{Key: "$setOnInsert", Value: bson.D{{Key: "createdAt", Value: document.UpdatedAt.UTC()}}},
		}
		opts := options.UpdateOne().SetUpsert(true)
		if _, err := r.storeSettings.UpdateOne(writeCtx, bson.D{{Key: "_id", Value: domain.StoreSettingsID}}, update, opts); err != nil {
			return fmt.Errorf("update pix settings: %w", err)
		}
		return nil
	}

	return r.withStoreSettingsTransaction(ctx, write, audit)
}

func (r *CatalogRepository) withStoreSettingsTransaction(ctx context.Context, write func(context.Context) error, audit domain.AuditLog) error {
	session, err := r.storeSettings.Database().Client().StartSession()
	if err != nil {
		return fmt.Errorf("start store settings session: %w", err)
	}
	defer session.EndSession(ctx)

	transactionStarted := false
	writesStarted := false
	transactionErr := session.StartTransaction()
	if transactionErr == nil {
		transactionStarted = true
		transactionErr = mongo.WithSession(ctx, session, func(sessionCtx context.Context) error {
			if err := write(sessionCtx); err != nil {
				return err
			}
			writesStarted = true
			if err := r.insertAuditLog(sessionCtx, audit); err != nil {
				return err
			}
			return session.CommitTransaction(sessionCtx)
		})
	}
	if transactionErr != nil {
		if transactionStarted {
			_ = session.AbortTransaction(ctx)
		}
		if writesStarted || !isTransactionUnsupported(transactionErr) {
			return fmt.Errorf("store settings mutation in transaction: %w", transactionErr)
		}
		if err := write(ctx); err != nil {
			return err
		}
		return r.insertAuditLog(ctx, audit)
	}
	return nil
}

func newStoreSettingsDocument(settings domain.StoreSettings) storeSettingsDocument {
	document := storeSettingsDocument{
		ID:           domain.StoreSettingsID,
		Name:         settings.Name,
		IsOpen:       settings.IsOpen,
		OpeningHours: make([]openingHoursDocument, 0, len(settings.OpeningHours)),
		DeliveryTime: settings.DeliveryTime,
		WhatsApp:     settings.WhatsApp,
		Instagram:    settings.Instagram,
		Address:      settings.Address,
		Webhook: webhookSettingsDocument{
			URL:       settings.Webhook.URL,
			Active:    settings.Webhook.Active,
			Events:    copyWebhookEvents(settings.Webhook.Events),
			CreatedAt: settings.Webhook.CreatedAt.UTC(),
			UpdatedAt: settings.Webhook.UpdatedAt.UTC(),
		},
		Pix: pixSettingsDocument{
			Active:       settings.Pix.Active,
			Key:          settings.Pix.Key,
			MerchantName: settings.Pix.MerchantName,
			MerchantCity: settings.Pix.MerchantCity,
			CreatedAt:    settings.Pix.CreatedAt.UTC(),
			UpdatedAt:    settings.Pix.UpdatedAt.UTC(),
		},
		CreatedAt: settings.CreatedAt.UTC(),
		UpdatedAt: settings.UpdatedAt.UTC(),
	}
	if document.UpdatedAt.IsZero() {
		document.UpdatedAt = time.Now().UTC()
	}

	for _, day := range settings.OpeningHours {
		hours := openingHoursDocument{
			Day:       day.Day,
			Closed:    day.Closed,
			Intervals: make([]openingIntervalDocument, 0, len(day.Intervals)),
		}
		for _, interval := range day.Intervals {
			hours.Intervals = append(hours.Intervals, openingIntervalDocument{
				Open:  interval.Open,
				Close: interval.Close,
			})
		}
		document.OpeningHours = append(document.OpeningHours, hours)
	}
	return document
}

// copyWebhookEvents preserva a distinção entre seleção vazia (não enviar nada)
// e configuração legada (nil), que equivale a todos os eventos.
func copyWebhookEvents(events []string) []string {
	if events == nil {
		return nil
	}
	copied := make([]string, len(events))
	copy(copied, events)
	return copied
}
