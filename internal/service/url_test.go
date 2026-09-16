package service

import "testing"

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
