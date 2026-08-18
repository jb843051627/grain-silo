package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"grain-silo/internal/model"
	"grain-silo/internal/service"
)

// TransferHandler 调拨处理器
type TransferHandler struct {
	svc *service.TransferService
}

func NewTransferHandler(svc *service.TransferService) *TransferHandler {
	return &TransferHandler{svc: svc}
}

func (h *TransferHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FromSiloID string           `json:"from_silo_id"`
		ToSiloID   string           `json:"to_silo_id"`
		GrainType  model.GrainType  `json:"grain_type"`
		Quantity   float64          `json:"quantity"`
		Operator   string           `json:"operator"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	transfer, err := h.svc.CreateTransfer(req.FromSiloID, req.ToSiloID, req.GrainType, req.Quantity, req.Operator)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, transfer)
}

func (h *TransferHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	transfer, err := h.svc.GetTransfer(id)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, transfer)
}

func (h *TransferHandler) List(w http.ResponseWriter, r *http.Request) {
	siloID := r.URL.Query().Get("silo_id")
	status := r.URL.Query().Get("status")
	if siloID != "" {
		transfers, err := h.svc.ListTransfersBySilo(siloID)
		if err != nil {
			writeDomainError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, transfers)
		return
	}
	if status != "" {
		transfers, err := h.svc.ListAllTransfers()
		if err != nil {
			writeDomainError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, transfers)
		return
	}
	transfers, err := h.svc.ListAllTransfers()
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, transfers)
}

func (h *TransferHandler) Approve(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Operator string `json:"operator"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	transfer, err := h.svc.ApproveTransfer(id, req.Operator)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, transfer)
}

func (h *TransferHandler) Execute(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Operator string `json:"operator"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	transfer, err := h.svc.ExecuteTransfer(id, req.Operator)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, transfer)
}

func (h *TransferHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Operator string `json:"operator"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	transfer, err := h.svc.CancelTransfer(id, req.Operator)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, transfer)
}

func (h *TransferHandler) Reject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Operator string `json:"operator"`
		Reason   string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	transfer, err := h.svc.RejectTransfer(id, req.Operator, req.Reason)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, transfer)
}

func (h *TransferHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.svc.GetTransferSummary()
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (h *TransferHandler) ListByDate(w http.ResponseWriter, r *http.Request) {
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid start time")
		return
	}
	end, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid end time")
		return
	}
	transfers, err := h.svc.ListTransfersByDateRange(start, end)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, transfers)
}
