package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// CollectionOrders e CollectionCounters seguem docs/05-modelo-de-dados.md.
const (
	CollectionOrders   = "orders"
	CollectionCounters = "counters"

	orderNumberCounterID = "order_number"
)

type orderDocument struct {
	ID               bson.ObjectID          `bson:"_id,omitempty"`
	Number           int64                  `bson:"number"`
	PublicID         string                 `bson:"publicId"`
	Status           string                 `bson:"status"`
	Fulfillment      string                 `bson:"fulfillment"`
	Customer         customerDocument       `bson:"customer"`
	Address          *addressDocument       `bson:"address,omitempty"`
	Zone             *zoneSnapshotDocument  `bson:"zone,omitempty"`
	Items            []orderItemDocument    `bson:"items"`
	SubtotalCents    int64                  `bson:"subtotalCents"`
	DeliveryFeeCents int64                  `bson:"deliveryFeeCents"`
	TotalCents       int64                  `bson:"totalCents"`
	Payment          paymentDocument        `bson:"payment"`
	Note             string                 `bson:"note,omitempty"`
	IdempotencyKey   string                 `bson:"idempotencyKey"`
	RequestHash      string                 `bson:"requestHash"`
	AccessTokenHash  string                 `bson:"accessTokenHash,omitempty"`
	StatusHistory    []statusChangeDocument `bson:"statusHistory"`
	ConfirmedAt      *time.Time             `bson:"confirmedAt,omitempty"`
	CancelledAt      *time.Time             `bson:"cancelledAt,omitempty"`
	CancelReason     string                 `bson:"cancelReason,omitempty"`
	CreatedAt        time.Time              `bson:"createdAt"`
	UpdatedAt        time.Time              `bson:"updatedAt"`
}

type customerDocument struct {
	Name  string `bson:"name"`
	Phone string `bson:"phone"`
}

type addressDocument struct {
	Street     string `bson:"street"`
	Number     string `bson:"number"`
	Complement string `bson:"complement,omitempty"`
	District   string `bson:"district"`
	City       string `bson:"city"`
	State      string `bson:"state"`
	ZipCode    string `bson:"zipCode"`
	Reference  string `bson:"reference,omitempty"`
}

type zoneSnapshotDocument struct {
	ID       bson.ObjectID `bson:"id"`
	Name     string        `bson:"name"`
	FeeCents int64         `bson:"feeCents"`
}

type orderItemDocument struct {
	ProductID      bson.ObjectID        `bson:"productId"`
	Name           string               `bson:"name"`
	UnitPriceCents int64                `bson:"unitPriceCents"`
	Quantity       int                  `bson:"quantity"`
	Options        []itemOptionDocument `bson:"options"`
	Note           string               `bson:"note,omitempty"`
	LineTotalCents int64                `bson:"lineTotalCents"`
}

type itemOptionDocument struct {
	GroupID    string `bson:"groupId,omitempty"`
	GroupName  string `bson:"groupName,omitempty"`
	OptionID   string `bson:"optionId"`
	Name       string `bson:"name"`
	Kind       string `bson:"kind"`
	PriceCents int64  `bson:"priceCents"`
}

type paymentDocument struct {
	Method         string `bson:"method"`
	ChangeForCents *int64 `bson:"changeForCents,omitempty"`
}

type statusChangeDocument struct {
	Status  string    `bson:"status"`
	At      time.Time `bson:"at"`
	ActorID string    `bson:"actorId,omitempty"`
}

type orderCounterDocument struct {
	ID    string `bson:"_id"`
	Value int64  `bson:"value"`
}

// OrderRepository grava e lê pedidos. As regras de negócio ficam no domínio e
// no service; aqui há apenas tradução BSON e garantia de unicidade.
type OrderRepository struct {
	*CatalogRepository
	database *mongo.Database
	orders   *mongo.Collection
	counters *mongo.Collection
}

func NewOrderRepository(database *mongo.Database) (*OrderRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	catalog, err := NewCatalogRepository(database)
	if err != nil {
		return nil, err
	}
	return &OrderRepository{
		CatalogRepository: catalog,
		database:          database,
		orders:            database.Collection(CollectionOrders),
		counters:          database.Collection(CollectionCounters),
	}, nil
}

