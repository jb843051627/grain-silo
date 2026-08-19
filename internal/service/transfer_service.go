package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"grain-silo/internal/model"
	"grain-silo/internal/store"
)

// TransferService 调拨服务
type TransferService struct {
	transferStore *store.TransferStore
	siloStore     *store.SiloStore
	auditStore    *store.AuditStore
}

// NewTransferService 创建调拨服务
func NewTransferService(transferStore *store.TransferStore, siloStore *store.SiloStore, auditStore *store.AuditStore) *TransferService {
	return &TransferService{
		transferStore: transferStore,
		siloStore:     siloStore,
		auditStore:    auditStore,
	}
}

// CreateTransfer 创建调拨
func (s *TransferService) CreateTransfer(fromID, toID string, grainType model.GrainType, quantity float64, operator string) (*model.TransferRecord, error) {
	if quantity <= 0 {
		return nil, model.ErrInvalidQuantity
	}
	if fromID == toID {
		return nil, model.ErrValidationError
	}
	fromSilo, err := s.siloStore.GetByID(fromID)
	if err != nil {
		return nil, err
	}
	toSilo, err := s.siloStore.GetByID(toID)
	if err != nil {
		return nil, err
	}
	if fromSilo.CurrentLoad < quantity {
		return nil, model.ErrInvalidQuantity
	}
	if toSilo.CurrentLoad+quantity > toSilo.Capacity {
		return nil, model.ErrCapacityExceeded
	}
	if fromSilo.Status == model.SILO_LOCKED || toSilo.Status == model.SILO_LOCKED {
		return nil, model.ErrSilolocked
	}
	now := time.Now().UTC()
	transfer := &model.TransferRecord{
		ID:          uuid.NewString(),
		FromSiloID:  fromID,
		ToSiloID:    toID,
		GrainType:   grainType,
		Quantity:    quantity,
		Status:      model.TRANSFER_PENDING,
		ScheduledAt: now,
		Operator:    operator,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.transferStore.Create(transfer); err != nil {
		return nil, fmt.Errorf("create transfer: %w", err)
	}
	s.logAudit("create_transfer", "transfer", transfer.ID, operator, fmt.Sprintf("from %s to %s, %.2f tons", fromID, toID, quantity))
	return transfer, nil
}

// ApproveTransfer 审批调拨
func (s *TransferService) ApproveTransfer(id, operator string) (*model.TransferRecord, error) {
	transfer, err := s.transferStore.GetByID(id)
	if err != nil {
		return nil, err
	}
	if transfer.Status != model.TRANSFER_PENDING {
		return nil, model.ErrTransferNotPending
	}
	transfer.Status = model.TRANSFER_APPROVED
	transfer.UpdatedAt = time.Now().UTC()
	if err := s.transferStore.Update(transfer); err != nil {
		return nil, err
	}
	s.logAudit("approve_transfer", "transfer", id, operator, "transfer approved")
	return transfer, nil
}

// ExecuteTransfer 执行调拨
func (s *TransferService) ExecuteTransfer(id, operator string) (*model.TransferRecord, error) {
	transfer, err := s.transferStore.GetByID(id)
	if err != nil {
		return nil, err
	}
	if transfer.Status != model.TRANSFER_APPROVED {
		return nil, model.ErrInvalidStatus
	}
	fromSilo, err := s.siloStore.GetByID(transfer.FromSiloID)
	if err != nil {
		return nil, err
	}
	toSilo, err := s.siloStore.GetByID(transfer.ToSiloID)
	if err != nil {
		return nil, err
	}
if fromSilo.CurrentLoad < transfer.Quantity {
		return nil, model.ErrInvalidQuantity
	}
	if toSilo.CurrentLoad+transfer.Quantity > toSilo.Capacity {
		return nil, model.ErrCapacityExceeded
	}
	if err := s.siloStore.TransferLoad(transfer.FromSiloID, transfer.ToSiloID, transfer.Quantity); err != nil {
		return nil, err
	}
	fromSilo.CurrentLoad = fromSilo.CurrentLoad - transfer.Quantity
	if fromSilo.CurrentLoad == 0 {
		fromSilo.Status = model.SILO_EMPTY
	} else {
		fromSilo.Status = model.SILO_EMPTYING
	}
	toSilo.CurrentLoad = toSilo.CurrentLoad + transfer.Quantity
	if toSilo.CurrentLoad == toSilo.Capacity {
		toSilo.Status = model.SILO_FULL
	} else {
		toSilo.Status = model.SILO_FILLING
	}
	if err := s.siloStore.Update(fromSilo); err != nil {
		return nil, err
	}
	if err := s.siloStore.Update(toSilo); err != nil {
		return nil, err
	}
	transfer.Status = model.TRANSFER_COMPLETED
	completedAt := time.Now().UTC()
	transfer.CompletedAt = &completedAt
	transfer.UpdatedAt = completedAt
	if err := s.transferStore.Update(transfer); err != nil {
		return nil, err
	}
	s.logAudit("execute_transfer", "transfer", id, operator, "transfer completed")
	return transfer, nil
}

// CancelTransfer 取消调拨
func (s *TransferService) CancelTransfer(id, operator string) (*model.TransferRecord, error) {
	transfer, err := s.transferStore.GetByID(id)
	if err != nil {
		return nil, err
	}
	if transfer.Status == model.TRANSFER_COMPLETED || transfer.Status == model.TRANSFER_IN_TRANSIT {
		return nil, model.ErrInvalidStatus
	}
	transfer.Status = model.TRANSFER_CANCELLED
	transfer.UpdatedAt = time.Now().UTC()
	if err := s.transferStore.Update(transfer); err != nil {
		return nil, err
	}
	s.logAudit("cancel_transfer", "transfer", id, operator, "transfer cancelled")
	return transfer, nil
}

// GetTransfer 获取调拨详情
func (s *TransferService) GetTransfer(id string) (*model.TransferRecord, error) {
	return s.transferStore.GetByID(id)
}

// ListPendingTransfers 列出待处理调拨
func (s *TransferService) ListPendingTransfers() ([]*model.TransferRecord, error) {
	return s.transferStore.ListByStatus(model.TRANSFER_PENDING)
}

// ListTransfersBySilo 按筒仓列出调拨
func (s *TransferService) ListTransfersBySilo(siloID string) ([]*model.TransferRecord, error) {
	return s.transferStore.ListBySilo(siloID)
}

// ListTransfersByDateRange 按日期范围列出调拨
func (s *TransferService) ListTransfersByDateRange(start, end time.Time) ([]*model.TransferRecord, error) {
	return s.transferStore.ListByDateRange(start, end)
}

// GetTransferSummary 获取调拨汇总
func (s *TransferService) GetTransferSummary() (*model.TransferSummary, error) {
	all, err := s.transferStore.ListAll()
	if err != nil {
		return nil, err
	}
	byType, err := s.transferStore.SumByGrainType()
	if err != nil {
		return nil, err
	}
	summary := &model.TransferSummary{
		ByGrainType: byType,
	}
	for _, t := range all {
		summary.TotalTransfers++
		summary.TotalQuantity = summary.TotalQuantity + t.Quantity
		switch t.Status {
		case model.TRANSFER_COMPLETED:
			summary.CompletedCount++
		case model.TRANSFER_PENDING:
			summary.PendingCount++
		}
	}
	return summary, nil
}

// ListAllTransfers 列出全部调拨
func (s *TransferService) ListAllTransfers() ([]*model.TransferRecord, error) {
	return s.transferStore.ListAll()
}

// RejectTransfer 拒绝调拨
func (s *TransferService) RejectTransfer(id, operator, reason string) (*model.TransferRecord, error) {
	transfer, err := s.transferStore.GetByID(id)
	if err != nil {
		return nil, err
	}
	if transfer.Status != model.TRANSFER_PENDING {
		return nil, model.ErrTransferNotPending
	}
	transfer.Status = model.TRANSFER_REJECTED
	transfer.Remark = reason
	transfer.UpdatedAt = time.Now().UTC()
	if err := s.transferStore.Update(transfer); err != nil {
		return nil, err
	}
	s.logAudit("reject_transfer", "transfer", id, operator, fmt.Sprintf("rejected: %s", reason))
	return transfer, nil
}

// logAudit 记录审计日志
func (s *TransferService) logAudit(action, entity, entityID, operator, detail string) {
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
