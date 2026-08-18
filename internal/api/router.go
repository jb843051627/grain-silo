package api

import (
	"net/http"

	"grain-silo/internal/handler"
)

// Router HTTP 路由
type Router struct {
	mux *http.ServeMux
}

// NewRouter 创建路由
func NewRouter(
	siloHandler *handler.SiloHandler,
	transferHandler *handler.TransferHandler,
	readingHandler *handler.ReadingHandler,
	alertHandler *handler.AlertHandler,
	maintHandler *handler.MaintenanceHandler,
	inspHandler *handler.InspectionHandler,
) *Router {
	r := &Router{mux: http.NewServeMux()}

	// 筒仓路由
	r.mux.HandleFunc("GET /api/silos", siloHandler.List)
	r.mux.HandleFunc("POST /api/silos", siloHandler.Create)
	r.mux.HandleFunc("GET /api/silos/summary", siloHandler.GetSummary)
	r.mux.HandleFunc("GET /api/silos/{id}", siloHandler.Get)
	r.mux.HandleFunc("PUT /api/silos/{id}/status", siloHandler.UpdateStatus)
	r.mux.HandleFunc("POST /api/silos/{id}/fill", siloHandler.Fill)
	r.mux.HandleFunc("POST /api/silos/{id}/discharge", siloHandler.Discharge)
	r.mux.HandleFunc("GET /api/silos/{id}/metrics", siloHandler.GetMetrics)
	r.mux.HandleFunc("POST /api/silos/{id}/lock", siloHandler.Lock)
	r.mux.HandleFunc("POST /api/silos/{id}/unlock", siloHandler.Unlock)
	r.mux.HandleFunc("DELETE /api/silos/{id}", siloHandler.Delete)

	// 调拨路由
	r.mux.HandleFunc("GET /api/transfers", transferHandler.List)
	r.mux.HandleFunc("POST /api/transfers", transferHandler.Create)
	r.mux.HandleFunc("GET /api/transfers/summary", transferHandler.GetSummary)
	r.mux.HandleFunc("GET /api/transfers/{id}", transferHandler.Get)
	r.mux.HandleFunc("POST /api/transfers/{id}/approve", transferHandler.Approve)
	r.mux.HandleFunc("POST /api/transfers/{id}/execute", transferHandler.Execute)
	r.mux.HandleFunc("POST /api/transfers/{id}/cancel", transferHandler.Cancel)
	r.mux.HandleFunc("POST /api/transfers/{id}/reject", transferHandler.Reject)

	// 读数路由
	r.mux.HandleFunc("POST /api/silos/{id}/readings", readingHandler.Create)
	r.mux.HandleFunc("GET /api/silos/{id}/readings", readingHandler.List)
	r.mux.HandleFunc("GET /api/silos/{id}/readings/latest", readingHandler.GetLatest)
	r.mux.HandleFunc("POST /api/readings/batch", readingHandler.BatchCreate)
	r.mux.HandleFunc("DELETE /api/readings/cleanup", readingHandler.Cleanup)

	// 告警路由
	r.mux.HandleFunc("GET /api/alerts", alertHandler.List)
	r.mux.HandleFunc("POST /api/alerts", alertHandler.Create)
	r.mux.HandleFunc("GET /api/alerts/summary", alertHandler.GetSummary)
	r.mux.HandleFunc("GET /api/alerts/{id}", alertHandler.Get)
	r.mux.HandleFunc("POST /api/alerts/{id}/ack", alertHandler.Ack)
	r.mux.HandleFunc("POST /api/alerts/{id}/resolve", alertHandler.Resolve)

	// 维护路由
	r.mux.HandleFunc("GET /api/maintenance", maintHandler.List)
	r.mux.HandleFunc("POST /api/maintenance", maintHandler.Schedule)
	r.mux.HandleFunc("GET /api/maintenance/summary", maintHandler.GetSummary)
	r.mux.HandleFunc("GET /api/maintenance/{id}", maintHandler.Get)
	r.mux.HandleFunc("POST /api/maintenance/{id}/start", maintHandler.Start)
	r.mux.HandleFunc("POST /api/maintenance/{id}/complete", maintHandler.Complete)
	r.mux.HandleFunc("POST /api/maintenance/{id}/cancel", maintHandler.Cancel)

	// 质检路由
	r.mux.HandleFunc("GET /api/inspections", inspHandler.List)
	r.mux.HandleFunc("POST /api/inspections", inspHandler.Create)
	r.mux.HandleFunc("GET /api/inspections/pest", inspHandler.ListPest)
	r.mux.HandleFunc("GET /api/inspections/{id}", inspHandler.Get)

	return r
}

// ServeHTTP 实现 http.Handler
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}