// EnsureOrderIndexes cria os índices de pedidos de forma idempotente.
func EnsureOrderIndexes(ctx context.Context, database *mongo.Database) error {
	if database == nil {
		return errors.New("mongo database is required")
	}

	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "number", Value: 1}}, Options: options.Index().SetName("orders_number_unique").SetUnique(true)},
		{Keys: bson.D{{Key: "publicId", Value: 1}}, Options: options.Index().SetName("orders_public_id_unique").SetUnique(true)},
		{Keys: bson.D{{Key: "idempotencyKey", Value: 1}}, Options: options.Index().SetName("orders_idempotency_unique").SetUnique(true).SetSparse(true)},
		{Keys: bson.D{{Key: "status", Value: 1}, {Key: "createdAt", Value: -1}}, Options: options.Index().SetName("orders_status_created")},
		{Keys: bson.D{{Key: "createdAt", Value: -1}}, Options: options.Index().SetName("orders_created")},
		{Keys: bson.D{{Key: "customer.phone", Value: 1}, {Key: "createdAt", Value: -1}}, Options: options.Index().SetName("orders_phone_created")},
		{Keys: bson.D{{Key: "accessTokenHash", Value: 1}, {Key: "createdAt", Value: -1}}, Options: options.Index().SetName("orders_access_token_created").SetSparse(true)},
	}
	if _, err := database.Collection(CollectionOrders).Indexes().CreateMany(ctx, indexes); err != nil {
		return fmt.Errorf("ensure indexes on %s: %w", CollectionOrders, err)
	}
	return nil
}

// FindByIdempotencyKey localiza um pedido já gravado com a mesma chave.
func (r *OrderRepository) FindByIdempotencyKey(ctx context.Context, key string) (domain.Order, bool, error) {
	var document orderDocument
	err := r.orders.FindOne(ctx, bson.D{{Key: "idempotencyKey", Value: key}}).Decode(&document)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return domain.Order{}, false, nil
	}
	if err != nil {
		return domain.Order{}, false, fmt.Errorf("find order by idempotency key: %w", err)
	}
	return document.toDomain(), true, nil
}

// FindByPublicID localiza um pedido pelo identificador opaco.
func (r *OrderRepository) FindByPublicID(ctx context.Context, publicID string) (domain.Order, error) {
	var document orderDocument
	if err := r.orders.FindOne(ctx, bson.D{{Key: "publicId", Value: publicID}}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Order{}, domain.ErrNotFound
		}
		return domain.Order{}, fmt.Errorf("find order by public id: %w", err)
	}
	return document.toDomain(), nil
}

// ListByAccessToken devolve uma página do histórico associado a um token. O
// filtro usa apenas o hash: o token em claro nunca chega ao banco.
func (r *OrderRepository) ListByAccessToken(ctx context.Context, accessTokenHash string, query domain.OrderHistoryQuery) (orders []domain.Order, total int64, err error) {
	if accessTokenHash == "" {
		return nil, 0, errors.New("access token hash is required")
	}
	page := query.Page
	if page < 1 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize < 1 {
		pageSize = 20
	}

	filter := bson.D{{Key: "accessTokenHash", Value: accessTokenHash}}
	total, err = r.orders.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count orders by access token: %w", err)
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}, {Key: "_id", Value: -1}}).
		SetSkip(int64(page-1) * int64(pageSize)).
		SetLimit(int64(pageSize))
	cursor, err := r.orders.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("list orders by access token: %w", err)
	}
	defer func() {
		err = errors.Join(err, cursor.Close(ctx))
	}()

	orders = make([]domain.Order, 0)
	for cursor.Next(ctx) {
		var document orderDocument
		if err := cursor.Decode(&document); err != nil {
			return nil, 0, fmt.Errorf("decode order history: %w", err)
		}
		orders = append(orders, document.toDomain())
	}
	if err := cursor.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate order history: %w", err)
	}
	return orders, total, nil
}

// SetAccessTokenHash associa um pedido já gravado ao token apresentado pelo
// navegador. É usado em retries idempotentes e em pedidos anteriores ao token.
func (r *OrderRepository) SetAccessTokenHash(ctx context.Context, id, accessTokenHash string) error {
	if accessTokenHash == "" {
		return errors.New("access token hash is required")
	}
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid order id %q: %w", id, err)
	}
	result, err := r.orders.UpdateOne(ctx,
		bson.D{{Key: "_id", Value: objectID}},
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "accessTokenHash", Value: accessTokenHash},
			{Key: "updatedAt", Value: time.Now().UTC()},
		}}},
	)
	if err != nil {
		return fmt.Errorf("set order access token hash: %w", err)
	}
	if result.MatchedCount == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// Create grava o pedido e seu número sequencial. Emimplantações com suporte a
