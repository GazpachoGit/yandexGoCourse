package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"sync"
)

type fileData struct {
	Data []fileRecord `json:"data"`
}
type fileRecord struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}

type URLMap struct {
	data    *sync.Map
	count   int
	file    *os.File
	scanner *bufio.Scanner
	writer  *bufio.Writer
}

func NewURLMap(FilePath string) (*URLMap, error) {
	file, err := os.OpenFile(FilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return nil, err
	}
	newURLMap := &URLMap{
		file:    file,
		scanner: bufio.NewScanner(file),
		writer:  bufio.NewWriter(file),
		data:    &sync.Map{},
		count:   0,
	}
	if err = newURLMap.getDataFromFile(); err != nil {
		return nil, err
	}
	return newURLMap, nil
}

func (m *URLMap) GetCount() int {
	return m.count
}

func (m *URLMap) Close() error {
	return m.file.Close()
}

func (m *URLMap) getDataFromFile() error {
	dataArray := make([]*fileRecord, 0)
	for m.scanner.Scan() {
		data := m.scanner.Bytes()
		record := &fileRecord{}
		err := json.Unmarshal(data, record)
		if err != nil {
			return err
		}
		dataArray = append(dataArray, record)
	}
	if err := m.scanner.Err(); err != nil {
		return err
	}
	for _, val := range dataArray {
		m.count = m.setNewValToMap(val)
	}

	return nil
}

func (m *URLMap) setNewValToMap(record *fileRecord) int {
	m.count = record.ID
	m.data.Store(record.ID, record.URL)
	return m.count
}

func (m *URLMap) putFileRecord(record *fileRecord) error {
	newRecord, err := json.Marshal(record)
	if err != nil {
		return err
	}
	if _, err := m.writer.Write(newRecord); err != nil {
		return err
	}

	if err := m.writer.WriteByte('\n'); err != nil {
		return err
	}

	return m.writer.Flush()
}

func (m *URLMap) Set(val string) (int, error) {
	m.count++
	record := &fileRecord{ID: m.count, URL: val}
	err := m.putFileRecord(record)
	if err != nil {
		return 0, err
	}
	m.data.Store(m.count, val)
	return m.count, nil
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
