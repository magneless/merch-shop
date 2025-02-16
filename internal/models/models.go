package models

type InventoryItem struct {
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
}

type CoinHistory struct {
	Received []CoinTransaction `json:"received"`
	Sent     []CoinTransaction `json:"sent"`
}

type CoinTransaction struct {
	FromUser string `json:"fromUser"`
	ToUser   string `json:"toUser"`
	Amount   int    `json:"amount"`
}