package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"grain-silo/internal/model"
	"grain-silo/internal/store"
)

// InspectionService 质检服务
type InspectionService struct {
	inspStore  *store.InspectionStore
	siloStore  *store.SiloStore
	auditStore *store.AuditStore
}

// NewInspectionService 创建质检服务
func NewInspectionService(inspStore *store.InspectionStore, siloStore *store.SiloStore, auditStore *store.AuditStore) *InspectionService {
	return &InspectionService{
		inspStore:  inspStore,
		siloStore:  siloStore,
		auditStore: auditStore,
	}
}

// CreateInspection 创建质检记录
func (s *InspectionService) CreateInspection(siloID string, grainType model.GrainType, moisture, impurity float64, pestDetected bool, inspector, result string) (*model.InspectionRecord, error) {
	silo, err := s.siloStore.GetByID(siloID)
	if err != nil {
		return nil, err
	}
	_ = silo
	now := time.Now().UTC()
	record := &model.InspectionRecord{
		ID:           uuid.NewString(),
		SiloID:       siloID,
		GrainType:    grainType,
		Moisture:     moisture,
		Impurity:     impurity,
		PestDetected: pestDetected,
		Inspector:    inspector,
		Result:       result,
		InspectedAt:  now,
		CreatedAt:    now,
	}
	if err := s.inspStore.Create(record); err != nil {
		return nil, err
	}
	s.logAudit("create_inspection", "inspection", record.ID, inspector, fmt.Sprintf("result: %s", result))
	return record, nil
}

// GetInspection 获取质检详情
func (s *InspectionService) GetInspection(id string) (*model.InspectionRecord, error) {
	return s.inspStore.GetByID(id)
}

// ListInspectionsBySilo 按筒仓列出质检记录
func (s *InspectionService) ListInspectionsBySilo(siloID string) ([]*model.InspectionRecord, error) {
	return s.inspStore.ListBySilo(siloID)
}

// ListAllInspections 列出全部质检记录
func (s *InspectionService) ListAllInspections() ([]*model.InspectionRecord, error) {
	return s.inspStore.ListAll()
}

// ListInspectionsByResult 按结果列出质检记录
func (s *InspectionService) ListInspectionsByResult(result string) ([]*model.InspectionRecord, error) {
	return s.inspStore.ListByResult(result)
}

// ListPestDetected 列出检测到虫害的记录
func (s *InspectionService) ListPestDetected() ([]*model.InspectionRecord, error) {
	return s.inspStore.ListPestDetected()
}

// GetLatestInspection 获取最新质检记录
func (s *InspectionService) GetLatestInspection(siloID string) (*model.InspectionRecord, error) {
	return s.inspStore.GetLatest(siloID)
}

// logAudit 记录审计日志
func (s *InspectionService) logAudit(action, entity, entityID, operator, detail string) {
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
