package store

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"grain-silo/internal/model"
)

func TestBug008_TransferListBySiloReturnsCopy(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	store := NewTransferStore(db)

	// Create some transfers
	for i := 0; i < 3; i++ {
		store.Create(&model.TransferRecord{
			ID:          fmt.Sprintf("t%d", i),
			FromSiloID:  "silo-1",
			ToSiloID:    "silo-2",
			GrainType:   model.GRAIN_WHEAT,
			Quantity:    10.0,
			Status:      model.TRANSFER_COMPLETED,
			ScheduledAt: time.Now().UTC(),
			Operator:    "op",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		})
	}

	first, err := store.ListBySilo("silo-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(first) == 0 {
		t.Fatal("expected transfers, got empty")
	}

	// Mutate the returned slice
	originalQty := first[0].Quantity
	first[0].Quantity = 999.0

	// Get again - should not see the mutation
	second, err := store.ListBySilo("silo-1")
	if err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}
	if second[0].Quantity == 999.0 {
		t.Error("ListBySilo returned shared reference - mutation polluted store cache")
	}
	if second[0].Quantity != originalQty {
		t.Errorf("expected quantity %f, got %f", originalQty, second[0].Quantity)
	}
}