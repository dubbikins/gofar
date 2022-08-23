package models

type Category struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Order       int      `json:"list_order"`
	Images      []string `json:"images"`
}
