package models

type Product struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Color       string  `json:"color"`
	Weight      string  `json:"weight"`
	Category    string  `json:"category"`
	Cost        float64 `json:"cost"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
}
