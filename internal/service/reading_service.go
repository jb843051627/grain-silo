package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"grain-silo/internal/model"
	"grain-silo/internal/store"
)

// ReadingService 读数服务
type ReadingService struct {
	readingStore *store.ReadingStore
	siloStore    *store.SiloStore
	alertStore   *store.AlertStore
	auditStore   *store.AuditStore
	latestReadings map[string]*model.Reading
	mu            sync.RWMutex
}

// NewReadingService 创建读数服务
func NewReadingService(readingStore *store.ReadingStore, siloStore *store.SiloStore, alertStore *store.AlertStore, auditStore *store.AuditStore) *ReadingService {
	return &ReadingService{
		readingStore:   readingStore,
		siloStore:      siloStore,
		alertStore:     alertStore,
		auditStore:     auditStore,
		latestReadings: make(map[string]*model.Reading),
	}
}

// RecordReading 记录读数
func (s *ReadingService) RecordReading(siloID string, temp, humidity float64) (*model.Reading, error) {
	silo, err := s.siloStore.GetByID(siloID)
	if err != nil {
		return nil, err
	}
	_ = silo
	now := time.Now().UTC()
	reading := &model.Reading{
		ID:        uuid.NewString(),
		SiloID:    siloID,
		Temp:      temp,
		Humidity:  humidity,
		RecordedAt: now,
	}
	if err := s.readingStore.Create(reading); err != nil {
		return nil, err
	}
	s.mu.Lock()
	s.latestReadings[siloID] = reading
	s.mu.Unlock()
	s.checkThresholds(siloID, temp, humidity)
	return reading, nil
}

// BatchIngestReadings 批量录入读数
func (s *ReadingService) BatchIngestReadings(ctx context.Context, readings []*model.Reading) (int, error) {
	count := 0
	for i, r := range readings {
		select {
		case <-ctx.Done():
			return count, ctx.Err()
		default:
		}
		if ctx.Err() != nil {
			return count, ctx.Err()
		}
		r.ID = uuid.NewString()
		if r.RecordedAt.IsZero() {
			r.RecordedAt = time.Now().UTC()
		}
		if err := s.readingStore.Create(r); err != nil {
			return count, fmt.Errorf("batch insert at %d: %w", i, err)
		}
		s.mu.Lock()
		s.latestReadings[r.SiloID] = r
		s.mu.Unlock()
		s.checkThresholds(r.SiloID, r.Temp, r.Humidity)
		count++
	}
	return count, nil
}

// GetLatestReading 获取最新读数
func (s *ReadingService) GetLatestReading(siloID string) (*model.Reading, error) {
	s.mu.RLock()
	if r, ok := s.latestReadings[siloID]; ok {
		s.mu.RUnlock()
		return r, nil
	}
	s.mu.RUnlock()
	return s.readingStore.GetLatest(siloID)
}

// ListReadingsBySilo 按筒仓列出读数
func (s *ReadingService) ListReadingsBySilo(siloID string, limit int) ([]*model.Reading, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	return s.readingStore.ListBySilo(siloID, limit)
}

// ListReadingsByTimeRange 按时间范围列出读数
func (s *ReadingService) ListReadingsByTimeRange(siloID string, start, end time.Time) ([]*model.Reading, error) {
	return s.readingStore.ListByTimeRange(siloID, start, end)
}

// GetSiloMetrics 获取筒仓传感器指标
func (s *ReadingService) GetSiloMetrics(siloID string) (*model.SiloMetrics, error) {
	silo, err := s.siloStore.GetByID(siloID)
	if err != nil {
		return nil, err
	}
	avgTemp, avgHumidity, err := s.readingStore.GetAvg24h(siloID)
	if err != nil {
		return nil, err
	}
	alertCount, err := s.alertStore.CountActive(siloID)
	if err != nil {
		return nil, err
	}
	utilization := 0.0
	if silo.Capacity > 0 {
		utilization = silo.CurrentLoad / silo.Capacity * 100
	}
	return &model.SiloMetrics{
		SiloID:         siloID,
		CurrentLoad:    silo.CurrentLoad,
		Utilization:    utilization,
		AvgTemp24h:     avgTemp,
		AvgHumidity24h: avgHumidity,
		AlertCount:     alertCount,
	}, nil
}

// CleanupOldReadings 清理旧读数
func (s *ReadingService) CleanupOldReadings(days int) (int64, error) {
	if days <= 0 {
		days = 90
	}
	before := time.Now().UTC().AddDate(0, 0, -days)
	return s.readingStore.DeleteOlderThan(before)
}

// checkThresholds 检查温度湿度阈值
func (s *ReadingService) checkThresholds(siloID string, temp, humidity float64) {
	const (
		tempHigh     = 35.0
		tempLow      = 0.0
		humidityHigh = 75.0
	)
	if temp > tempHigh {
		s.createAlert(siloID, model.ALERT_CRITICAL, fmt.Sprintf("temperature %.1f exceeds high threshold %.1f", temp, tempHigh))
	} else if temp < tempLow {
		s.createAlert(siloID, model.ALERT_WARNING, fmt.Sprintf("temperature %.1f below low threshold %.1f", temp, tempLow))
	}
	if humidity > humidityHigh {
		s.createAlert(siloID, model.ALERT_WARNING, fmt.Sprintf("humidity %.1f exceeds threshold %.1f", humidity, humidityHigh))
	}
}

// createAlert 创建告警
func (s *ReadingService) createAlert(siloID string, level model.AlertLevel, message string) {
	now := time.Now().UTC()
	alert := &model.Alert{
		ID:        uuid.NewString(),
		SiloID:    siloID,
		Level:     level,
		Message:   message,
		Status:    model.ALERT_ACTIVE,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.alertStore.Create(alert)
}

// logAudit 记录审计日志
func (s *ReadingService) logAudit(action, entity, entityID, operator, detail string) {
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
