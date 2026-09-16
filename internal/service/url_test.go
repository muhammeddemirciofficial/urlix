package service

import "testing"

func TestGenerateCode(t *testing.T) {
	code, err := GenerateCode(6)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(code) != 6 {
		t.Fatalf("expected code length of 6, got %d", len(code))
	}
}
