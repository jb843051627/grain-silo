package model

import "time"

// SiloMetrics 筒仓指标
type SiloMetrics struct {
	SiloID         string  `json:"silo_id"`
	CurrentLoad    float64 `json:"current_load"`
	Utilization    float64 `json:"utilization"`
	AvgTemp24h     float64 `json:"avg_temp_24h"`
	AvgHumidity24h float64 `json:"avg_humidity_24h"`
	AlertCount     int     `json:"alert_count"`
}

// TransferSummary 调拨汇总
type TransferSummary struct {
	TotalTransfers   int     `json:"total_transfers"`
	CompletedCount   int     `json:"completed_count"`
	PendingCount     int     `json:"pending_count"`
	TotalQuantity     float64 `json:"total_quantity"`
	ByGrainType      map[GrainType]float64 `json:"by_grain_type"`
}

// SiloSummary 筒仓汇总
type SiloSummary struct {
	TotalSilos     int     `json:"total_silos"`
	ActiveSilos    int     `json:"active_silos"`
	TotalCapacity  float64 `json:"total_capacity"`
	TotalLoad      float64 `json:"total_load"`
	Utilization    float64 `json:"utilization"`
	ByGrainType    map[GrainType]float64 `json:"by_grain_type"`
	GeneratedAt    time.Time `json:"generated_at"`
}

// AlertSummary 告警汇总
type AlertSummary struct {
	TotalAlerts   int `json:"total_alerts"`
	ActiveAlerts  int `json:"active_alerts"`
	CriticalCount int `json:"critical_count"`
	WarningCount  int `json:"warning_count"`
	InfoCount     int `json:"info_count"`
}

// MaintenanceSummary 维护汇总
type MaintenanceSummary struct {
	TotalTasks    int `json:"total_tasks"`
	ScheduledCount int `json:"scheduled_count"`
	InProgressCount int `json:"in_progress_count"`
	CompletedCount int `json:"completed_count"`
	TotalCost     float64 `json:"total_cost"`
}
