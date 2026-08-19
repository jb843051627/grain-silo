package store

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"grain-silo/internal/model"
)

func TestBug002_ReadingListBySiloReturnsCopy(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	store := NewReadingStore(db)
	for i := 0; i < 2; i++ {
		store.Create(&model.Reading{
			ID:          fmt.Sprintf("r%d", i),
			SiloID:      "silo-1",
			Temp:        25.0,
			Humidity:    60.0,
			RecordedAt:  time.Now().UTC(),
		})
	}

	first, _ := store.ListBySilo("silo-1", 10)
	if len(first) == 0 {
		t.Fatal("expected readings, got empty")
	}
	first[0].Temp = 999.0

	second, _ := store.ListBySilo("silo-1", 10)
	if second[0].Temp == 999.0 {
		t.Error("ListBySilo returned shared reference - mutation polluted store data")
	}
}