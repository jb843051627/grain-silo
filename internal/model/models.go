package model

import "time"

// GrainType 谷物种类
type GrainType string

const (
	GRAIN_WHEAT     GrainType = "wheat"
	GRAIN_RICE      GrainType = "rice"
	GRAIN_CORN      GrainType = "corn"
	GRAIN_SOYBEAN   GrainType = "soybean"
	GRAIN_BARLEY    GrainType = "barley"
)

// SiloStatus 筒仓状态
type SiloStatus string

const (
	SILO_EMPTY    SiloStatus = "empty"
	SILO_FILLING  SiloStatus = "filling"
	SILO_FULL     SiloStatus = "full"
	SILO_EMPTYING SiloStatus = "emptying"
	SILO_MAINTENANCE SiloStatus = "maintenance"
	SILO_LOCKED   SiloStatus = "locked"
)

// TransferStatus 调拨状态
type TransferStatus string

const (
	TRANSFER_PENDING   TransferStatus = "pending"
	TRANSFER_APPROVED  TransferStatus = "approved"
	TRANSFER_IN_TRANSIT TransferStatus = "in_transit"
	TRANSFER_COMPLETED TransferStatus = "completed"
	TRANSFER_CANCELLED TransferStatus = "cancelled"
	TRANSFER_REJECTED  TransferStatus = "rejected"
)

// AlertLevel 告警级别
type AlertLevel string

const (
	ALERT_INFO     AlertLevel = "info"
	ALERT_WARNING  AlertLevel = "warning"
	ALERT_CRITICAL AlertLevel = "critical"
)

// AlertStatus 告警状态
type AlertStatus string

const (
	ALERT_ACTIVE  AlertStatus = "active"
	ALERT_ACKED   AlertStatus = "acked"
	ALERT_RESOLVED AlertStatus = "resolved"
)

// MaintenanceStatus 维护状态
type MaintenanceStatus string

const (
	MAINT_SCHEDULED  MaintenanceStatus = "scheduled"
	MAINT_IN_PROGRESS MaintenanceStatus = "in_progress"
	MAINT_COMPLETED   MaintenanceStatus = "completed"
	MAINT_CANCELLED   MaintenanceStatus = "cancelled"
)

// Silo 筒仓
type Silo struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Capacity  float64     `json:"capacity"`  // 吨
	CurrentLoad float64   `json:"current_load"` // 吨
	GrainType GrainType   `json:"grain_type"`
	Status    SiloStatus  `json:"status"`
	Location  string      `json:"location"`
	TempSensor string     `json:"temp_sensor"`
	HumidSensor string    `json:"humid_sensor"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// TransferRecord 调拨记录
type TransferRecord struct {
	ID            string          `json:"id"`
	FromSiloID    string          `json:"from_silo_id"`
	ToSiloID      string          `json:"to_silo_id"`
	GrainType     GrainType       `json:"grain_type"`
	Quantity      float64         `json:"quantity"`
	Status        TransferStatus  `json:"status"`
	ScheduledAt   time.Time       `json:"scheduled_at"`
	CompletedAt   *time.Time      `json:"completed_at,omitempty"`
	Operator      string          `json:"operator"`
	Remark        string          `json:"remark"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// Reading 传感器读数
type Reading struct {
	ID        string    `json:"id"`
	SiloID    string    `json:"silo_id"`
	Temp      float64   `json:"temp"`
	Humidity  float64   `json:"humidity"`
	RecordedAt time.Time `json:"recorded_at"`
}

// Alert 告警
type Alert struct {
	ID         string       `json:"id"`
	SiloID     string       `json:"silo_id"`
	Level      AlertLevel   `json:"level"`
	Message    string       `json:"message"`
	Status     AlertStatus  `json:"status"`
	AckedBy    string       `json:"acked_by"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`
}

// MaintenanceTask 维护任务
type MaintenanceTask struct {
	ID          string             `json:"id"`
	SiloID      string             `json:"silo_id"`
	Type        string             `json:"type"`
	Description string             `json:"description"`
	Status      MaintenanceStatus  `json:"status"`
	ScheduledAt time.Time          `json:"scheduled_at"`
	StartedAt   *time.Time         `json:"started_at,omitempty"`
	CompletedAt *time.Time         `json:"completed_at,omitempty"`
	Technician  string             `json:"technician"`
	Cost        float64            `json:"cost"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// InspectionRecord 质检记录
type InspectionRecord struct {
	ID           string    `json:"id"`
	SiloID       string    `json:"silo_id"`
	GrainType    GrainType `json:"grain_type"`
	Moisture     float64   `json:"moisture"`
	Impurity     float64   `json:"impurity"`
	PestDetected bool      `json:"pest_detected"`
	Inspector    string    `json:"inspector"`
	Result       string    `json:"result"`
	InspectedAt  time.Time `json:"inspected_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// AuditLog 审计日志
type AuditLog struct {
	ID        string    `json:"id"`
	Action    string    `json:"action"`
	Entity    string    `json:"entity"`
	EntityID  string    `json:"entity_id"`
	Operator  string    `json:"operator"`
	Detail    string    `json:"detail"`
	CreatedAt time.Time `json:"created_at"`
}
