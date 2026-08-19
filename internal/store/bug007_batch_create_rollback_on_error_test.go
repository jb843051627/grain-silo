package store

import (
	"path/filepath"
	"testing"
	"time"

	"grain-silo/internal/model"
)

func TestBug007_BatchCreateRollbackOnError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	store := NewReadingStore(db)
	store.Create(&model.Reading{
		ID:         "valid-1",
		SiloID:     "silo-1",
		Temp:       25.0,
		Humidity:   60.0,
		RecordedAt: time.Now().UTC(),
	})

	readings := []*model.Reading{
		{ID: "dup-1", SiloID: "silo-1", Temp: 20.0, Humidity: 50.0, RecordedAt: time.Now().UTC()},
		{ID: "valid-1", SiloID: "silo-1", Temp: 30.0, Humidity: 70.0, RecordedAt: time.Now().UTC()},
	}

	count, err := store.BatchCreate(readings)
	if err == nil {
		t.Error("expected error on batch with duplicate ID")
	}
	if count != 1 {
		t.Logf("count = %d (expected 1 before failure)", count)
	}

	_, err = store.GetByID("dup-1")
	if err == nil {
		t.Error("expected dup-1 to be rolled back, but it exists")
	}
}