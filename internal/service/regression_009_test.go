package service

import (
	"path/filepath"
	"testing"

	"grain-silo/internal/store"
)

func newTestMaintenanceService(t *testing.T) *MaintenanceService {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := store.NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	maintStore := store.NewMaintenanceStore(db)
	siloStore := store.NewSiloStore(db)
	auditStore := store.NewAuditStore(db)

	return NewMaintenanceService(maintStore, siloStore, auditStore)
}

func TestBug009_CompleteMaintenanceReturnsErrorWhenNotFound(t *testing.T) {
	svc := newTestMaintenanceService(t)

	_, err := svc.CompleteMaintenance("nonexistent-id", "operator", 100.0)
	if err == nil {
		t.Fatal("expected error for nonexistent maintenance task")
	}
}