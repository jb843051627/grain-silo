package store

import (
	"database/sql"
	"fmt"
	"time"

	"grain-silo/internal/model"
)

// TransferStore 调拨存储
type TransferStore struct {
	db *DB
}

// NewTransferStore 创建调拨存储
func NewTransferStore(db *DB) *TransferStore {
	return &TransferStore{db: db}
}

// Create 创建调拨记录
func (s *TransferStore) Create(t *model.TransferRecord) error {
	_, err := s.db.conn.Exec(
		`INSERT INTO transfers (id, from_silo_id, to_silo_id, grain_type, quantity, status, scheduled_at, completed_at, operator, remark, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		t.ID, t.FromSiloID, t.ToSiloID, t.GrainType, t.Quantity, t.Status,
		t.ScheduledAt, t.CompletedAt, t.Operator, t.Remark, t.CreatedAt, t.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert transfer: %w", err)
	}
	return nil
}

// GetByID 根据ID查询调拨记录
func (s *TransferStore) GetByID(id string) (*model.TransferRecord, error) {
	row := s.db.conn.QueryRow(
		`SELECT id, from_silo_id, to_silo_id, grain_type, quantity, status, scheduled_at, completed_at, operator, remark, created_at, updated_at
		 FROM transfers WHERE id = ?`, id,
	)
	var t model.TransferRecord
	var completedAt sql.NullTime
	err := row.Scan(
		&t.ID, &t.FromSiloID, &t.ToSiloID, &t.GrainType, &t.Quantity, &t.Status,
		&t.ScheduledAt, &completedAt, &t.Operator, &t.Remark, &t.CreatedAt, &t.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query transfer: %w", err)
	}
	if completedAt.Valid {
		t.CompletedAt = &completedAt.Time
	}
	return &t, nil
}

// ListByStatus 按状态查询调拨列表
func (s *TransferStore) ListByStatus(status model.TransferStatus) ([]*model.TransferRecord, error) {
	rows, err := s.db.conn.Query(
		`SELECT id, from_silo_id, to_silo_id, grain_type, quantity, status, scheduled_at, completed_at, operator, remark, created_at, updated_at
		 FROM transfers WHERE status = ? ORDER BY created_at DESC`, status,
	)
	if err != nil {
		return nil, fmt.Errorf("query transfers: %w", err)
	}
	defer rows.Close()

	var transfers []*model.TransferRecord
	for rows.Next() {
		var t model.TransferRecord
		var completedAt sql.NullTime
		err := rows.Scan(
			&t.ID, &t.FromSiloID, &t.ToSiloID, &t.GrainType, &t.Quantity, &t.Status,
			&t.ScheduledAt, &completedAt, &t.Operator, &t.Remark, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan transfer: %w", err)
		}
		if completedAt.Valid {
			t.CompletedAt = &completedAt.Time
		}
		transfers = append(transfers, &t)
	}
	return transfers, nil
}

// ListBySilo 按筒仓查询调拨列表
func (s *TransferStore) ListBySilo(siloID string) ([]*model.TransferRecord, error) {
	rows, err := s.db.conn.Query(
		`SELECT id, from_silo_id, to_silo_id, grain_type, quantity, status, scheduled_at, completed_at, operator, remark, created_at, updated_at
		 FROM transfers WHERE from_silo_id = ? OR to_silo_id = ? ORDER BY created_at DESC`, siloID, siloID,
	)
	if err != nil {
		return nil, fmt.Errorf("query transfers by silo: %w", err)
	}
	defer rows.Close()

	var transfers []*model.TransferRecord
	for rows.Next() {
		var t model.TransferRecord
		var completedAt sql.NullTime
		err := rows.Scan(
			&t.ID, &t.FromSiloID, &t.ToSiloID, &t.GrainType, &t.Quantity, &t.Status,
			&t.ScheduledAt, &completedAt, &t.Operator, &t.Remark, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan transfer: %w", err)
		}
		if completedAt.Valid {
			t.CompletedAt = &completedAt.Time
		}
		transfers = append(transfers, &t)
	}
	return transfers, nil
}

// ListByDateRange 按日期范围查询
func (s *TransferStore) ListByDateRange(start, end time.Time) ([]*model.TransferRecord, error) {
	rows, err := s.db.conn.Query(
		`SELECT id, from_silo_id, to_silo_id, grain_type, quantity, status, scheduled_at, completed_at, operator, remark, created_at, updated_at
		 FROM transfers WHERE scheduled_at >= ? AND scheduled_at <= ? ORDER BY scheduled_at DESC`, start, end,
	)
	if err != nil {
		return nil, fmt.Errorf("query transfers by date: %w", err)
	}
	defer rows.Close()

	var transfers []*model.TransferRecord
	for rows.Next() {
		var t model.TransferRecord
		var completedAt sql.NullTime
		err := rows.Scan(
			&t.ID, &t.FromSiloID, &t.ToSiloID, &t.GrainType, &t.Quantity, &t.Status,
			&t.ScheduledAt, &completedAt, &t.Operator, &t.Remark, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan transfer: %w", err)
		}
		if completedAt.Valid {
			t.CompletedAt = &completedAt.Time
		}
		transfers = append(transfers, &t)
	}
	return transfers, nil
}

// Update 更新调拨记录
func (s *TransferStore) Update(t *model.TransferRecord) error {
	t.UpdatedAt = now()
	_, err := s.db.conn.Exec(
		`UPDATE transfers SET from_silo_id=?, to_silo_id=?, grain_type=?, quantity=?, status=?, scheduled_at=?, completed_at=?, operator=?, remark=?, updated_at=? WHERE id=?`,
		t.FromSiloID, t.ToSiloID, t.GrainType, t.Quantity, t.Status,
		t.ScheduledAt, t.CompletedAt, t.Operator, t.Remark, t.UpdatedAt, t.ID,
	)
	if err != nil {
		return fmt.Errorf("update transfer: %w", err)
	}
	return nil
}

// UpdateStatus 更新状态
func (s *TransferStore) UpdateStatus(id string, status model.TransferStatus) error {
	_, err := s.db.conn.Exec(
		`UPDATE transfers SET status=?, updated_at=? WHERE id=?`,
		status, now(), id,
	)
	if err != nil {
		return fmt.Errorf("update transfer status: %w", err)
	}
	return nil
}

// CompleteTransfer 完成调拨
func (s *TransferStore) CompleteTransfer(id string) error {
	now := time.Now().UTC()
	_, err := s.db.conn.Exec(
		`UPDATE transfers SET status='completed', completed_at=?, updated_at=? WHERE id=?`,
		now, now, id,
	)
	if err != nil {
		return fmt.Errorf("complete transfer: %w", err)
	}
	return nil
}

// ListAll 查询全部调拨
func (s *TransferStore) ListAll() ([]*model.TransferRecord, error) {
	rows, err := s.db.conn.Query(
		`SELECT id, from_silo_id, to_silo_id, grain_type, quantity, status, scheduled_at, completed_at, operator, remark, created_at, updated_at
		 FROM transfers ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("query all transfers: %w", err)
	}
	defer rows.Close()

	var transfers []*model.TransferRecord
	for rows.Next() {
		var t model.TransferRecord
		var completedAt sql.NullTime
		err := rows.Scan(
			&t.ID, &t.FromSiloID, &t.ToSiloID, &t.GrainType, &t.Quantity, &t.Status,
			&t.ScheduledAt, &completedAt, &t.Operator, &t.Remark, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan transfer: %w", err)
		}
		if completedAt.Valid {
			t.CompletedAt = &completedAt.Time
		}
		transfers = append(transfers, &t)
	}
	return transfers, nil
}

// SumByGrainType 按谷物种类汇总数量
func (s *TransferStore) SumByGrainType() (map[model.GrainType]float64, error) {
	rows, err := s.db.conn.Query(
		`SELECT grain_type, COALESCE(SUM(quantity),0) FROM transfers WHERE status='completed' GROUP BY grain_type`,
	)
	if err != nil {
		return nil, fmt.Errorf("query sum by grain type: %w", err)
	}
	defer rows.Close()

	result := make(map[model.GrainType]float64)
	for rows.Next() {
		var gt string
		var qty float64
		if err := rows.Scan(&gt, &qty); err != nil {
			return nil, fmt.Errorf("scan grain type: %w", err)
		}
		result[model.GrainType(gt)] = qty
	}
	return result, nil
}
