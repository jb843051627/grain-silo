package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"grain-silo/internal/model"
	"grain-silo/internal/store"
)

// MaintenanceService 维护服务
type MaintenanceService struct {
	maintStore  *store.MaintenanceStore
	siloStore   *store.SiloStore
	auditStore  *store.AuditStore
}

// NewMaintenanceService 创建维护服务
func NewMaintenanceService(maintStore *store.MaintenanceStore, siloStore *store.SiloStore, auditStore *store.AuditStore) *MaintenanceService {
	return &MaintenanceService{
		maintStore: maintStore,
		siloStore:  siloStore,
		auditStore: auditStore,
	}
}

// ScheduleMaintenance 安排维护
func (s *MaintenanceService) ScheduleMaintenance(siloID, maintType, description, technician string, scheduledAt time.Time, cost float64) (*model.MaintenanceTask, error) {
	silo, err := s.siloStore.GetByID(siloID)
	if err != nil {
		return nil, err
	}
	_ = silo
	if maintType == "" {
		return nil, model.ErrValidationError
	}
	now := time.Now().UTC()
	task := &model.MaintenanceTask{
		ID:          uuid.NewString(),
		SiloID:      siloID,
		Type:        maintType,
		Description: description,
		Status:      model.MAINT_SCHEDULED,
		ScheduledAt: scheduledAt,
		Technician:  technician,
		Cost:        cost,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.maintStore.Create(task); err != nil {
		return nil, err
	}
	s.logAudit("schedule_maintenance", "maintenance", task.ID, "operator", fmt.Sprintf("scheduled %s for silo %s", maintType, siloID))
	return task, nil
}

// StartMaintenance 开始维护
func (s *MaintenanceService) StartMaintenance(id, technician string) (*model.MaintenanceTask, error) {
	task, err := s.maintStore.GetByID(id)
	if err != nil {
		return nil, err
	}
	if task.Status != model.MAINT_SCHEDULED {
		return nil, model.ErrMaintenanceInProgress
	}
	if err := s.maintStore.StartMaintenance(id, technician); err != nil {
		return nil, fmt.Errorf("start maintenance: %w", err)
	}
	task.Status = model.MAINT_IN_PROGRESS
	task.Technician = technician
	task.UpdatedAt = time.Now().UTC()
	s.logAudit("start_maintenance", "maintenance", id, technician, "maintenance started")
	return task, nil
}

// CompleteMaintenance 完成维护
func (s *MaintenanceService) CompleteMaintenance(id, operator string, cost float64) (*model.MaintenanceTask, error) {
	task, err := s.maintStore.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("get maintenance: %w", err)
	}
	if task.Status != model.MAINT_IN_PROGRESS {
		return nil, model.ErrInvalidStatus
	}
	if err := s.maintStore.CompleteMaintenance(id, cost); err != nil {
		return nil, fmt.Errorf("complete maintenance: %w", err)
	}
	task.Status = model.MAINT_COMPLETED
	task.Cost = cost
	task.UpdatedAt = time.Now().UTC()
	s.logAudit("complete_maintenance", "maintenance", id, operator, fmt.Sprintf("completed, cost: %.2f", cost))
	return task, nil
}

// CancelMaintenance 取消维护
func (s *MaintenanceService) CancelMaintenance(id, operator, reason string) (*model.MaintenanceTask, error) {
	task, err := s.maintStore.GetByID(id)
	if err != nil {
		return nil, err
	}
	if task.Status == model.MAINT_COMPLETED {
		return nil, model.ErrInvalidStatus
	}
	task.Status = model.MAINT_CANCELLED
	task.Description = reason
	task.UpdatedAt = time.Now().UTC()
	if err := s.maintStore.Update(task); err != nil {
		return nil, err
	}
	s.logAudit("cancel_maintenance", "maintenance", id, operator, fmt.Sprintf("cancelled: %s", reason))
	return task, nil
}

// GetMaintenance 获取维护详情
func (s *MaintenanceService) GetMaintenance(id string) (*model.MaintenanceTask, error) {
	return s.maintStore.GetByID(id)
}

// ListMaintenanceBySilo 按筒仓列出维护
func (s *MaintenanceService) ListMaintenanceBySilo(siloID string) ([]*model.MaintenanceTask, error) {
	return s.maintStore.ListBySilo(siloID)
}

// ListMaintenanceByStatus 按状态列出维护
func (s *MaintenanceService) ListMaintenanceByStatus(status model.MaintenanceStatus) ([]*model.MaintenanceTask, error) {
	return s.maintStore.ListByStatus(status)
}

// GetMaintenanceSummary 获取维护汇总
func (s *MaintenanceService) GetMaintenanceSummary() (*model.MaintenanceSummary, error) {
	scheduled, inProgress, completed, err := s.maintStore.CountByStatus()
	if err != nil {
		return nil, err
	}
	totalCost, err := s.maintStore.SumCost()
	if err != nil {
		return nil, err
	}
	return &model.MaintenanceSummary{
		TotalTasks:     scheduled + inProgress + completed,
		ScheduledCount: scheduled,
		InProgressCount: inProgress,
		CompletedCount:  completed,
		TotalCost:      totalCost,
	}, nil
}

// logAudit 记录审计日志
func (s *MaintenanceService) logAudit(action, entity, entityID, operator, detail string) {
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
