package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// CollectionStoreSettings é a coleção do documento único de configuração da loja.
const CollectionStoreSettings = "store_settings"

type webhookSettingsDocument struct {
	URL    string `bson:"url"`
	Active bool   `bson:"active"`
	// Sem omitempty: nil (configuração legada) e lista vazia (nenhum evento)
	// precisam ser gravados de forma distinta.
	Events    []string  `bson:"events"`
	CreatedAt time.Time `bson:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt"`
}

type pixSettingsDocument struct {
	Active       bool      `bson:"active"`
	Key          string    `bson:"key"`
	MerchantName string    `bson:"merchantName"`
	MerchantCity string    `bson:"merchantCity"`
	CreatedAt    time.Time `bson:"createdAt"`
	UpdatedAt    time.Time `bson:"updatedAt"`
}

type storeSettingsDocument struct {
	ID           string                  `bson:"_id"`
	Name         string                  `bson:"name"`
	IsOpen       bool                    `bson:"isOpen"`
	OpeningHours []openingHoursDocument  `bson:"openingHours"`
	DeliveryTime string                  `bson:"deliveryTime"`
	WhatsApp     string                  `bson:"whatsapp"`
	Instagram    string                  `bson:"instagram"`
	Address      string                  `bson:"address"`
	Webhook      webhookSettingsDocument `bson:"webhook,omitempty"`
	Pix          pixSettingsDocument     `bson:"pix,omitempty"`
	CreatedAt    time.Time               `bson:"createdAt"`
	UpdatedAt    time.Time               `bson:"updatedAt"`
}

type openingHoursDocument struct {
	Day       string                    `bson:"day"`
	Closed    bool                      `bson:"closed"`
	Intervals []openingIntervalDocument `bson:"intervals"`
}

type openingIntervalDocument struct {
	Open  string `bson:"open"`
	Close string `bson:"close"`
}

func (d storeSettingsDocument) toDomain() domain.StoreSettings {
	settings := domain.StoreSettings{
		Name:         d.Name,
		IsOpen:       d.IsOpen,
		OpeningHours: make([]domain.OpeningHours, 0, len(d.OpeningHours)),
		DeliveryTime: d.DeliveryTime,
		WhatsApp:     d.WhatsApp,
		Instagram:    d.Instagram,
		Address:      d.Address,
		Webhook: domain.WebhookSettings{
			URL:       d.Webhook.URL,
			Active:    d.Webhook.Active,
			Events:    d.Webhook.Events,
			CreatedAt: d.Webhook.CreatedAt,
			UpdatedAt: d.Webhook.UpdatedAt,
		},
		Pix: domain.PixSettings{
			Active:       d.Pix.Active,
			Key:          d.Pix.Key,
			MerchantName: d.Pix.MerchantName,
			MerchantCity: d.Pix.MerchantCity,
			CreatedAt:    d.Pix.CreatedAt,
			UpdatedAt:    d.Pix.UpdatedAt,
		},
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}

	for _, day := range d.OpeningHours {
		hours := domain.OpeningHours{
			Day:       day.Day,
			Closed:    day.Closed,
			Intervals: make([]domain.OpeningInterval, 0, len(day.Intervals)),
		}
		for _, interval := range day.Intervals {
			hours.Intervals = append(hours.Intervals, domain.OpeningInterval{
				Open:  interval.Open,
				Close: interval.Close,
			})
		}
		settings.OpeningHours = append(settings.OpeningHours, hours)
	}
	return settings
}

// GetStoreSettings lê o documento único de configuração da loja.
func (r *CatalogRepository) GetStoreSettings(ctx context.Context) (domain.StoreSettings, error) {
	var document storeSettingsDocument
	if err := r.storeSettings.FindOne(ctx, bson.D{{Key: "_id", Value: domain.StoreSettingsID}}).Decode(&document); err != nil {
		// Configuração ausente é erro de operação do servidor, não "recurso
		// não encontrado" para o cliente da cotação.
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.StoreSettings{}, fmt.Errorf("store settings document is missing: %w", err)
		}
		return domain.StoreSettings{}, fmt.Errorf("find store settings: %w", err)
	}
	return document.toDomain(), nil
}
