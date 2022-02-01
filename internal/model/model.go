package model

type HandlerURLInfo struct {
	CorrelationID string `json:"correlation_id,omitempty"`
	OriginalURL   string `json:"original_url,omitempty"`
	ShortURL      string `json:"short_url,omitempty"`
}

type StorageURLInfo struct {
	ID          int    `db:"id"`
	OriginalURL string `db:"original_url"`
	UserID      string `db:"user_id"`
}

type StorageInsertInfo struct {
	ID   int  `db:"id"`
	Conf bool `db:"conf"`
}

type StorageTables struct {
	URLTable string
}
