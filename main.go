package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"grain-silo/internal/api"
	"grain-silo/internal/handler"
	"grain-silo/internal/service"
	"grain-silo/internal/store"
)

func main() {
	dbPath := os.Getenv("GRAIN_SILO_DB")
	if dbPath == "" {
		dbPath = filepath.Join(os.TempDir(), "grain_silo.db")
	}

	db, err := store.NewDB(dbPath)
	if err != nil {
		log.Fatalf("failed to init database: %v", err)
	}
	defer db.Close()

	siloStore := store.NewSiloStore(db)
	transferStore := store.NewTransferStore(db)
	readingStore := store.NewReadingStore(db)
	alertStore := store.NewAlertStore(db)
	maintStore := store.NewMaintenanceStore(db)
	inspStore := store.NewInspectionStore(db)
	auditStore := store.NewAuditStore(db)

	siloSvc := service.NewSiloService(siloStore, readingStore, alertStore, auditStore)
	transferSvc := service.NewTransferService(transferStore, siloStore, auditStore)
	readingSvc := service.NewReadingService(readingStore, siloStore, alertStore, auditStore)
	alertSvc := service.NewAlertService(alertStore, siloStore, auditStore)
	maintSvc := service.NewMaintenanceService(maintStore, siloStore, auditStore)
	inspSvc := service.NewInspectionService(inspStore, siloStore, auditStore)

	siloHandler := handler.NewSiloHandler(siloSvc)
	transferHandler := handler.NewTransferHandler(transferSvc)
	readingHandler := handler.NewReadingHandler(readingSvc)
	alertHandler := handler.NewAlertHandler(alertSvc)
	maintHandler := handler.NewMaintenanceHandler(maintSvc)
	inspHandler := handler.NewInspectionHandler(inspSvc)

	router := api.NewRouter(siloHandler, transferHandler, readingHandler, alertHandler, maintHandler, inspHandler)

	port := os.Getenv("GRAIN_SILO_PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("grain-silo server starting on :%s (db=%s)\n", port, dbPath)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
