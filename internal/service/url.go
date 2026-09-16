package service

import (
	"crypto/rand"
	"errors"
	"math/big"
)

var ErrDuplicateCode = errors.New("duplicate URL code")

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type URLStats struct {
	Code       string `json:"code"`
	URL        string `json:"url"`
	ClickCount int    `json:"click_count"`
}

type URLRepository interface {
	Save(code string, url string) error
	Get(code string) (string, error)
	GetStats(code string) (URLStats, error)
	IncrementClickCount(code string) error
}

type URLService struct {
	repository URLRepository
}

func NewURLService(repository URLRepository) *URLService {
	return &URLService{
		repository: repository,
	}
}

func (s *URLService) CreateURL(url string) (string, error) {
	const maxAttempts = 5

	for attempt := 0; attempt < maxAttempts; attempt++ {
		code, err := GenerateCode(6)
		if err != nil {
			return "", err
		}

		err = s.repository.Save(code, url)
		if err == nil {
			return code, nil
		}

		if !errors.Is(err, ErrDuplicateCode) {
			return "", err
		}
	}

	return "", ErrDuplicateCode
}

func (s *URLService) GetURL(code string) (string, error) {
	url, err := s.repository.Get(code)
	if err != nil {
		return "", err
	}
	return url, nil
}

func (s *URLService) GetStats(code string) (URLStats, error) {
	stats, err := s.repository.GetStats(code)
	if err != nil {
		return URLStats{}, err
	}
	return stats, nil
}

func (s *URLService) IncrementClickCount(code string) error {
	return s.repository.IncrementClickCount(code)
}

func GenerateCode(length int) (string, error) {
	code := make([]byte, length)

	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		code[i] = charset[n.Int64()]
	}

	return string(code), nil
}
