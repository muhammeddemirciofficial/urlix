package service

import (
	"crypto/rand"
	"errors"
	"math/big"
	"net/url"
	"strings"
)

var ErrDuplicateCode = errors.New("duplicate URL code")
var ErrInvalidURL = errors.New("invalid URL")
var ErrReservedAlias = errors.New("reserved alias")

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
	if err := ValidateURL(url); err != nil {
		return "", ErrInvalidURL
	}

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

func (s *URLService) CreateURLWithAlias(rawURL string, alias string) (string, error) {
	if err := ValidateURL(rawURL); err != nil {
		return "", ErrInvalidURL
	}

	if err := ValidateAlias(alias); err != nil {
		return "", err
	}

	if err := s.repository.Save(alias, rawURL); err != nil {
		if errors.Is(err, ErrDuplicateCode) {
			return "", ErrDuplicateCode
		}

		return "", err
	}

	return alias, nil
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

func ValidateURL(rawURL string) error {
	if rawURL == "" {
		return errors.New("url is required")
	}

	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return errors.New("invalid url format")
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New("URL must start with http or https")
	}

	if parsed.Host == "" {
		return errors.New("url host is required")
	}

	return nil
}

func ValidateAlias(alias string) error {
	if alias == "" {
		return errors.New("alias is required")
	}

	if len(alias) < 3 {
		return errors.New("alias must be at least 3 characters long")
	}

	if len(alias) > 16 {
		return errors.New("alias must be no more than 16 characters long")
	}

	if strings.ContainsAny(alias, " /?#%") {
		return errors.New("alias contains invalid characters")
	}

	var reservedAliases = map[string]struct{}{
		"health": {},
		"hello":  {},
		"api":    {},
		"r":      {},
	}

	if _, exists := reservedAliases[alias]; exists {
		return ErrReservedAlias
	}

	return nil
}
