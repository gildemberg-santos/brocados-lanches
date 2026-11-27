package repository

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/brocados-lanches/api/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var orderSortFields = []string{"createdAt", "number", "status", "totalCents"}

// AdminListOrders devolve uma página da fila. Filtros e ordenação são
// construídos aqui a partir de valores já validados pelo service; o cliente
// nunca escolhe um campo do documento para ordenar.
func (r *OrderRepository) AdminListOrders(ctx context.Context, query domain.AdminOrderListQuery) (orders []domain.Order, total int64, err error) {
	filter := bson.D{}
	if query.Status != "" {
		filter = append(filter, bson.E{Key: "status", Value: query.Status})
	}
	if query.Fulfillment != "" {
		filter = append(filter, bson.E{Key: "fulfillment", Value: string(query.Fulfillment)})
	}
	if query.From != nil {
		filter = append(filter, bson.E{Key: "createdAt", Value: bson.D{{Key: "$gte", Value: query.From.UTC()}}})
	}
	if query.To != nil {
		filter = append(filter, bson.E{Key: "createdAt", Value: bson.D{{Key: "$lte", Value: query.To.UTC()}}})
	}
	if digits := phoneDigits(query.Search); digits != "" {
		conditions := bson.A{}
		if number, err := strconv.ParseInt(digits, 10, 64); err == nil {
			conditions = append(conditions, bson.D{{Key: "number", Value: number}})
		}
		conditions = append(conditions, bson.D{{Key: "customer.phone", Value: bson.D{
			{Key: "$regex", Value: regexp.QuoteMeta(digits)},
		}}})
		filter = append(filter, bson.E{Key: "$or", Value: conditions})
	}

	total, err = r.orders.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count admin orders: %w", err)
	}

	page, pageSize := query.Page, query.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	sortQuery := domain.AdminListQuery{Page: page, PageSize: pageSize, Sort: query.Sort, Order: query.Order}
	if sortQuery.Sort == "" {
		sortQuery.Sort = "createdAt"
	}
	sortDocument, err := adminSortDocument(sortQuery, orderSortFields)
	if err != nil {
		return nil, 0, err
	}
	cursor, err := r.orders.Find(ctx, filter, options.Find().
		SetSort(sortDocument).
		SetSkip(int64((page-1)*pageSize)).
		SetLimit(int64(pageSize)))
	if err != nil {
		return nil, 0, fmt.Errorf("find admin orders: %w", err)
	}
	defer func() {
		err = errors.Join(err, cursor.Close(ctx))
	}()

	var documents []orderDocument
	if err := cursor.All(ctx, &documents); err != nil {
		return nil, 0, fmt.Errorf("decode admin orders: %w", err)
	}
	orders = convertDocuments(documents, orderDocument.toDomain)
	return orders, total, nil
}

// FindByID localiza um pedido pelo ID hexadecimal usado no painel.
func (r *OrderRepository) FindByID(ctx context.Context, id string) (domain.Order, error) {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return domain.Order{}, domain.ErrNotFound
	}
	var document orderDocument
	if err := r.orders.FindOne(ctx, bson.D{{Key: "_id", Value: objectID}}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Order{}, domain.ErrNotFound
		}
		return domain.Order{}, fmt.Errorf("find order by id: %w", err)
	}
	return document.toDomain(), nil
}

// AdminUpdateOrderStatus grava a transição somente se o pedido ainda estiver
// no status lido pelo service (compare-and-set). Um segundo operador tentando
// a mesma transição recebe ErrConflict em vez de sobrescrever a primeira.
func (r *OrderRepository) AdminUpdateOrderStatus(ctx context.Context, id, expectedStatus string, order domain.Order, audit domain.AuditLog) error {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return domain.ErrNotFound
	}
	document, err := newOrderDocument(order, order.Number)
	if err != nil {
		return err
	}

	write := func(writeCtx context.Context) error {
		result, err := r.orders.ReplaceOne(writeCtx, bson.D{
			{Key: "_id", Value: objectID},
			{Key: "status", Value: expectedStatus},
		}, document)
		if err != nil {
			return fmt.Errorf("replace order status: %w", err)
		}
		if result.MatchedCount > 0 {
			return nil
		}
		count, err := r.orders.CountDocuments(writeCtx, bson.D{{Key: "_id", Value: objectID}}, options.Count().SetLimit(1))
		if err != nil {
			return fmt.Errorf("count order after stale transition: %w", err)
		}
		if count == 0 {
			return domain.ErrNotFound
		}
		return domain.ErrConflict
	}

	session, err := r.database.Client().StartSession()
	if err != nil {
		return fmt.Errorf("start order status session: %w", err)
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
			return transactionErr
		}
		// Em standalone a transação falha antes de qualquer escrita; aplica
		// as duas operações em sequência, como no resto do repositório.
		if err := write(ctx); err != nil {
			return err
		}
		return r.insertAuditLog(ctx, audit)
	}
	return nil
}

// AdminDeleteOrder remove o pedido e registra a auditoria na mesma operação
// quando o MongoDB suporta transação. Em standalone, onde não há transação, as
// duas escritas acontecem em sequência.
func (r *OrderRepository) AdminDeleteOrder(ctx context.Context, id string, audit domain.AuditLog) error {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return domain.ErrNotFound
	}

	write := func(writeCtx context.Context) error {
		result, err := r.orders.DeleteOne(writeCtx, bson.D{{Key: "_id", Value: objectID}})
		if err != nil {
			return fmt.Errorf("delete order: %w", err)
		}
		if result.DeletedCount == 0 {
			return domain.ErrNotFound
		}
		return nil
	}

	session, err := r.database.Client().StartSession()
	if err != nil {
		return fmt.Errorf("start order delete session: %w", err)
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
			return transactionErr
		}
		if err := write(ctx); err != nil {
			return err
		}
		return r.insertAuditLog(ctx, audit)
	}
	return nil
}

// phoneDigits extrai apenas dígitos para que a busca funcione com o telefone
// formatado ou mascarado exibido na fila.
func phoneDigits(value string) string {
	var builder strings.Builder
	for _, character := range value {
		if character >= '0' && character <= '9' {
			builder.WriteRune(character)
		}
	}
	return builder.String()
}
