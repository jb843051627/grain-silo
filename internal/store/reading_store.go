package store

import (
	"database/sql"
	"fmt"
	"time"

	"grain-silo/internal/model"
)

// ReadingStore 读数存储
type ReadingStore struct {
	db *DB
}

// NewReadingStore 创建读数存储
func NewReadingStore(db *DB) *ReadingStore {
	return &ReadingStore{db: db}
}

// Create 创建读数
func (s *ReadingStore) Create(r *model.Reading) error {
	_, err := s.db.conn.Exec(
		`INSERT INTO readings (id, silo_id, temp, humidity, recorded_at) VALUES (?, ?, ?, ?, ?)`,
		r.ID, r.SiloID, r.Temp, r.Humidity, r.RecordedAt,
	)
	if err != nil {
		return fmt.Errorf("insert reading: %w", err)
	}
	return nil
}

// GetByID 根据ID查询读数
func (s *ReadingStore) GetByID(id string) (*model.Reading, error) {
	row := s.db.conn.QueryRow(
		`SELECT id, silo_id, temp, humidity, recorded_at FROM readings WHERE id = ?`, id,
	)
	var r model.Reading
	err := row.Scan(&r.ID, &r.SiloID, &r.Temp, &r.Humidity, &r.RecordedAt)
	if err == sql.ErrNoRows {
		return nil, model.ErrReadingNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query reading: %w", err)
	}
	return &r, nil
}

// ListBySilo 按筒仓查询读数列表
func (s *ReadingStore) ListBySilo(siloID string, limit int) ([]*model.Reading, error) {
	rows, err := s.db.conn.Query(
		`SELECT id, silo_id, temp, humidity, recorded_at FROM readings WHERE silo_id = ? ORDER BY recorded_at DESC LIMIT ?`,
		siloID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("query readings by silo: %w", err)
	}
	defer rows.Close()

	var readings []*model.Reading
	for rows.Next() {
		var r model.Reading
		err := rows.Scan(&r.ID, &r.SiloID, &r.Temp, &r.Humidity, &r.RecordedAt)
		if err != nil {
			return nil, fmt.Errorf("scan reading: %w", err)
		}
		readings = append(readings, &r)
	}
	return readings, nil
}

// ListByTimeRange 按时间范围查询
func (s *ReadingStore) ListByTimeRange(siloID string, start, end time.Time) ([]*model.Reading, error) {
	rows, err := s.db.conn.Query(
		`SELECT id, silo_id, temp, humidity, recorded_at FROM readings WHERE silo_id = ? AND recorded_at >= ? AND recorded_at <= ? ORDER BY recorded_at DESC`,
		siloID, start, end,
	)
	if err != nil {
		return nil, fmt.Errorf("query readings by time: %w", err)
	}
	defer rows.Close()

	var readings []*model.Reading
	for rows.Next() {
		var r model.Reading
		err := rows.Scan(&r.ID, &r.SiloID, &r.Temp, &r.Humidity, &r.RecordedAt)
		if err != nil {
			return nil, fmt.Errorf("scan reading: %w", err)
		}
		readings = append(readings, &r)
	}
	return readings, nil
}

// GetLatest 获取最新读数
func (s *ReadingStore) GetLatest(siloID string) (*model.Reading, error) {
	row := s.db.conn.QueryRow(
		`SELECT id, silo_id, temp, humidity, recorded_at FROM readings WHERE silo_id = ? ORDER BY recorded_at DESC LIMIT 1`,
		siloID,
	)
	var r model.Reading
	err := row.Scan(&r.ID, &r.SiloID, &r.Temp, &r.Humidity, &r.RecordedAt)
	if err == sql.ErrNoRows {
		return nil, model.ErrReadingNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query latest reading: %w", err)
	}
	return &r, nil
}

// GetAvg24h 获取24小时平均
func (s *ReadingStore) GetAvg24h(siloID string) (avgTemp, avgHumidity float64, err error) {
	since := time.Now().UTC().Add(-24 * time.Hour)
	row := s.db.conn.QueryRow(
		`SELECT COALESCE(AVG(temp),0), COALESCE(AVG(humidity),0) FROM readings WHERE silo_id = ? AND recorded_at >= ?`,
		siloID, since,
	)
	err = row.Scan(&avgTemp, &avgHumidity)
	if err != nil {
		return 0, 0, fmt.Errorf("query avg 24h: %w", err)
	}
	return
}

// BatchCreate 批量创建读数
func (s *ReadingStore) BatchCreate(readings []*model.Reading) (int, error) {
	tx, err := s.db.conn.Begin()
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	for i, r := range readings {
		_, err := tx.Exec(
			`INSERT INTO readings (id, silo_id, temp, humidity, recorded_at) VALUES (?, ?, ?, ?, ?)`,
			r.ID, r.SiloID, r.Temp, r.Humidity, r.RecordedAt,
		)
		if err != nil {
			tx.Rollback()
			return i, fmt.Errorf("batch insert reading at %d: %w", i, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit batch: %w", err)
	}
	return len(readings), nil
}

// DeleteOlderThan 删除旧数据
func (s *ReadingStore) DeleteOlderThan(before time.Time) (int64, error) {
	res, err := s.db.conn.Exec(`DELETE FROM readings WHERE recorded_at < ?`, before)
	if err != nil {
		return 0, fmt.Errorf("delete old readings: %w", err)
	}
	count, _ := res.RowsAffected()
	return count, nil
}
