package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"grain-silo/internal/model"
	"grain-silo/internal/service"
)

// InspectionHandler 质检处理器
type InspectionHandler struct {
	svc *service.InspectionService
}

func NewInspectionHandler(svc *service.InspectionService) *InspectionHandler {
	return &InspectionHandler{svc: svc}
}

func (h *InspectionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SiloID       string           `json:"silo_id"`
		GrainType    model.GrainType  `json:"grain_type"`
		Moisture     float64          `json:"moisture"`
		Impurity     float64          `json:"impurity"`
		PestDetected bool             `json:"pest_detected"`
		Inspector    string           `json:"inspector"`
		Result       string           `json:"result"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	record, err := h.svc.CreateInspection(req.SiloID, req.GrainType, req.Moisture, req.Impurity, req.PestDetected, req.Inspector, req.Result)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, record)
}

func (h *InspectionHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	record, err := h.svc.GetInspection(id)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (h *InspectionHandler) List(w http.ResponseWriter, r *http.Request) {
	siloID := r.URL.Query().Get("silo_id")
	if siloID != "" {
		records, err := h.svc.ListInspectionsBySilo(siloID)
		if err != nil {
			writeDomainError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, records)
		return
	}
	result := r.URL.Query().Get("result")
	if result != "" {
		records, err := h.svc.ListInspectionsByResult(result)
		if err != nil {
			writeDomainError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, records)
		return
	}
	records, err := h.svc.ListAllInspections()
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, records)
}

func (h *InspectionHandler) ListPest(w http.ResponseWriter, r *http.Request) {
	records, err := h.svc.ListPestDetected()
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, records)
}

// parseTimeRFC3339 解析 RFC3339 时间
func parseTimeRFC3339(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}
