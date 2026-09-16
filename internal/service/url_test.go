package service

import (
	"errors"
	"testing"
)

type mockURLRepository struct {
	saveCalls  int
	saveErrors []error
}

func (m *mockURLRepository) Save(code string, url string) error {
	m.saveCalls++

	if len(m.saveErrors) == 0 {
		return nil
	}

	err := m.saveErrors[0]
	m.saveErrors = m.saveErrors[1:]
	return err
}

func (m *mockURLRepository) Get(code string) (string, error) {
	return "", nil
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