// transação, ambas as gravações são atômicas; em Mongo standalone (desenvolvimento
// e E2E), a gravação usa o mesmo código sem transação.
func (r *OrderRepository) Create(ctx context.Context, order domain.Order) (domain.Order, error) {
	client := r.database.Client()
	session, err := client.StartSession()
	if err != nil {
		return domain.Order{}, fmt.Errorf("start order session: %w", err)
	}
	defer session.EndSession(ctx)

	var stored orderDocument
	transactionStarted := false
	writesStarted := false
	transactionErr := session.StartTransaction()
	if transactionErr == nil {
		transactionStarted = true
		transactionErr = mongo.WithSession(ctx, session, func(sessionCtx context.Context) error {
			document, err := r.insert(sessionCtx, order)
			writesStarted = err == nil
			if err != nil {
				return err
			}
			stored = document
			return session.CommitTransaction(sessionCtx)
		})
	}

	if transactionErr != nil {
		if transactionStarted {
			_ = session.AbortTransaction(ctx)
		}
		if writesStarted || !isTransactionUnsupported(transactionErr) {
			return domain.Order{}, fmt.Errorf("create order in transaction: %w", transactionErr)
		}
		stored, err = r.insert(ctx, order)
		if err != nil {
			return domain.Order{}, err
		}
	}

	return stored.toDomain(), nil
}

func (r *OrderRepository) insert(ctx context.Context, order domain.Order) (orderDocument, error) {
	number, err := r.nextOrderNumber(ctx)
	if err != nil {
		return orderDocument{}, err
	}

	document, err := newOrderDocument(order, number)
	if err != nil {
		return orderDocument{}, err
	}
	if _, err := r.orders.InsertOne(ctx, document); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return orderDocument{}, domain.ErrConflict
		}
		return orderDocument{}, fmt.Errorf("insert order: %w", err)
	}
	return document, nil
}

func (r *OrderRepository) nextOrderNumber(ctx context.Context) (int64, error) {
	filter := bson.D{{Key: "_id", Value: orderNumberCounterID}}
	update := bson.D{{Key: "$inc", Value: bson.D{{Key: "value", Value: int64(1)}}}}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var counter orderCounterDocument
	if err := r.counters.FindOneAndUpdate(ctx, filter, update, opts).Decode(&counter); err != nil {
		return 0, fmt.Errorf("next order number: %w", err)
	}
	if counter.Value < 1 {
		return 0, errors.New("invalid order number")
	}
	return counter.Value, nil
}

func isTransactionUnsupported(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "transaction numbers are only allowed") ||
		strings.Contains(message, "transactions are not supported")
}

func newOrderDocument(order domain.Order, number int64) (orderDocument, error) {
	objectID, err := bson.ObjectIDFromHex(order.ID)
	if err != nil {
		return orderDocument{}, fmt.Errorf("invalid order id %q: %w", order.ID, err)
	}

	document := orderDocument{
		ID:               objectID,
		Number:           number,
		PublicID:         order.PublicID,
		Status:           order.Status,
		Fulfillment:      string(order.Fulfillment),
		Customer:         customerDocument{Name: order.Customer.Name, Phone: order.Customer.Phone},
		Items:            make([]orderItemDocument, 0, len(order.Items)),
		SubtotalCents:    order.SubtotalCents,
		DeliveryFeeCents: order.DeliveryFeeCents,
		TotalCents:       order.TotalCents,
		Payment: paymentDocument{
			Method:         order.Payment.Method,
			ChangeForCents: order.Payment.ChangeForCents,
		},
		Note:            order.Note,
		IdempotencyKey:  order.IdempotencyKey,
		RequestHash:     order.RequestHash,
		AccessTokenHash: order.AccessTokenHash,
		StatusHistory:   make([]statusChangeDocument, 0, len(order.StatusHistory)),
		CreatedAt:       order.CreatedAt.UTC(),
		UpdatedAt:       order.UpdatedAt.UTC(),
	}

	if order.Address != nil {
		document.Address = &addressDocument{
			Street:     order.Address.Street,
			Number:     order.Address.Number,
			Complement: order.Address.Complement,
			District:   order.Address.District,
			City:       order.Address.City,
			State:      order.Address.State,
			ZipCode:    order.Address.ZipCode,
			Reference:  order.Address.Reference,
		}
	}
	if order.Zone != nil {
		zoneID, err := bson.ObjectIDFromHex(order.Zone.ID)
		if err != nil {
			return orderDocument{}, fmt.Errorf("invalid zone id %q: %w", order.Zone.ID, err)
		}
		document.Zone = &zoneSnapshotDocument{ID: zoneID, Name: order.Zone.Name, FeeCents: order.Zone.FeeCents}
	}

	for _, item := range order.Items {
		productID, err := bson.ObjectIDFromHex(item.ProductID)
		if err != nil {
			return orderDocument{}, fmt.Errorf("invalid product id %q: %w", item.ProductID, err)
		}
		options := make([]itemOptionDocument, 0, len(item.Options))
		for _, option := range item.Options {
			options = append(options, itemOptionDocument{
				GroupID:    option.GroupID,
				GroupName:  option.GroupName,
				OptionID:   option.OptionID,
				Name:       option.Name,
				Kind:       option.Kind,
				PriceCents: option.PriceCents,
			})
		}
		document.Items = append(document.Items, orderItemDocument{
			ProductID:      productID,
			Name:           item.Name,
			UnitPriceCents: item.UnitPriceCents,
			Quantity:       item.Quantity,
			Options:        options,
			Note:           item.Note,
			LineTotalCents: item.LineTotalCents,
		})
	}

	for _, change := range order.StatusHistory {
		document.StatusHistory = append(document.StatusHistory, statusChangeDocument{
			Status:  change.Status,
			At:      change.At.UTC(),
			ActorID: change.ActorID,
		})
	}
	if order.ConfirmedAt != nil {
		confirmedAt := order.ConfirmedAt.UTC()
		document.ConfirmedAt = &confirmedAt
	}
	if order.CancelledAt != nil {
		cancelledAt := order.CancelledAt.UTC()
		document.CancelledAt = &cancelledAt
	}
	document.CancelReason = order.CancelReason

	return document, nil
}

