package service

import (
	"path/filepath"
	"sync"
	"testing"
	"time"

	"grain-silo/internal/model"
	"grain-silo/internal/store"
)

func newTestReadingService(t *testing.T) *ReadingService {
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

func TestBug003_RecordReadingConcurrent(t *testing.T) {
	svc := newTestReadingService(t)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			svc.RecordReading("silo-1", 25.0, 60.0)
		}()
	}
	wg.Wait()

	latest, err := svc.GetLatestReading("silo-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if latest == nil {
		t.Fatal("expected latest reading, got nil")
	}
}