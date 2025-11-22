package storage

import (
	"context"
	"fmt"

	"order-service-wbtech/internal/config"
	"order-service-wbtech/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	Pool *pgxpool.Pool
}

func NewPostgres(cfg *config.Config) (*Postgres, error) {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.DBcfg.User,
		cfg.DBcfg.Password,
		cfg.DBcfg.Host,
		cfg.DBcfg.Port,
		cfg.DBcfg.Name,
	)

	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, err
	}

	return &Postgres{Pool: pool}, nil
}

func (p *Postgres) SaveOrder(ctx context.Context, order *model.Order) error {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO orders (
			order_uid, track_number, entry, locale, internal_signature,
			customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale,
		order.InternalSignature, order.CustomerID, order.DeliveryService,
		order.ShardKey, order.SMID, order.DateCreated, order.OOFShard,
	)
	if err != nil {
		return fmt.Errorf("insert orders: %w", err)
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO delivery (
			order_uid, name, phone, zip, city, address, region, email
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		order.OrderUID, order.Delivery.Name, order.Delivery.Phone,
		order.Delivery.Zip, order.Delivery.City, order.Delivery.Address,
		order.Delivery.Region, order.Delivery.Email,
	)
	if err != nil {
		return fmt.Errorf("insert delivery: %w", err)
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO payment (
			order_uid, transaction, request_id, currency, provider,
			amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		order.OrderUID, order.Payment.Transaction, order.Payment.RequestID,
		order.Payment.Currency, order.Payment.Provider, order.Payment.Amount,
		order.Payment.PaymentDt, order.Payment.Bank, order.Payment.DeliveryCost,
		order.Payment.GoodsTotal, order.Payment.CustomFee,
	)
	if err != nil {
		return fmt.Errorf("insert payment: %w", err)
	}

	for _, item := range order.Items {
		_, err := tx.Exec(ctx,
			`INSERT INTO items (
                order_uid, chrt_id, track_number, price, rid, name, sale, size,
                total_price, nm_id, brand, status
            ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
			order.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.Rid, item.Name,
			item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status,
		)
		if err != nil {
			return fmt.Errorf("insert item: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (p *Postgres) GetOrder(ctx context.Context, orderUID string) (*model.Order, error) {
	order := &model.Order{}

	err := p.Pool.QueryRow(ctx,
		`SELECT order_uid, track_number, entry, locale, internal_signature,
		        customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
		  FROM orders WHERE order_uid=$1`, orderUID).Scan(
		&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale,
		&order.InternalSignature, &order.CustomerID, &order.DeliveryService,
		&order.ShardKey, &order.SMID, &order.DateCreated, &order.OOFShard,
	)
	if err != nil {
		return nil, err
	}

	err = p.Pool.QueryRow(ctx,
		`SELECT name, phone, zip, city, address, region, email FROM delivery WHERE order_uid=$1`,
		orderUID).Scan(
		&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip,
		&order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region,
		&order.Delivery.Email,
	)
	if err != nil {
		return nil, err
	}

	err = p.Pool.QueryRow(ctx,
		`SELECT transaction, request_id, currency, provider, amount, payment_dt, bank,
		        delivery_cost, goods_total, custom_fee
		  FROM payment WHERE order_uid=$1`,
		orderUID).Scan(
		&order.Payment.Transaction, &order.Payment.RequestID, &order.Payment.Currency,
		&order.Payment.Provider, &order.Payment.Amount, &order.Payment.PaymentDt,
		&order.Payment.Bank, &order.Payment.DeliveryCost, &order.Payment.GoodsTotal,
		&order.Payment.CustomFee,
	)
	if err != nil {
		return nil, err
	}

	rows, err := p.Pool.Query(ctx,
		`SELECT chrt_id, price, rid, name, sale, size, total_price, nm_id, brand, status
		  FROM items WHERE order_uid=$1`, orderUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	order.Items = []model.Item{}
	for rows.Next() {
		var item model.Item
		if err := rows.Scan(&item.ChrtID, &item.Price, &item.Rid, &item.Name,
			&item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status); err != nil {
			return nil, err
		}
		order.Items = append(order.Items, item)
	}

	return order, nil
}

func (p *Postgres) LoadOrders(ctx context.Context) ([]*model.Order, error) {
	query := `
		SELECT order_uid
		FROM orders
		WHERE date_created >= NOW() - INTERVAL '3 hours'
	`
	rows, err := p.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]*model.Order, 0)

	for rows.Next() {
		var orderUID string
		if err := rows.Scan(&orderUID); err != nil {
			return nil, err
		}

		order, err := p.GetOrder(ctx, orderUID)
		if err != nil {
			fmt.Printf("Failed to load order %s: %v\n", orderUID, err)
			continue
		}

		orders = append(orders, order)
	}
	return orders, err
}
