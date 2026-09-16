package service

import (
	"crypto/rand"
	"errors"
	"math/big"
	"net/url"
	"strings"
	"time"
)

var ErrDuplicateCode = errors.New("duplicate URL code")
var ErrInvalidURL = errors.New("invalid URL")
var ErrReservedAlias = errors.New("reserved alias")
var ErrURLExpired = errors.New("URL has expired")
var ErrInvalidExpiration = errors.New("expiration must be in the future")

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type URLStats struct {
	Code       string `json:"code"`
	URL        string `json:"url"`
	ClickCount int    `json:"click_count"`
}

type URL struct {
	Code      string
	URL       string
	ExpiresAt *time.Time
}

type URLRepository interface {
	Save(code string, url string, expiresAt *time.Time) error
	Get(code string) (URL, error)
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
	return s.CreateURLWithExpiration(url, nil)
}

func (s *URLService) CreateURLWithExpiration(rawURL string, expiresAt *time.Time) (string, error) {
	if err := ValidateURL(rawURL); err != nil {
		return "", ErrInvalidURL
	}

	if expiresAt != nil && !expiresAt.After(time.Now()) {
		return "", ErrInvalidExpiration
	}

	const maxAttempts = 5

	for attempt := 0; attempt < maxAttempts; attempt++ {
		code, err := GenerateCode(6)
		if err != nil {
			return "", err
		}

		err = s.repository.Save(code, rawURL, expiresAt)
		if err == nil {
			return code, nil
		}

		if errors.Is(err, ErrDuplicateCode) {
			continue
		}

		return "", err
	}

	return "", ErrDuplicateCode
}

func (s *URLService) GetURL(code string) (string, error) {
	result, err := s.repository.Get(code)
	if err != nil {
		return "", err
	}

	if result.ExpiresAt != nil && !result.ExpiresAt.After(time.Now()) {
		return "", ErrURLExpired
	}

	return result.URL, nil
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
	return s.CreateURLWithAliasAndExpiration(rawURL, alias, nil)
}

func (s *URLService) CreateURLWithAliasAndExpiration(rawURL string, alias string, expiresAt *time.Time) (string, error) {
	if err := ValidateURL(rawURL); err != nil {
		return "", ErrInvalidURL
	}

	if err := ValidateAlias(alias); err != nil {
		return "", err
	}

	if expiresAt != nil && !expiresAt.After(time.Now()) {
		return "", ErrInvalidExpiration
	}

	if err := s.repository.Save(alias, rawURL, expiresAt); err != nil {
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
