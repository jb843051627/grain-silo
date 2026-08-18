package store

import (
	"database/sql"
	"fmt"

	"grain-silo/internal/model"
)

// InspectionStore 质检存储
type InspectionStore struct {
	db *DB
}

// NewInspectionStore 创建质检存储
func NewInspectionStore(db *DB) *InspectionStore {
	return &InspectionStore{db: db}
}

// Create 创建质检记录
func (s *InspectionStore) Create(r *model.InspectionRecord) error {
	pestVal := 0
	if r.PestDetected {
		pestVal = 1
	}
	_, err := s.db.conn.Exec(
		`INSERT INTO inspections (id, silo_id, grain_type, moisture, impurity, pest_detected, inspector, result, inspected_at, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.ID, r.SiloID, r.GrainType, r.Moisture, r.Impurity, pestVal, r.Inspector, r.Result, r.InspectedAt, r.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert inspection: %w", err)
	}
	return nil
}

// GetByID 根据ID查询质检记录
func (s *InspectionStore) GetByID(id string) (*model.InspectionRecord, error) {
	row := s.db.conn.QueryRow(
		`SELECT id, silo_id, grain_type, moisture, impurity, pest_detected, inspector, result, inspected_at, created_at FROM inspections WHERE id = ?`, id,
	)
	var r model.InspectionRecord
	var pestVal int
	err := row.Scan(&r.ID, &r.SiloID, &r.GrainType, &r.Moisture, &r.Impurity, &pestVal, &r.Inspector, &r.Result, &r.InspectedAt, &r.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, model.ErrInspectionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query inspection: %w", err)
	}
	r.PestDetected = pestVal == 1
	return &r, nil
}

// ListBySilo 按筒仓查询质检记录
func (s *InspectionStore) ListBySilo(siloID string) ([]*model.InspectionRecord, error) {
	rows, err := s.db.conn.Query(
		`SELECT id, silo_id, grain_type, moisture, impurity, pest_detected, inspector, result, inspected_at, created_at FROM inspections WHERE silo_id = ? ORDER BY inspected_at DESC`,
		siloID,
	)
	if err != nil {
		return nil, fmt.Errorf("query inspections by silo: %w", err)
	}
	defer rows.Close()

	var records []*model.InspectionRecord
	for rows.Next() {
		var r model.InspectionRecord
		var pestVal int
		err := rows.Scan(&r.ID, &r.SiloID, &r.GrainType, &r.Moisture, &r.Impurity, &pestVal, &r.Inspector, &r.Result, &r.InspectedAt, &r.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan inspection: %w", err)
		}
		r.PestDetected = pestVal == 1
		records = append(records, &r)
	}
	return records, nil
}

// ListAll 查询全部质检记录
func (s *InspectionStore) ListAll() ([]*model.InspectionRecord, error) {
	rows, err := s.db.conn.Query(
		`SELECT id, silo_id, grain_type, moisture, impurity, pest_detected, inspector, result, inspected_at, created_at FROM inspections ORDER BY inspected_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("query all inspections: %w", err)
	}
	defer rows.Close()

	var records []*model.InspectionRecord
	for rows.Next() {
		var r model.InspectionRecord
		var pestVal int
		err := rows.Scan(&r.ID, &r.SiloID, &r.GrainType, &r.Moisture, &r.Impurity, &pestVal, &r.Inspector, &r.Result, &r.InspectedAt, &r.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan inspection: %w", err)
		}
		r.PestDetected = pestVal == 1
		records = append(records, &r)
	}
	return records, nil
}

// ListByResult 按结果查询
func (s *InspectionStore) ListByResult(result string) ([]*model.InspectionRecord, error) {
	rows, err := s.db.conn.Query(
		`SELECT id, silo_id, grain_type, moisture, impurity, pest_detected, inspector, result, inspected_at, created_at FROM inspections WHERE result = ? ORDER BY inspected_at DESC`,
		result,
	)
	if err != nil {
		return nil, fmt.Errorf("query inspections by result: %w", err)
	}
	defer rows.Close()

	var records []*model.InspectionRecord
	for rows.Next() {
		var r model.InspectionRecord
		var pestVal int
		err := rows.Scan(&r.ID, &r.SiloID, &r.GrainType, &r.Moisture, &r.Impurity, &pestVal, &r.Inspector, &r.Result, &r.InspectedAt, &r.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan inspection: %w", err)
		}
		r.PestDetected = pestVal == 1
		records = append(records, &r)
	}
	return records, nil
}

// ListPestDetected 查询检测到虫害的记录
func (s *InspectionStore) ListPestDetected() ([]*model.InspectionRecord, error) {
	rows, err := s.db.conn.Query(
		`SELECT id, silo_id, grain_type, moisture, impurity, pest_detected, inspector, result, inspected_at, created_at FROM inspections WHERE pest_detected = 1 ORDER BY inspected_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("query pest detected: %w", err)
	}
	defer rows.Close()

	var records []*model.InspectionRecord
	for rows.Next() {
		var r model.InspectionRecord
		var pestVal int
		err := rows.Scan(&r.ID, &r.SiloID, &r.GrainType, &r.Moisture, &r.Impurity, &pestVal, &r.Inspector, &r.Result, &r.InspectedAt, &r.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan inspection: %w", err)
		}
		r.PestDetected = pestVal == 1
		records = append(records, &r)
	}
	return records, nil
}

// GetLatest 获取最新质检记录
func (s *InspectionStore) GetLatest(siloID string) (*model.InspectionRecord, error) {
	row := s.db.conn.QueryRow(
		`SELECT id, silo_id, grain_type, moisture, impurity, pest_detected, inspector, result, inspected_at, created_at FROM inspections WHERE silo_id = ? ORDER BY inspected_at DESC LIMIT 1`,
		siloID,
	)
	var r model.InspectionRecord
	var pestVal int
	err := row.Scan(&r.ID, &r.SiloID, &r.GrainType, &r.Moisture, &r.Impurity, &pestVal, &r.Inspector, &r.Result, &r.InspectedAt, &r.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, model.ErrInspectionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query latest inspection: %w", err)
	}
	r.PestDetected = pestVal == 1
	return &r, nil
}
