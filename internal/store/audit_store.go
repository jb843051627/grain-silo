package store

import (
	"fmt"

	"grain-silo/internal/model"
)

// AuditStore 审计日志存储
type AuditStore struct {
	db *DB
}

// NewAuditStore 创建审计日志存储
func NewAuditStore(db *DB) *AuditStore {
	return &AuditStore{db: db}
}

// Create 创建审计日志
func (s *AuditStore) Create(log *model.AuditLog) error {
	_, err := s.db.conn.Exec(
		`INSERT INTO audit_logs (id, action, entity, entity_id, operator, detail, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		log.ID, log.Action, log.Entity, log.EntityID, log.Operator, log.Detail, log.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}
	return nil
}

// ListByEntity 按实体查询审计日志
func (s *AuditStore) ListByEntity(entity, entityID string) ([]*model.AuditLog, error) {
	rows, err := s.db.conn.Query(
		`SELECT id, action, entity, entity_id, operator, detail, created_at FROM audit_logs WHERE entity = ? AND entity_id = ? ORDER BY created_at DESC`,
		entity, entityID,
	)
	if err != nil {
		return nil, fmt.Errorf("query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []*model.AuditLog
	for rows.Next() {
		var l model.AuditLog
		err := rows.Scan(&l.ID, &l.Action, &l.Entity, &l.EntityID, &l.Operator, &l.Detail, &l.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan audit log: %w", err)
		}
		logs = append(logs, &l)
	}
	return logs, nil
}

// ListByOperator 按操作人查询
func (s *AuditStore) ListByOperator(operator string) ([]*model.AuditLog, error) {
	rows, err := s.db.conn.Query(
		`SELECT id, action, entity, entity_id, operator, detail, created_at FROM audit_logs WHERE operator = ? ORDER BY created_at DESC`,
		operator,
	)
	if err != nil {
		return nil, fmt.Errorf("query audit by operator: %w", err)
	}
	defer rows.Close()

	var logs []*model.AuditLog
	for rows.Next() {
		var l model.AuditLog
		err := rows.Scan(&l.ID, &l.Action, &l.Entity, &l.EntityID, &l.Operator, &l.Detail, &l.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan audit log: %w", err)
		}
		logs = append(logs, &l)
	}
	return logs, nil
}
