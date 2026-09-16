package repository

import (
	"database/sql"
	"sync"
	"time"

	"github.com/muhammeddemirciofficial/urlix/internal/service"
)

type memoryURL struct {
	id        int64
	url       string
	expiresAt *time.Time
}

type memoryClick struct {
	urlID     int64
	userAgent string
	referer   string
	clickedAt time.Time
}

type MemoryURLRepository struct {
	mu     sync.RWMutex
	urls   map[string]memoryURL
	clicks []memoryClick
	nextID int64
}

func NewMemoryURLRepository() *MemoryURLRepository {
	return &MemoryURLRepository{
		urls:   make(map[string]memoryURL),
		clicks: make([]memoryClick, 0),
		nextID: 1,
	}
}

func (s *MemoryURLRepository) Save(code, url string, expiresAt *time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.urls[code]; exists {
		return service.ErrDuplicateCode
	}

	s.urls[code] = memoryURL{
		id:        s.nextID,
		url:       url,
		expiresAt: expiresAt,
	}

	s.nextID++

	return nil
}

func (s *MemoryURLRepository) Get(code string) (service.URL, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result, ok := s.urls[code]
	if !ok {
		return service.URL{}, sql.ErrNoRows
	}

	return service.URL{
		ID:        result.id,
		Code:      code,
		URL:       result.url,
		ExpiresAt: result.expiresAt,
	}, nil
}

func (s *MemoryURLRepository) IncrementClickCount(code string) error {
	return nil
}

func (s *MemoryURLRepository) GetStats(code string) (service.URLStats, error) {
	result, err := s.Get(code)
	if err != nil {
		return service.URLStats{}, err
	}

	return service.URLStats{
		Code:       code,
		URL:        result.URL,
		ClickCount: 0,
	}, nil
}

func (s *MemoryURLRepository) RecordClick(
	urlID int64,
	userAgent string,
	referer string,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.clicks = append(s.clicks, memoryClick{
		urlID:     urlID,
		userAgent: userAgent,
		referer:   referer,
		clickedAt: time.Now(),
	})

	return nil
}

func (s *MemoryURLRepository) GetAnalytics(code string) (service.URLAnalytics, error) {
	result, err := s.Get(code)
	if err != nil {
		return service.URLAnalytics{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	analytics := service.URLAnalytics{
		Code:       code,
		URL:        result.URL,
		ClickCount: 0,
		Clicks:     make([]service.Click, 0),
	}

	for _, click := range s.clicks {
		if click.urlID == result.ID {
			analytics.Clicks = append(analytics.Clicks, service.Click{
				ClickedAt: click.clickedAt,
				UserAgent: click.userAgent,
				Referer:   click.referer,
			})
		}
	}

	analytics.ClickCount = len(analytics.Clicks)

	return analytics, nil
}
