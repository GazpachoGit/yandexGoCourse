package storage

type IStorage interface {
	Set(original_url string, user string) (int, error)
	Get(int) (string, error)
	PingDB() error
	GetUserURLs(string) ([]URLInfo, error)
}
