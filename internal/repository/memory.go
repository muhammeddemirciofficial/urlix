package repository

import (
	"database/sql"
	"sync"

	"github.com/muhammeddemirciofficial/urlix/internal/service"
)

type MemoryURLRepository struct {
	mu   sync.RWMutex
	urls map[string]string
}

func NewMemoryURLRepository() *MemoryURLRepository {
	return &MemoryURLRepository{
		urls: make(map[string]string),
	}
}

func (s *MemoryURLRepository) Save(code, url string) error {
	if _, exists := s.urls[code]; exists {
		return service.ErrDuplicateCode
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.urls[code] = url

	return nil
}

func (s *MemoryURLRepository) Get(code string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, ok := s.urls[code]

	if !ok {
		return "", sql.ErrNoRows
	}

	return url, nil
}

func (s *MemoryURLRepository) IncrementClickCount(code string) error {
	// In-memory repository does not track click counts, so this is a no-op.
	return nil
}

func (r *MemoryURLRepository) GetStats(code string) (service.URLStats, error) {
	url, err := r.Get(code)
	if err != nil {
		return service.URLStats{}, err
	}

	return service.URLStats{
		Code:       code,
		URL:        url,
		ClickCount: 0,
	}, nil
}
