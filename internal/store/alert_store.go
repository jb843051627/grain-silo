package store

import (
	"database/sql"
	"fmt"

	"grain-silo/internal/model"
)

// AlertStore 告警存储
type AlertStore struct {
	db *DB
}

// NewAlertStore 创建告警存储
func NewAlertStore(db *DB) *AlertStore {
	return &AlertStore{db: db}
}

// Create 创建告警
func (s *AlertStore) Create(a *model.Alert) error {
	_, err := s.db.conn.Exec(
		`INSERT INTO alerts (id, silo_id, level, message, status, acked_by, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		a.ID, a.SiloID, a.Level, a.Message, a.Status, a.AckedBy, a.CreatedAt, a.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert alert: %w", err)
	}
	return nil
}

// GetByID 根据ID查询告警
func (s *AlertStore) GetByID(id string) (*model.Alert, error) {
	row := s.db.conn.QueryRow(
		`SELECT id, silo_id, level, message, status, acked_by, created_at, updated_at FROM alerts WHERE id = ?`, id,
	)
	var a model.Alert
	err := row.Scan(&a.ID, &a.SiloID, &a.Level, &a.Message, &a.Status, &a.AckedBy, &a.CreatedAt, &a.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, model.ErrAlertNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query alert: %w", err)
	}
	return &a, nil
}

// ListBySilo 按筒仓查询告警
func (s *AlertStore) ListBySilo(siloID string) ([]*model.Alert, error) {
	rows, err := s.db.conn.Query(
		`SELECT id, silo_id, level, message, status, acked_by, created_at, updated_at FROM alerts WHERE silo_id = ? ORDER BY created_at DESC`,
		siloID,
	)
	if err != nil {
		return nil, fmt.Errorf("query alerts by silo: %w", err)
	}
	defer rows.Close()

	var alerts []*model.Alert
	for rows.Next() {
		var a model.Alert
		err := rows.Scan(&a.ID, &a.SiloID, &a.Level, &a.Message, &a.Status, &a.AckedBy, &a.CreatedAt, &a.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan alert: %w", err)
		}
		alerts = append(alerts, &a)
	}
	return alerts, nil
}

// ListByStatus 按状态查询告警
func (s *AlertStore) ListByStatus(status model.AlertStatus) ([]*model.Alert, error) {
	rows, err := s.db.conn.Query(
		`SELECT id, silo_id, level, message, status, acked_by, created_at, updated_at FROM alerts WHERE status = ? ORDER BY created_at DESC`,
		status,
	)
	if err != nil {
		return nil, fmt.Errorf("query alerts by status: %w", err)
	}
	defer rows.Close()

	var alerts []*model.Alert
	for rows.Next() {
		var a model.Alert
		err := rows.Scan(&a.ID, &a.SiloID, &a.Level, &a.Message, &a.Status, &a.AckedBy, &a.CreatedAt, &a.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan alert: %w", err)
		}
		alerts = append(alerts, &a)
	}
	return alerts, nil
}

// Ack 确认告警
func (s *AlertStore) Ack(id, ackedBy string) error {
	_, err := s.db.conn.Exec(
		`UPDATE alerts SET status='acked', acked_by=?, updated_at=? WHERE id=?`,
		ackedBy, now(), id,
	)
	if err != nil {
		return fmt.Errorf("ack alert: %w", err)
	}
	return nil
}

// Resolve 解决告警
func (s *AlertStore) Resolve(id string) error {
	_, err := s.db.conn.Exec(
		`UPDATE alerts SET status='resolved', updated_at=? WHERE id=?`,
		now(), id,
	)
	if err != nil {
		return fmt.Errorf("resolve alert: %w", err)
	}
	return nil
}

// CountByLevel 按级别统计
func (s *AlertStore) CountByLevel() (info, warning, critical int, err error) {
	row := s.db.conn.QueryRow(
		`SELECT
			COALESCE(SUM(CASE WHEN level='info' THEN 1 ELSE 0 END),0),
			COALESCE(SUM(CASE WHEN level='warning' THEN 1 ELSE 0 END),0),
			COALESCE(SUM(CASE WHEN level='critical' THEN 1 ELSE 0 END),0)
		 FROM alerts WHERE status='active'`,
	)
	err = row.Scan(&info, &warning, &critical)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("count by level: %w", err)
	}
	return
}

// CountActive 统计活跃告警数
func (s *AlertStore) CountActive(siloID string) (int, error) {
	row := s.db.conn.QueryRow(
		`SELECT COUNT(*) FROM alerts WHERE silo_id = ? AND status = 'active'`, siloID,
	)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count active: %w", err)
	}
	return count, nil
}

// ListActive 查询活跃告警
func (s *AlertStore) ListActive() ([]*model.Alert, error) {
	rows, err := s.db.conn.Query(
		`SELECT id, silo_id, level, message, status, acked_by, created_at, updated_at FROM alerts WHERE status = 'active' ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("query active alerts: %w", err)
	}
	defer rows.Close()

	var alerts []*model.Alert
	for rows.Next() {
		var a model.Alert
		err := rows.Scan(&a.ID, &a.SiloID, &a.Level, &a.Message, &a.Status, &a.AckedBy, &a.CreatedAt, &a.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan alert: %w", err)
		}
		alerts = append(alerts, &a)
	}
	return alerts, nil
}
