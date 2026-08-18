package model

import "errors"

// 领域错误定义
var (
	ErrSiloNotFound        = errors.New("silo not found")
	ErrSiloExists          = errors.New("silo already exists")
	ErrTransferNotFound    = errors.New("transfer not found")
	ErrReadingNotFound     = errors.New("reading not found")
	ErrAlertNotFound       = errors.New("alert not found")
	ErrMaintenanceNotFound = errors.New("maintenance task not found")
	ErrInspectionNotFound  = errors.New("inspection record not found")
	ErrCapacityExceeded    = errors.New("capacity exceeded")
	ErrInvalidStatus       = errors.New("invalid status transition")
	ErrInvalidQuantity     = errors.New("invalid quantity")
	ErrSilolocked          = errors.New("silo is locked")
	ErrSiloNotReady        = errors.New("silo not ready for operation")
	ErrTransferNotPending  = errors.New("transfer is not pending")
	ErrDuplicateTransfer   = errors.New("duplicate transfer request")
	ErrValidationError    = errors.New("validation error")
	ErrAlertAlreadyAcked   = errors.New("alert already acknowledged")
	ErrMaintenanceInProgress = errors.New("maintenance already in progress")
)
