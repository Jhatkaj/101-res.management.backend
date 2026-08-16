package model

import "time"

type MenuItem struct { ID int64 `json:"id"`; Name string `json:"name"`; Description string `json:"description"`; Price float64 `json:"price"`; Emoji string `json:"emoji"` }
type Special struct { ID int64 `json:"id"`; Name string `json:"name"`; Description string `json:"description"`; Price float64 `json:"price"`; Emoji string `json:"emoji"` }
type StockItem struct { ID int64 `json:"id"`; Name string `json:"name"`; Quantity float64 `json:"quantity"`; Unit string `json:"unit"`; Icon string `json:"icon"` }
type OrderItem struct { ID int64 `json:"id,omitempty"`; Name string `json:"name"`; Price float64 `json:"price"`; Qty int `json:"qty"`; Emoji string `json:"emoji,omitempty"` }
type Order struct { ID string `json:"id"`; Customer string `json:"customer"`; Date string `json:"date"`; CreatedAt time.Time `json:"createdAt"`; Items []OrderItem `json:"items"`; Total float64 `json:"total"`; Status string `json:"status"` }
type Bootstrap struct { Menu []MenuItem `json:"menu"`; Specials []Special `json:"specials"`; Orders []Order `json:"orders"`; Stock []StockItem `json:"stock"` }
