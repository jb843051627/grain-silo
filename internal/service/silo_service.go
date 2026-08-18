package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"grain-silo/internal/model"
	"grain-silo/internal/store"
)

// SiloService 筒仓服务
type SiloService struct {
	siloStore   *store.SiloStore
	readingStore *store.ReadingStore
	alertStore  *store.AlertStore
	auditStore  *store.AuditStore
}

// NewSiloService 创建筒仓服务
func NewSiloService(siloStore *store.SiloStore, readingStore *store.ReadingStore, alertStore *store.AlertStore, auditStore *store.AuditStore) *SiloService {
	return &SiloService{
		siloStore:    siloStore,
		readingStore: readingStore,
		alertStore:   alertStore,
		auditStore:   auditStore,
	}
}

// CreateSilo 创建筒仓
func (s *SiloService) CreateSilo(name, location string, capacity float64, grainType model.GrainType) (*model.Silo, error) {
	if name == "" {
		return nil, model.ErrValidationError
	}
	if capacity <= 0 {
		return nil, model.ErrValidationError
	}
	now := time.Now().UTC()
	silo := &model.Silo{
		ID:        uuid.NewString(),
		Name:      name,
		Capacity:  capacity,
		GrainType: grainType,
		Status:    model.SILO_EMPTY,
		Location:  location,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.siloStore.Create(silo); err != nil {
		return nil, fmt.Errorf("create silo: %w", err)
	}
	s.logAudit("create", "silo", silo.ID, "system", fmt.Sprintf("created silo %s", name))
	return silo, nil
}

// GetSilo 获取筒仓详情
func (s *SiloService) GetSilo(id string) (*model.Silo, error) {
	silo, err := s.siloStore.GetByID(id)
	if err != nil {
		return nil, err
	}
	if silo == nil {
		return nil, model.ErrSiloNotFound
	}
	return silo, nil
}

// ListSilos 列出筒仓
func (s *SiloService) ListSilos() ([]*model.Silo, error) {
	return s.siloStore.List()
}

// ListSilosByStatus 按状态列出筒仓
func (s *SiloService) ListSilosByStatus(status model.SiloStatus) ([]*model.Silo, error) {
	return s.siloStore.ListByStatus(status)
}

// UpdateSiloStatus 更新筒仓状态
func (s *SiloService) UpdateSiloStatus(id string, status model.SiloStatus) error {
	silo, err := s.siloStore.GetByID(id)
	if err != nil {
		return err
	}
	if silo.Status == model.SILO_LOCKED && status != model.SILO_MAINTENANCE {
		return model.ErrSilolocked
	}
	silo.Status = status
	if err := s.siloStore.Update(silo); err != nil {
		return err
	}
	s.logAudit("update_status", "silo", id, "operator", fmt.Sprintf("status -> %s", status))
	return nil
}

// FillSilo 装粮
func (s *SiloService) FillSilo(id string, quantity float64) (*model.Silo, error) {
	if quantity <= 0 {
		return nil, model.ErrInvalidQuantity
	}
	silo, err := s.siloStore.GetByID(id)
	if err != nil {
		return nil, err
	}
	if silo.Status == model.SILO_LOCKED {
		return nil, model.ErrSilolocked
	}
	if silo.Status == model.SILO_MAINTENANCE {
		return nil, model.ErrSiloNotReady
	}
	newLoad := silo.CurrentLoad + quantity
	if newLoad > silo.Capacity {
		return nil, model.ErrCapacityExceeded
	}
	silo.CurrentLoad = newLoad
	if silo.CurrentLoad == silo.Capacity {
		silo.Status = model.SILO_FULL
	} else {
		silo.Status = model.SILO_FILLING
	}
	if err := s.siloStore.Update(silo); err != nil {
		return nil, err
	}
	s.logAudit("fill", "silo", id, "operator", fmt.Sprintf("filled %.2f tons", quantity))
	return silo, nil
}

// DischargeSilo 出粮
func (s *SiloService) DischargeSilo(id string, quantity float64) (*model.Silo, error) {
	if quantity <= 0 {
		return nil, model.ErrInvalidQuantity
	}
	silo, err := s.siloStore.GetByID(id)
	if err != nil {
		return nil, err
	}
	if silo.Status == model.SILO_LOCKED {
		return nil, model.ErrSilolocked
	}
	if silo.CurrentLoad < quantity {
		return nil, model.ErrInvalidQuantity
	}
	silo.CurrentLoad = silo.CurrentLoad - quantity
	if silo.CurrentLoad == 0 {
		silo.Status = model.SILO_EMPTY
	} else {
		silo.Status = model.SILO_EMPTYING
	}
	if err := s.siloStore.Update(silo); err != nil {
		return nil, err
	}
	s.logAudit("discharge", "silo", id, "operator", fmt.Sprintf("discharged %.2f tons", quantity))
	return silo, nil
}

// GetSiloMetrics 获取筒仓指标
func (s *SiloService) GetSiloMetrics(id string) (*model.SiloMetrics, error) {
	silo, err := s.siloStore.GetByID(id)
	if err != nil {
		return nil, err
	}
	avgTemp, avgHumidity, err := s.readingStore.GetAvg24h(id)
	if err != nil {
		return nil, err
	}
	alertCount, err := s.alertStore.CountActive(id)
	if err != nil {
		return nil, err
	}
	utilization := 0.0
	if silo.Capacity > 0 {
		utilization = silo.CurrentLoad / silo.Capacity * 100
	}
	return &model.SiloMetrics{
		SiloID:         id,
		CurrentLoad:    silo.CurrentLoad,
		Utilization:    utilization,
		AvgTemp24h:     avgTemp,
		AvgHumidity24h: avgHumidity,
		AlertCount:     alertCount,
	}, nil
}

// GetSiloSummary 获取筒仓汇总
func (s *SiloService) GetSiloSummary() (*model.SiloSummary, error) {
	totalCap, totalLoad, count, err := s.siloStore.GetCapacity()
	if err != nil {
		return nil, err
	}
	byType, err := s.siloStore.CountByGrainType()
	if err != nil {
		return nil, err
	}
	utilization := 0.0
	if totalCap > 0 {
		utilization = totalLoad / totalCap * 100
	}
	return &model.SiloSummary{
		TotalSilos:    count,
		ActiveSilos:   count,
		TotalCapacity: totalCap,
		TotalLoad:     totalLoad,
		Utilization:   utilization,
		ByGrainType:   byType,
		GeneratedAt:   time.Now().UTC(),
	}, nil
}

// LockSilo 锁定筒仓
func (s *SiloService) LockSilo(id, operator string) error {
	silo, err := s.siloStore.GetByID(id)
	if err != nil {
		return err
	}
	silo.Status = model.SILO_LOCKED
	if err := s.siloStore.Update(silo); err != nil {
		return err
	}
	s.logAudit("lock", "silo", id, operator, "silo locked")
	return nil
}

// UnlockSilo 解锁筒仓
func (s *SiloService) UnlockSilo(id, operator string) error {
	silo, err := s.siloStore.GetByID(id)
	if err != nil {
		return err
	}
	if silo.Status != model.SILO_LOCKED {
		return model.ErrInvalidStatus
	}
	if silo.CurrentLoad == 0 {
		silo.Status = model.SILO_EMPTY
	} else if silo.CurrentLoad >= silo.Capacity {
		silo.Status = model.SILO_FULL
	} else {
		silo.Status = model.SILO_FILLING
	}
	if err := s.siloStore.Update(silo); err != nil {
		return err
	}
	s.logAudit("unlock", "silo", id, operator, "silo unlocked")
	return nil
}

// DeleteSilo 删除筒仓
func (s *SiloService) DeleteSilo(id string) error {
	silo, err := s.siloStore.GetByID(id)
	if err != nil {
		return err
	}
	if silo.CurrentLoad > 0 {
		return model.ErrValidationError
	}
	if err := s.siloStore.Delete(id); err != nil {
		return err
	}
	s.logAudit("delete", "silo", id, "operator", fmt.Sprintf("deleted silo %s", silo.Name))
	return nil
}

// logAudit 记录审计日志
func (s *SiloService) logAudit(action, entity, entityID, operator, detail string) {
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
