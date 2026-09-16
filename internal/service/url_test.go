package service

import (
	"errors"
	"testing"
	"time"
)

type mockURLRepository struct {
	saveCalls  int
	saveErrors []error
}

func (m *mockURLRepository) Save(code string, url string, expiresAt *time.Time) error {
	m.saveCalls++

	if len(m.saveErrors) == 0 {
		return nil
	}
	err := m.saveErrors[0]
	m.saveErrors = m.saveErrors[1:]
	return err
}

func (m *mockURLRepository) Get(code string) (URL, error) {
	return URL{}, nil
}

func (m *mockURLRepository) GetStats(code string) (URLStats, error) {
	return URLStats{}, nil
}

func (m *mockURLRepository) IncrementClickCount(code string) error {
	return nil
}

func TestGenerateCode(t *testing.T) {
	code, err := GenerateCode(6)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(code) != 6 {
		t.Fatalf("expected code length of 6, got %d", len(code))
	}
}

func TestCreateURLRetriesOnDuplicateCode(t *testing.T) {
	repo := &mockURLRepository{
		saveErrors: []error{
			ErrDuplicateCode,
			nil,
		},
	}

	svc := NewURLService(repo)

	code, err := svc.CreateURL("http://example.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if code == "" {
		t.Fatal("expected a code, got empty string")
	}

	if repo.saveCalls != 2 {
		t.Fatalf("expected 2 save calls, got %d", repo.saveCalls)
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid https",
			url:     "https://example.com",
			wantErr: false,
		},
		{
			name:    "valid http",
			url:     "http://example.com",
			wantErr: false,
		},
		{
			name:    "missing scheme",
			url:     "example.com",
			wantErr: true,
		},
		{
			name:    "unsupported scheme",
			url:     "ftp://example.com",
			wantErr: true,
		},
		{
			name:    "empty url",
			url:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateURL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
		})
	}
}

func TestCreateURLRejectsInvalidURL(t *testing.T) {
	repo := &mockURLRepository{}
	svc := NewURLService(repo)

	_, err := svc.CreateURL("not-a-url")

	if !errors.Is(err, ErrInvalidURL) {
		t.Fatalf("expected ErrInvalidURL, got %v", err)
	}

	if repo.saveCalls != 0 {
		t.Fatalf("expected Save not to be called, got %d calls", repo.saveCalls)
	}
}

func TestCreateURLWithAlias(t *testing.T) {
	repo := &mockURLRepository{}
	svc := NewURLService(repo)

	code, err := svc.CreateURLWithAlias("http://example.com", "google")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if code != "google" {
		t.Fatalf("expected code google, got %s", code)
	}

	if repo.saveCalls != 1 {
		t.Fatalf("expected 1 save call got %d", repo.saveCalls)
	}
}

func TestValidateAlias(t *testing.T) {
	tests := []struct {
		name    string
		alias   string
		wantErr bool
	}{
		{
			name:    "valid alias",
			alias:   "google",
			wantErr: false,
		},
		{
			name:    "minimum length",
			alias:   "abc",
			wantErr: false,
		},
		{
			name:    "too short",
			alias:   "ab",
			wantErr: true,
		},
		{
			name:    "too long",
			alias:   "abcdefghijklmnopq",
			wantErr: true,
		},
		{
			name:    "contains slash",
			alias:   "foo/bar",
			wantErr: true,
		},
		{
			name:    "contains space",
			alias:   "foo bar",
			wantErr: true,
		},
		{
			name:    "contains question mark",
			alias:   "foo?bar",
			wantErr: true,
		},
		{
			name:    "contains hash",
			alias:   "foo#bar",
			wantErr: true,
		},
		{
			name:    "empty",
			alias:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAlias(tt.alias)

			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"ValidateAlias(%q) error = %v, wantErr %v",
					tt.alias,
					err,
					tt.wantErr,
				)
			}
		})
	}
}

func TestValidateAliasRejectsReservedAliases(t *testing.T) {
	reservedAliases := []string{
		"health",
		"hello",
		"api",
	}

	for _, alias := range reservedAliases {
		t.Run(alias, func(t *testing.T) {
			err := ValidateAlias(alias)

			if !errors.Is(err, ErrReservedAlias) {
				t.Fatalf("expected ErrReservedAlias, got %v", err)
			}
		})
	}
}

func TestCreateURLWithExpiration(t *testing.T) {
	repo := &mockURLRepository{}

	svc := NewURLService(repo)

	expiresAt := time.Now().Add(time.Hour)

	code, err := svc.CreateURLWithExpiration("https://example.com", &expiresAt)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if code == "" {
		t.Fatal("expected code, got empty string")
	}

	if repo.saveCalls != 1 {
		t.Fatalf("expected 1 save call, got %d", repo.saveCalls)
	}
}
