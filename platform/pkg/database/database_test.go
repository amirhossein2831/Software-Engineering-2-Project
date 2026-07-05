package database

import (
	"os"
	"testing"
)

func TestOpenEmptyDSN(t *testing.T) {
	if _, err := Open(""); err == nil {
		t.Error("Open(\"\") should return an error")
	}
}

func TestOpenRealDatabase(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping real-database test")
	}
	db, err := Open(dsn)
	if err != nil {
		t.Fatalf("Open(TEST_DATABASE_URL) error: %v", err)
	}
	var one int
	if err := db.Raw("SELECT 1").Scan(&one).Error; err != nil {
		t.Fatalf("SELECT 1 failed: %v", err)
	}
	if one != 1 {
		t.Errorf("SELECT 1 = %d; want 1", one)
	}
}
