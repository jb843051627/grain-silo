package store

import (
	"path/filepath"
	"testing"

	"grain-silo/internal/model"
)

func TestBug001_SiloGetByIDReturnsErrorWhenNotFound(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	store := NewSiloStore(db)
	silo, err := store.GetByID("nonexistent-id")
	if err == nil {
		t.Error("expected error when silo not found, got nil")
	}
	if err != model.ErrSiloNotFound {
		t.Errorf("expected ErrSiloNotFound, got: %v", err)
	}
	if silo != nil {
		t.Error("expected nil silo when not found")
	}
}