package store

import (
	"database/sql"
	"fmt"
	"time"

	"grain-silo/internal/model"
)

// SiloStore 筒仓储
type SiloStore struct {
	db *DB
}

// NewSiloStore 创建筒仓储
func NewSiloStore(db *DB) *SiloStore {
	return &SiloStore{db: db}
}

// Create 创建筒仓
func (s *SiloStore) Create(silo *model.Silo) error {
	_, err := s.db.conn.Exec(
		`INSERT INTO silos (id, name, capacity, current_load, grain_type, status, location, temp_sensor, humid_sensor, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		silo.ID, silo.Name, silo.Capacity, silo.CurrentLoad, silo.GrainType, silo.Status, silo.Location,
		silo.TempSensor, silo.HumidSensor, silo.CreatedAt, silo.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert silo: %w", err)
	}
	return nil
}

// GetByID 根据ID查询筒仓
func (s *SiloStore) GetByID(id string) (*model.Silo, error) {
	row := s.db.conn.QueryRow(
		`SELECT id, name, capacity, current_load, grain_type, status, location, temp_sensor, humid_sensor, created_at, updated_at
		 FROM silos WHERE id = ?`, id,
	)
	var silo model.Silo
	err := row.Scan(
		&silo.ID, &silo.Name, &silo.Capacity, &silo.CurrentLoad, &silo.GrainType,
		&silo.Status, &silo.Location, &silo.TempSensor, &silo.HumidSensor,
		&silo.CreatedAt, &silo.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, model.ErrSiloNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query silo: %w", err)
	}
	return &silo, nil
}

// List 查询筒仓列表
func (s *SiloStore) List() ([]*model.Silo, error) {
	rows, err := s.db.conn.Query(
		`SELECT id, name, capacity, current_load, grain_type, status, location, temp_sensor, humid_sensor, created_at, updated_at
		 FROM silos ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("query silos: %w", err)
	}
	defer rows.Close()

	var silos []*model.Silo
	for rows.Next() {
		var silo model.Silo
		err := rows.Scan(
			&silo.ID, &silo.Name, &silo.Capacity, &silo.CurrentLoad, &silo.GrainType,
			&silo.Status, &silo.Location, &silo.TempSensor, &silo.HumidSensor,
			&silo.CreatedAt, &silo.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan silo: %w", err)
		}
		silos = append(silos, &silo)
	}
	return silos, nil
}

// ListByStatus 按状态查询筒仓
func (s *SiloStore) ListByStatus(status model.SiloStatus) ([]*model.Silo, error) {
	rows, err := s.db.conn.Query(
		`SELECT id, name, capacity, current_load, grain_type, status, location, temp_sensor, humid_sensor, created_at, updated_at
		 FROM silos WHERE status = ? ORDER BY created_at DESC`, status,
	)
	if err != nil {
		return nil, fmt.Errorf("query silos by status: %w", err)
	}
	defer rows.Close()

	var silos []*model.Silo
	for rows.Next() {
		var silo model.Silo
		err := rows.Scan(
			&silo.ID, &silo.Name, &silo.Capacity, &silo.CurrentLoad, &silo.GrainType,
			&silo.Status, &silo.Location, &silo.TempSensor, &silo.HumidSensor,
			&silo.CreatedAt, &silo.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan silo: %w", err)
		}
		silos = append(silos, &silo)
	}
	return silos, nil
}

// Update 更新筒仓
func (s *SiloStore) Update(silo *model.Silo) error {
	silo.UpdatedAt = now()
	_, err := s.db.conn.Exec(
		`UPDATE silos SET name=?, capacity=?, current_load=?, grain_type=?, status=?, location=?, temp_sensor=?, humid_sensor=?, updated_at=? WHERE id=?`,
		silo.Name, silo.Capacity, silo.CurrentLoad, silo.GrainType, silo.Status, silo.Location,
		silo.TempSensor, silo.HumidSensor, silo.UpdatedAt, silo.ID,
	)
	if err != nil {
		return fmt.Errorf("update silo: %w", err)
	}
	return nil
}

// UpdateLoad 更新载量
func (s *SiloStore) UpdateLoad(id string, load float64) error {
	_, err := s.db.conn.Exec(
		`UPDATE silos SET current_load=?, updated_at=? WHERE id=?`,
		load, now(), id,
	)
	if err != nil {
		return fmt.Errorf("update silo load: %w", err)
	}
	return nil
}

// UpdateStatus 更新状态
func (s *SiloStore) UpdateStatus(id string, status model.SiloStatus) error {
	_, err := s.db.conn.Exec(
		`UPDATE silos SET status=?, updated_at=? WHERE id=?`,
		status, now(), id,
	)
	if err != nil {
		return fmt.Errorf("update silo status: %w", err)
	}
	return nil
}

// Delete 删除筒仓
func (s *SiloStore) Delete(id string) error {
	_, err := s.db.conn.Exec(`DELETE FROM silos WHERE id=?`, id)
	if err != nil {
		return fmt.Errorf("delete silo: %w", err)
	}
	return nil
}

// GetCapacity 获取总容量
func (s *SiloStore) GetCapacity() (totalCap, totalLoad float64, count int, err error) {
	row := s.db.conn.QueryRow(
		`SELECT COALESCE(SUM(capacity),0), COALESCE(SUM(current_load),0), COUNT(*) FROM silos`,
	)
	err = row.Scan(&totalCap, &totalLoad, &count)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("query capacity: %w", err)
	}
	return
}

// CountByGrainType 按谷物种类统计
func (s *SiloStore) CountByGrainType() (map[model.GrainType]float64, error) {
	rows, err := s.db.conn.Query(
		`SELECT grain_type, COALESCE(SUM(current_load),0) FROM silos WHERE current_load > 0 GROUP BY grain_type`,
	)
	if err != nil {
		return nil, fmt.Errorf("query by grain type: %w", err)
	}
	defer rows.Close()

	result := make(map[model.GrainType]float64)
	for rows.Next() {
		var gt string
		var load float64
		if err := rows.Scan(&gt, &load); err != nil {
			return nil, fmt.Errorf("scan grain type: %w", err)
		}
		result[model.GrainType(gt)] = load
	}
	return result, nil
}
