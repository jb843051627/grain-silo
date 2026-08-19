package service

import (
	"path/filepath"
	"testing"
	"time"

	"grain-silo/internal/model"
	"grain-silo/internal/store"
)

func newTestTransferServiceForExec(t *testing.T) (*TransferService, *store.SiloStore, *store.TransferStore) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := store.NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	siloStore := store.NewSiloStore(db)
	transferStore := store.NewTransferStore(db)
	auditStore := store.NewAuditStore(db)

	siloStore.Create(&model.Silo{
		ID: "s1", Name: "silo1", Capacity: 100, CurrentLoad: 50, Status: model.SILO_EMPTYING,
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	})
	siloStore.Create(&model.Silo{
		ID: "s2", Name: "silo2", Capacity: 10, CurrentLoad: 5, Status: model.SILO_FILLING,
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	})

	return NewTransferService(transferStore, siloStore, auditStore), siloStore, transferStore
}

func TestBug010_ExecuteTransferNoPartialDebit(t *testing.T) {
	svc, siloStore, transferStore := newTestTransferServiceForExec(t)

	now := time.Now().UTC()
	transferStore.Create(&model.TransferRecord{
		ID: "t1", FromSiloID: "s1", ToSiloID: "s2",
		GrainType: model.GRAIN_WHEAT, Quantity: 30, Status: model.TRANSFER_APPROVED,
		ScheduledAt: now, Operator: "op", CreatedAt: now, UpdatedAt: now,
	})

	_, err := svc.ExecuteTransfer("t1", "op")
	if err == nil {
		t.Fatal("expected ErrCapacityExceeded on transfer exceeding target capacity, got nil")
	}

	fromSilo, err := siloStore.GetByID("s1")
	if err != nil {
		t.Fatalf("get from silo: %v", err)
	}
	if fromSilo.CurrentLoad != 50 {
		t.Errorf("source silo current_load changed to %v; expected 50 (must not be debited when capacity check fails)", fromSilo.CurrentLoad)
	}
}