func (d orderDocument) toDomain() domain.Order {
	order := domain.Order{
		ID:               d.ID.Hex(),
		Number:           d.Number,
		PublicID:         d.PublicID,
		Status:           d.Status,
		Fulfillment:      domain.Fulfillment(d.Fulfillment),
		Customer:         domain.OrderCustomer{Name: d.Customer.Name, Phone: d.Customer.Phone},
		Items:            make([]domain.OrderItem, 0, len(d.Items)),
		SubtotalCents:    d.SubtotalCents,
		DeliveryFeeCents: d.DeliveryFeeCents,
		TotalCents:       d.TotalCents,
		Payment: domain.OrderPayment{
			Method:         d.Payment.Method,
			ChangeForCents: d.Payment.ChangeForCents,
		},
		Note:            d.Note,
		IdempotencyKey:  d.IdempotencyKey,
		RequestHash:     d.RequestHash,
		AccessTokenHash: d.AccessTokenHash,
		StatusHistory:   make([]domain.OrderStatusChange, 0, len(d.StatusHistory)),
		CreatedAt:       d.CreatedAt,
		UpdatedAt:       d.UpdatedAt,
	}

	if d.Address != nil {
		order.Address = &domain.OrderAddress{
			Street:     d.Address.Street,
			Number:     d.Address.Number,
			Complement: d.Address.Complement,
			District:   d.Address.District,
			City:       d.Address.City,
			State:      d.Address.State,
			ZipCode:    d.Address.ZipCode,
			Reference:  d.Address.Reference,
		}
	}
	if d.Zone != nil {
		order.Zone = &domain.OrderZone{ID: d.Zone.ID.Hex(), Name: d.Zone.Name, FeeCents: d.Zone.FeeCents}
	}

	for _, item := range d.Items {
		options := make([]domain.OrderItemOption, 0, len(item.Options))
		for _, option := range item.Options {
			options = append(options, domain.OrderItemOption{
				GroupID:    option.GroupID,
				GroupName:  option.GroupName,
				OptionID:   option.OptionID,
				Name:       option.Name,
				Kind:       option.Kind,
				PriceCents: option.PriceCents,
			})
		}
		order.Items = append(order.Items, domain.OrderItem{
			ProductID:      item.ProductID.Hex(),
			Name:           item.Name,
			UnitPriceCents: item.UnitPriceCents,
			Quantity:       item.Quantity,
			Options:        options,
			Note:           item.Note,
			LineTotalCents: item.LineTotalCents,
		})
	}

	for _, change := range d.StatusHistory {
		order.StatusHistory = append(order.StatusHistory, domain.OrderStatusChange{
			Status:  change.Status,
			At:      change.At,
			ActorID: change.ActorID,
		})
	}
	if d.ConfirmedAt != nil {
		confirmedAt := d.ConfirmedAt.UTC()
		order.ConfirmedAt = &confirmedAt
	}
	if d.CancelledAt != nil {
		cancelledAt := d.CancelledAt.UTC()
		order.CancelledAt = &cancelledAt
	}
	order.CancelReason = d.CancelReason
	return order
}
