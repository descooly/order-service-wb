package database

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/descooly/order-service-wb/internal/model"

	_ "github.com/lib/pq"
)

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func ConnectDB() (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "myuser"),
		getEnv("DB_PASSWORD", "myuserpass"),
		getEnv("DB_NAME", "myappdb"),
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open DB connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping DB: %w", err)
	}

	return db, nil
}

func InsertOrder(db *sql.DB, order *model.OrderStruct) error {
	trnsact, err := db.Begin()
	if err != nil {
		return err
	}
	defer trnsact.Rollback()

	var OrderID int
	err = trnsact.QueryRow(`INSERT INTO order_info (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) 
	ON CONFLICT (order_uid) DO NOTHING 
	RETURNING ID`, order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature, order.CustomerId, order.DeliveryService, order.Shardkey, order.SmId, order.DateCreated, order.OofShard).Scan(&OrderID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("failed to insert into order_info: %w", err)
	}

	_, err = trnsact.Exec(`INSERT INTO delivery (order_id, d_name, phone, zip, city, address, region, email
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		OrderID,
		order.Delivery.Name,
		order.Delivery.Phone,
		order.Delivery.Zip,
		order.Delivery.City,
		order.Delivery.Address,
		order.Delivery.Region,
		order.Delivery.Email)
	if err != nil {
		return fmt.Errorf("failed to insert into delivery: %w", err)
	}

	_, err = trnsact.Exec(`
		INSERT INTO payment (order_id, p_transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		OrderID,
		order.Payment.Transaction,
		order.Payment.RequestId,
		order.Payment.Currency,
		order.Payment.Provider,
		order.Payment.Amount,
		order.Payment.PaymentDt,
		order.Payment.Bank,
		order.Payment.DeliveryCost,
		order.Payment.GoodsTotal,
		order.Payment.CustomFee)
	if err != nil {
		return fmt.Errorf("failed to insert into payment: %w", err)
	}

	for _, item := range order.Items {
		_, err = trnsact.Exec(`
			INSERT INTO items (order_id, chrt_id, track_number, price, rid, i_name, sale, i_size, total_price, nm_id, brand, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
			OrderID,
			item.ChrtId,
			item.TrackNumber,
			item.Price,
			item.Rid,
			item.Name,
			item.Sale,
			item.Size,
			item.TotalPrice,
			item.NmId,
			item.Brand,
			item.Status)
		if err != nil {
			return fmt.Errorf("failed to insert into []Items: %w", err)
		}
	}

	return trnsact.Commit()
}

func LoadOrders(db *sql.DB) ([]model.OrderStruct, error) {
	var dummy int
	rows, err := db.Query(`SELECT id, order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard FROM order_info`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.OrderStruct
	orderMap := make(map[int]*model.OrderStruct)

	for rows.Next() {
		var order model.OrderStruct
		var id int
		err := rows.Scan(&id, &order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature, &order.CustomerId, &order.DeliveryService, &order.Shardkey, &order.SmId, &order.DateCreated, &order.OofShard)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
		orderMap[id] = &orders[len(orders)-1]
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	delRows, err := db.Query(`SELECT id, order_id, d_name, phone, zip, city, address, region, email FROM delivery`)
	if err != nil {
		return nil, err
	}
	defer delRows.Close()

	for delRows.Next() {
		var del model.Delivery
		var orderID int
		err := delRows.Scan(&dummy, &orderID, &del.Name, &del.Phone, &del.Zip, &del.City, &del.Address, &del.Region, &del.Email)
		if err != nil {
			return nil, err
		}
		if order, exists := orderMap[orderID]; exists {
			order.Delivery = del
		}
	}
	if err = delRows.Err(); err != nil {
		return nil, err
	}

	PayRows, err := db.Query(`SELECT id, order_id, p_transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee FROM payment`)
	if err != nil {
		return nil, err
	}
	defer PayRows.Close()

	for PayRows.Next() {
		var pay model.Payment
		var orderID int
		err := PayRows.Scan(&dummy, &orderID, &pay.Transaction, &pay.RequestId, &pay.Currency, &pay.Provider, &pay.Amount, &pay.PaymentDt, &pay.Bank, &pay.DeliveryCost, &pay.GoodsTotal, &pay.CustomFee)
		if err != nil {
			return nil, err
		}
		if order, exists := orderMap[orderID]; exists {
			order.Payment = pay
		}
	}
	if err = PayRows.Err(); err != nil {
		return nil, err
	}

	ItemRows, err := db.Query(`SELECT id, order_id, chrt_id, track_number, price, rid, i_name, sale, i_size, total_price, nm_id, brand, status FROM items`)
	if err != nil {
		return nil, err
	}
	defer ItemRows.Close()

	for ItemRows.Next() {
		var it model.Item
		var orderID int
		err := ItemRows.Scan(&dummy, &orderID, &it.ChrtId, &it.TrackNumber, &it.Price, &it.Rid, &it.Name, &it.Sale, &it.Size, &it.TotalPrice, &it.NmId, &it.Brand, &it.Status)
		if err != nil {
			return nil, err
		}
		if order, exists := orderMap[orderID]; exists {
			order.Items = append(order.Items, it)
		}
	}
	if err = ItemRows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}
