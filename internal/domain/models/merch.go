package models

// Merch представляет товар мерча, доступный для покупки
type Merch struct {
	ID    int64  // Уникальный идентификатор товара
	Name  string // Название товара (уникальное)
	Price int    // Цена товара в монетах
}
