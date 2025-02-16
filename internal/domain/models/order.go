package models

import "time"

// Order представляет заказ, созданный при покупке мерча
type Order struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	MerchID    int64     `json:"merch_id"`
	MerchName  string    `json:"merch_name"` // Имя товара; заполняется через JOIN с таблицей merch
	Quantity   int       `json:"quantity"`
	TotalPrice int       `json:"total_price"`
	CreatedAt  time.Time `json:"created_at"`
}
