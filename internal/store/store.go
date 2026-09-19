package store

import (
	"errors"
	"sync"

	"github.com/AbePlays/go-url-shortener-system-design/utils"
)

type UrlStore struct {
	urlMap map[string]string
	mutex  sync.RWMutex
}

func New() *UrlStore {
	return &UrlStore{
		urlMap: make(map[string]string),
		mutex:  sync.RWMutex{},
	}
}

func (urlStore *UrlStore) AddUrl(url string) string {
	shortCode := utils.GenerateShortCode(8)

	urlStore.mutex.Lock()
	defer urlStore.mutex.Unlock()

	for true {
		if _, ok := urlStore.urlMap[shortCode]; !ok {
			break
		}

		shortCode = utils.GenerateShortCode(8)
	}

	urlStore.urlMap[shortCode] = url
	return shortCode
}

func (urlStore *UrlStore) GetUrl(code string) (string, error) {
	urlStore.mutex.RLock()
	url, ok := urlStore.urlMap[code]
	urlStore.mutex.RUnlock()

	if !ok {
		return "", errors.New("code not found")
	}

	return url, nil
}
