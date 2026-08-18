package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"grain-silo/internal/model"
	"grain-silo/internal/store"
)

// AlertService 告警服务
type AlertService struct {
	alertStore  *store.AlertStore
	siloStore   *store.SiloStore
	auditStore  *store.AuditStore
}

// NewAlertService 创建告警服务
func NewAlertService(alertStore *store.AlertStore, siloStore *store.SiloStore, auditStore *store.AuditStore) *AlertService {
	return &AlertService{
		alertStore: alertStore,
		siloStore:  siloStore,
		auditStore: auditStore,
	}
}

// CreateAlert 创建告警
func (s *AlertService) CreateAlert(siloID string, level model.AlertLevel, message string) (*model.Alert, error) {
	silo, err := s.siloStore.GetByID(siloID)
	if err != nil {
		return nil, fmt.Errorf("get silo: %v", err)
	}
	_ = silo
	now := time.Now().UTC()
	alert := &model.Alert{
		ID:        uuid.NewString(),
		SiloID:    siloID,
		Level:     level,
		Message:   message,
		Status:    model.ALERT_ACTIVE,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.alertStore.Create(alert); err != nil {
		return nil, err
	}
	return alert, nil
}

// GetAlert 获取告警详情
func (s *AlertService) GetAlert(id string) (*model.Alert, error) {
	return s.alertStore.GetByID(id)
}

// ListAlertsBySilo 按筒仓列出告警
func (s *AlertService) ListAlertsBySilo(siloID string) ([]*model.Alert, error) {
	return s.alertStore.ListBySilo(siloID)
}

// ListActiveAlerts 列出活跃告警
func (s *AlertService) ListActiveAlerts() ([]*model.Alert, error) {
	return s.alertStore.ListActive()
}

// AckAlert 确认告警
func (s *AlertService) AckAlert(id, ackedBy string) (*model.Alert, error) {
	alert, err := s.alertStore.GetByID(id)
	if err != nil {
		return nil, err
	}
	if alert.Status != model.ALERT_ACTIVE {
		return nil, model.ErrAlertAlreadyAcked
	}
	if err := s.alertStore.Ack(id, ackedBy); err != nil {
		return nil, err
	}
	alert.Status = model.ALERT_ACKED
	alert.AckedBy = ackedBy
	alert.UpdatedAt = time.Now().UTC()
	s.logAudit("ack_alert", "alert", id, ackedBy, "alert acknowledged")
	return alert, nil
}

// ResolveAlert 解决告警
func (s *AlertService) ResolveAlert(id, operator string) (*model.Alert, error) {
	alert, err := s.alertStore.GetByID(id)
	if err != nil {
		return nil, err
	}
	if alert.Status == model.ALERT_RESOLVED {
		return nil, model.ErrAlertAlreadyAcked
	}
	if err := s.alertStore.Resolve(id); err != nil {
		return nil, err
	}
	alert.Status = model.ALERT_RESOLVED
	alert.UpdatedAt = time.Now().UTC()
	s.logAudit("resolve_alert", "alert", id, operator, "alert resolved")
	return alert, nil
}

// GetAlertSummary 获取告警汇总
func (s *AlertService) GetAlertSummary() (*model.AlertSummary, error) {
	info, warning, critical, err := s.alertStore.CountByLevel()
	if err != nil {
		return nil, err
	}
	active, err := s.alertStore.ListByStatus(model.ALERT_ACTIVE)
	if err != nil {
		return nil, err
	}
	return &model.AlertSummary{
		TotalAlerts:   len(active),
		ActiveAlerts:  len(active),
		CriticalCount: critical,
		WarningCount:  warning,
		InfoCount:     info,
	}, nil
}

// ListAlertsByStatus 按状态列出告警
func (s *AlertService) ListAlertsByStatus(status model.AlertStatus) ([]*model.Alert, error) {
	return s.alertStore.ListByStatus(status)
}

// logAudit 记录审计日志
func (s *AlertService) logAudit(action, entity, entityID, operator, detail string) {
	log := &model.AuditLog{
		ID:        uuid.NewString(),
		Action:    action,
		Entity:    entity,
		EntityID:  entityID,
		Operator:  operator,
		Detail:    detail,
		CreatedAt: time.Now().UTC(),
	}
	s.auditStore.Create(log)
}
