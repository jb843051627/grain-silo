package store

import (
	"database/sql"
	"fmt"

	"grain-silo/internal/model"
)

// MaintenanceStore 维护任务存储
type MaintenanceStore struct {
	db *DB
}

// NewMaintenanceStore 创建维护任务存储
func NewMaintenanceStore(db *DB) *MaintenanceStore {
	return &MaintenanceStore{db: db}
}

// Create 创建维护任务
func (s *MaintenanceStore) Create(m *model.MaintenanceTask) error {
	_, err := s.db.conn.Exec(
		`INSERT INTO maintenance_tasks (id, silo_id, type, description, status, scheduled_at, started_at, completed_at, technician, cost, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.ID, m.SiloID, m.Type, m.Description, m.Status, m.ScheduledAt, m.StartedAt, m.CompletedAt,
		m.Technician, m.Cost, m.CreatedAt, m.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert maintenance: %w", err)
	}
	return nil
}

// GetByID 根据ID查询维护任务
func (s *MaintenanceStore) GetByID(id string) (*model.MaintenanceTask, error) {
	row := s.db.conn.QueryRow(
		`SELECT id, silo_id, type, description, status, scheduled_at, started_at, completed_at, technician, cost, created_at, updated_at
		 FROM maintenance_tasks WHERE id = ?`, id,
	)
	var m model.MaintenanceTask
	var startedAt, completedAt sql.NullTime
	err := row.Scan(
		&m.ID, &m.SiloID, &m.Type, &m.Description, &m.Status, &m.ScheduledAt,
		&startedAt, &completedAt, &m.Technician, &m.Cost, &m.CreatedAt, &m.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, model.ErrMaintenanceNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query maintenance: %w", err)
	}
	if startedAt.Valid {
		m.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		m.CompletedAt = &completedAt.Time
	}
	return &m, nil
}

// ListBySilo 按筒仓查询维护任务
func (s *MaintenanceStore) ListBySilo(siloID string) ([]*model.MaintenanceTask, error) {
	rows, err := s.db.conn.Query(
		`SELECT id, silo_id, type, description, status, scheduled_at, started_at, completed_at, technician, cost, created_at, updated_at
		 FROM maintenance_tasks WHERE silo_id = ? ORDER BY scheduled_at DESC`, siloID,
	)
	if err != nil {
		return nil, fmt.Errorf("query maintenance by silo: %w", err)
	}
	defer rows.Close()

	var tasks []*model.MaintenanceTask
	for rows.Next() {
		var m model.MaintenanceTask
		var startedAt, completedAt sql.NullTime
		err := rows.Scan(
			&m.ID, &m.SiloID, &m.Type, &m.Description, &m.Status, &m.ScheduledAt,
			&startedAt, &completedAt, &m.Technician, &m.Cost, &m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan maintenance: %w", err)
		}
		if startedAt.Valid {
			m.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			m.CompletedAt = &completedAt.Time
		}
		tasks = append(tasks, &m)
	}
	return tasks, nil
}

// ListByStatus 按状态查询
func (s *MaintenanceStore) ListByStatus(status model.MaintenanceStatus) ([]*model.MaintenanceTask, error) {
	rows, err := s.db.conn.Query(
		`SELECT id, silo_id, type, description, status, scheduled_at, started_at, completed_at, technician, cost, created_at, updated_at
		 FROM maintenance_tasks WHERE status = ? ORDER BY scheduled_at DESC`, status,
	)
	if err != nil {
		return nil, fmt.Errorf("query maintenance by status: %w", err)
	}
	defer rows.Close()

	var tasks []*model.MaintenanceTask
	for rows.Next() {
		var m model.MaintenanceTask
		var startedAt, completedAt sql.NullTime
		err := rows.Scan(
			&m.ID, &m.SiloID, &m.Type, &m.Description, &m.Status, &m.ScheduledAt,
			&startedAt, &completedAt, &m.Technician, &m.Cost, &m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan maintenance: %w", err)
		}
		if startedAt.Valid {
			m.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			m.CompletedAt = &completedAt.Time
		}
		tasks = append(tasks, &m)
	}
	return tasks, nil
}

// Update 更新维护任务
func (s *MaintenanceStore) Update(m *model.MaintenanceTask) error {
	m.UpdatedAt = now()
	_, err := s.db.conn.Exec(
		`UPDATE maintenance_tasks SET silo_id=?, type=?, description=?, status=?, scheduled_at=?, started_at=?, completed_at=?, technician=?, cost=?, updated_at=? WHERE id=?`,
		m.SiloID, m.Type, m.Description, m.Status, m.ScheduledAt, m.StartedAt, m.CompletedAt,
		m.Technician, m.Cost, m.UpdatedAt, m.ID,
	)
	if err != nil {
		return fmt.Errorf("update maintenance: %w", err)
	}
	return nil
}

// StartMaintenance 开始维护
func (s *MaintenanceStore) StartMaintenance(id, technician string) error {
	_, err := s.db.conn.Exec(
		`UPDATE maintenance_tasks SET status='in_progress', started_at=?, technician=?, updated_at=? WHERE id=?`,
		now(), technician, now(), id,
	)
	if err != nil {
		return fmt.Errorf("start maintenance: %w", err)
	}
	return nil
}

// CompleteMaintenance 完成维护
func (s *MaintenanceStore) CompleteMaintenance(id string, cost float64) error {
	_, err := s.db.conn.Exec(
		`UPDATE maintenance_tasks SET status='completed', completed_at=?, cost=?, updated_at=? WHERE id=?`,
		now(), cost, now(), id,
	)
	if err != nil {
		return fmt.Errorf("complete maintenance: %w", err)
	}
	return nil
}

// SumCost 统计维护总费用
func (s *MaintenanceStore) SumCost() (float64, error) {
	row := s.db.conn.QueryRow(
		`SELECT COALESCE(SUM(cost),0) FROM maintenance_tasks WHERE status='completed'`,
	)
	var total float64
	err := row.Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("sum cost: %w", err)
	}
	return total, nil
}

// CountByStatus 按状态统计
func (s *MaintenanceStore) CountByStatus() (scheduled, inProgress, completed int, err error) {
	row := s.db.conn.QueryRow(
		`SELECT
			COALESCE(SUM(CASE WHEN status='scheduled' THEN 1 ELSE 0 END),0),
			COALESCE(SUM(CASE WHEN status='in_progress' THEN 1 ELSE 0 END),0),
			COALESCE(SUM(CASE WHEN status='completed' THEN 1 ELSE 0 END),0)
		 FROM maintenance_tasks`,
	)
	err = row.Scan(&scheduled, &inProgress, &completed)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("count by status: %w", err)
	}
	return
}
