package model

type HandlerURLInfo struct {
	Correlation_id string `json:"correlation_id,omitempty"`
	Original_url   string `json:"original_url,omitempty"`
	Short_url      string `json:"short_url,omitempty"`
}

type StorageURLInfo struct {
	Id           int    `db:"id"`
	Original_url string `db:"original_url"`
	User_id      string `db:"user_id"`
}
