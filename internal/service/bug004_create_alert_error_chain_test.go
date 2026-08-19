package service

import (
	"errors"
	"path/filepath"
	"testing"

	"grain-silo/internal/model"
	"grain-silo/internal/store"
)

func newTestAlertService(t *testing.T) *AlertService {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := store.NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	siloStore := store.NewSiloStore(db)
	alertStore := store.NewAlertStore(db)
	auditStore := store.NewAuditStore(db)

	return NewAlertService(alertStore, siloStore, auditStore)
}

func TestBug004_CreateAlertErrorChain(t *testing.T) {
	svc := newTestAlertService(t)

	_, err := svc.CreateAlert("nonexistent-silo", model.ALERT_CRITICAL, "test alert")
	if err == nil {
		t.Fatal("expected error for nonexistent silo")
	}
	if !errors.Is(err, model.ErrSiloNotFound) {
		t.Errorf("expected error chain to contain ErrSiloNotFound, got: %v", err)
	}
}