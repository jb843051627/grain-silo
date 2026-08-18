package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"grain-silo/internal/model"
	"grain-silo/internal/service"
)

// SiloHandler 筒仓处理器
type SiloHandler struct {
	svc *service.SiloService
}

// NewSiloHandler 创建筒仓处理器
func NewSiloHandler(svc *service.SiloService) *SiloHandler {
	return &SiloHandler{svc: svc}
}

// Create 创建筒仓
func (h *SiloHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name      string             `json:"name"`
		Location  string             `json:"location"`
		Capacity  float64           `json:"capacity"`
		GrainType model.GrainType   `json:"grain_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	silo, err := h.svc.CreateSilo(req.Name, req.Location, req.Capacity, req.GrainType)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, silo)
}

// Get 获取筒仓详情
func (h *SiloHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	silo, err := h.svc.GetSilo(id)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, silo)
}

// List 列出筒仓
func (h *SiloHandler) List(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if status != "" {
		silos, err := h.svc.ListSilosByStatus(model.SiloStatus(status))
		if err != nil {
			writeDomainError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, silos)
		return
	}
	silos, err := h.svc.ListSilos()
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, silos)
}

// Fill 装粮
func (h *SiloHandler) Fill(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Quantity float64 `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	silo, err := h.svc.FillSilo(id, req.Quantity)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, silo)
}

// Discharge 出粮
func (h *SiloHandler) Discharge(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Quantity float64 `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	silo, err := h.svc.DischargeSilo(id, req.Quantity)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, silo)
}

// UpdateStatus 更新状态
func (h *SiloHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Status model.SiloStatus `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.UpdateSiloStatus(id, req.Status); err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// GetMetrics 获取指标
func (h *SiloHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	metrics, err := h.svc.GetSiloMetrics(id)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, metrics)
}

// GetSummary 获取汇总
func (h *SiloHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.svc.GetSiloSummary()
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

// Lock 锁定
func (h *SiloHandler) Lock(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	operator := r.URL.Query().Get("operator")
	if err := h.svc.LockSilo(id, operator); err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "locked"})
}

// Unlock 解锁
func (h *SiloHandler) Unlock(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	operator := r.URL.Query().Get("operator")
	if err := h.svc.UnlockSilo(id, operator); err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "unlocked"})
}

// Delete 删除
func (h *SiloHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.svc.DeleteSilo(id); err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// writeJSON 写 JSON 响应
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError 写错误响应
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// writeDomainError 写领域错误
func writeDomainError(w http.ResponseWriter, err error) {
	switch err {
	case model.ErrSiloNotFound, model.ErrTransferNotFound, model.ErrReadingNotFound,
		model.ErrAlertNotFound, model.ErrMaintenanceNotFound, model.ErrInspectionNotFound:
		writeError(w, http.StatusNotFound, err.Error())
	case model.ErrCapacityExceeded, model.ErrInvalidStatus, model.ErrInvalidQuantity,
		model.ErrSilolocked, model.ErrSiloNotReady, model.ErrTransferNotPending,
		model.ErrAlertAlreadyAcked, model.ErrMaintenanceInProgress, model.ErrValidationError:
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, err.Error())
	}
}

// parseInt 解析整数
func parseInt(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}
