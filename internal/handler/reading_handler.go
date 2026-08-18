package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"grain-silo/internal/model"
	"grain-silo/internal/service"
)

// ReadingHandler 读数处理器
type ReadingHandler struct {
	svc *service.ReadingService
}

func NewReadingHandler(svc *service.ReadingService) *ReadingHandler {
	return &ReadingHandler{svc: svc}
}

func (h *ReadingHandler) Create(w http.ResponseWriter, r *http.Request) {
	siloID := r.PathValue("id")
	var req struct {
		Temp     float64 `json:"temp"`
		Humidity float64 `json:"humidity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	reading, err := h.svc.RecordReading(siloID, req.Temp, req.Humidity)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, reading)
}

func (h *ReadingHandler) GetLatest(w http.ResponseWriter, r *http.Request) {
	siloID := r.PathValue("id")
	reading, err := h.svc.GetLatestReading(siloID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, reading)
}

func (h *ReadingHandler) List(w http.ResponseWriter, r *http.Request) {
	siloID := r.PathValue("id")
	limit := parseInt(r.URL.Query().Get("limit"), 100)
	readings, err := h.svc.ListReadingsBySilo(siloID, limit)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, readings)
}

func (h *ReadingHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	siloID := r.PathValue("id")
	metrics, err := h.svc.GetSiloMetrics(siloID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, metrics)
}

func (h *ReadingHandler) Cleanup(w http.ResponseWriter, r *http.Request) {
	days := parseInt(r.URL.Query().Get("days"), 90)
	deleted, err := h.svc.CleanupOldReadings(days)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int64{"deleted": deleted})
}

func (h *ReadingHandler) BatchCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Readings []*model.Reading `json:"readings"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	count, err := h.svc.BatchIngestReadings(r.Context(), req.Readings)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"count": count})
}

// AlertHandler 告警处理器
type AlertHandler struct {
	svc *service.AlertService
}

func NewAlertHandler(svc *service.AlertService) *AlertHandler {
	return &AlertHandler{svc: svc}
}

func (h *AlertHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SiloID  string           `json:"silo_id"`
		Level   model.AlertLevel `json:"level"`
		Message string           `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	alert, err := h.svc.CreateAlert(req.SiloID, req.Level, req.Message)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, alert)
}

func (h *AlertHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	alert, err := h.svc.GetAlert(id)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, alert)
}

func (h *AlertHandler) List(w http.ResponseWriter, r *http.Request) {
	siloID := r.URL.Query().Get("silo_id")
	if siloID != "" {
		alerts, err := h.svc.ListAlertsBySilo(siloID)
		if err != nil {
			writeDomainError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, alerts)
		return
	}
	alerts, err := h.svc.ListActiveAlerts()
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, alerts)
}

func (h *AlertHandler) Ack(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		AckedBy string `json:"acked_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	alert, err := h.svc.AckAlert(id, req.AckedBy)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, alert)
}

func (h *AlertHandler) Resolve(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Operator string `json:"operator"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	alert, err := h.svc.ResolveAlert(id, req.Operator)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, alert)
}

func (h *AlertHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.svc.GetAlertSummary()
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

// MaintenanceHandler 维护处理器
type MaintenanceHandler struct {
	svc *service.MaintenanceService
}

func NewMaintenanceHandler(svc *service.MaintenanceService) *MaintenanceHandler {
	return &MaintenanceHandler{svc: svc}
}

func (h *MaintenanceHandler) Schedule(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SiloID      string  `json:"silo_id"`
		Type        string  `json:"type"`
		Description string  `json:"description"`
		Technician  string  `json:"technician"`
		ScheduledAt string  `json:"scheduled_at"`
		Cost        float64 `json:"cost"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	scheduledAt, err := parseTime(req.ScheduledAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid scheduled_at")
		return
	}
	task, err := h.svc.ScheduleMaintenance(req.SiloID, req.Type, req.Description, req.Technician, scheduledAt, req.Cost)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, task)
}

func (h *MaintenanceHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	task, err := h.svc.GetMaintenance(id)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (h *MaintenanceHandler) List(w http.ResponseWriter, r *http.Request) {
	siloID := r.URL.Query().Get("silo_id")
	if siloID != "" {
		tasks, err := h.svc.ListMaintenanceBySilo(siloID)
		if err != nil {
			writeDomainError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, tasks)
		return
	}
	status := r.URL.Query().Get("status")
	if status != "" {
		tasks, err := h.svc.ListMaintenanceByStatus(model.MaintenanceStatus(status))
		if err != nil {
			writeDomainError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, tasks)
		return
	}
	tasks, err := h.svc.ListMaintenanceByStatus(model.MAINT_SCHEDULED)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

func (h *MaintenanceHandler) Start(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Technician string `json:"technician"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	task, err := h.svc.StartMaintenance(id, req.Technician)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (h *MaintenanceHandler) Complete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Operator string  `json:"operator"`
		Cost     float64 `json:"cost"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	task, err := h.svc.CompleteMaintenance(id, req.Operator, req.Cost)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (h *MaintenanceHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Operator string `json:"operator"`
		Reason   string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	task, err := h.svc.CancelMaintenance(id, req.Operator, req.Reason)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (h *MaintenanceHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.svc.GetMaintenanceSummary()
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

// parseTime 解析时间
func parseTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}
