package storage

import "github.com/GazpachoGit/yandexGoCourse/internal/model"

type IStorage interface {
	Set(original_url string, user string) (int, error)
	Get(int) (string, error)
	PingDB() error
	GetUserURLs(string) ([]model.StorageURLInfo, error)
	SetBatchURLs(input *[]*model.HandlerURLInfo, username string) (*map[string]*model.StorageURLInfo, error)
}
