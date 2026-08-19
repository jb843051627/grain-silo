package store

import (
	"path/filepath"
	"testing"

	"grain-silo/internal/model"
)

func TestBug006_TransferGetByIDReturnsErrorWhenNotFound(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	store := NewTransferStore(db)
	transfer, err := store.GetByID("nonexistent-id")
	if err == nil {
		t.Error("expected error when transfer not found, got nil")
	}
	if err != model.ErrTransferNotFound {
		t.Errorf("expected ErrTransferNotFound, got: %v", err)
	}
	if transfer != nil {
		t.Error("expected nil transfer when not found")
	}
}