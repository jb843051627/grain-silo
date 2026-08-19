package service

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"grain-silo/internal/model"
	"grain-silo/internal/store"
)

func newTestReadingServiceForCtx(t *testing.T) *ReadingService {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := store.NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	siloStore := store.NewSiloStore(db)
	readingStore := store.NewReadingStore(db)
	alertStore := store.NewAlertStore(db)
	auditStore := store.NewAuditStore(db)

	siloStore.Create(&model.Silo{
		ID:        "silo-1",
		Name:      "test-silo",
		Capacity:  100,
		Status:    model.SILO_EMPTY,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})

	return NewReadingService(readingStore, siloStore, alertStore, auditStore)
}

func TestBug005_BatchIngestRespectsContext(t *testing.T) {
	svc := newTestReadingServiceForCtx(t)

	readings := make([]*model.Reading, 100)
	for i := range readings {
		readings[i] = &model.Reading{
			SiloID:     "silo-1",
			Temp:       25.0,
			Humidity:   60.0,
			RecordedAt: time.Now().UTC(),
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	count, err := svc.BatchIngestReadings(ctx, readings)
	if err == nil {
		t.Error("expected error when context is cancelled")
	}
	if count > 0 {
		t.Errorf("expected 0 readings processed, got %d", count)
	}
}