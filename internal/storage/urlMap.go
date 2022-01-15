package storage

import (
	"errors"
	"sync"
)

const (
	ErrNotFound = "can't find id"
)

type URLMap struct {
	data  *sync.Map
	count int
}

func NewUrlMap() GetSet {
	return &URLMap{data: &sync.Map{}, count: 0}
}

type GetSet interface {
	Set(string) int
	Get(int) (string, error)
}

func (m *URLMap) Set(val string) int {
	if m.data == nil {
		m.data = &sync.Map{}
		m.count = 0
	}
	m.count++
	m.data.Store(m.count, val)
	return m.count
}
func (m *URLMap) Get(key int) (string, error) {
	if m.data == nil {
		return "", errors.New(ErrNotFound)
	}
	if res, ok := m.data.Load(key); ok {
		return res.(string), nil
	} else {
		return "", errors.New(ErrNotFound)
	}
}
