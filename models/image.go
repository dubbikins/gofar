package models

type Image struct {
	Location string `json:"src"`
	Title    string `json:"title"`
	ID       string `json:"id"`
	Type     string `json:"type"`
}
