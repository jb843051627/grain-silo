package store

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// DB 数据库连接
type DB struct {
	conn *sql.DB
}

// NewDB 创建数据库连接
func NewDB(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	conn.SetMaxOpenConns(1)
	conn.SetMaxIdleConns(1)
	conn.SetConnMaxLifetime(0)

	db := &DB{conn: conn}
	if err := db.initSchema(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("init schema: %w", err)
	}
	return db, nil
}

// Close 关闭连接
func (db *DB) Close() error {
	return db.conn.Close()
}

// initSchema 初始化数据库表结构
func (db *DB) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS silos (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		capacity REAL NOT NULL,
		current_load REAL NOT NULL DEFAULT 0,
		grain_type TEXT NOT NULL DEFAULT '',
		status TEXT NOT NULL DEFAULT 'empty',
		location TEXT NOT NULL DEFAULT '',
		temp_sensor TEXT NOT NULL DEFAULT '',
		humid_sensor TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS transfers (
		id TEXT PRIMARY KEY,
		from_silo_id TEXT NOT NULL,
		to_silo_id TEXT NOT NULL,
		grain_type TEXT NOT NULL,
		quantity REAL NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending',
		scheduled_at DATETIME NOT NULL,
		completed_at DATETIME,
		operator TEXT NOT NULL DEFAULT '',
		remark TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS readings (
		id TEXT PRIMARY KEY,
		silo_id TEXT NOT NULL,
		temp REAL NOT NULL,
		humidity REAL NOT NULL,
		recorded_at DATETIME NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_readings_silo ON readings(silo_id, recorded_at);

	CREATE TABLE IF NOT EXISTS alerts (
		id TEXT PRIMARY KEY,
		silo_id TEXT NOT NULL,
		level TEXT NOT NULL,
		message TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'active',
		acked_by TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_alerts_silo ON alerts(silo_id, status);

	CREATE TABLE IF NOT EXISTS maintenance_tasks (
		id TEXT PRIMARY KEY,
		silo_id TEXT NOT NULL,
		type TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		status TEXT NOT NULL DEFAULT 'scheduled',
		scheduled_at DATETIME NOT NULL,
		started_at DATETIME,
		completed_at DATETIME,
		technician TEXT NOT NULL DEFAULT '',
		cost REAL NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS inspections (
		id TEXT PRIMARY KEY,
		silo_id TEXT NOT NULL,
		grain_type TEXT NOT NULL,
		moisture REAL NOT NULL,
		impurity REAL NOT NULL,
		pest_detected INTEGER NOT NULL DEFAULT 0,
		inspector TEXT NOT NULL,
		result TEXT NOT NULL,
		inspected_at DATETIME NOT NULL,
		created_at DATETIME NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_inspections_silo ON inspections(silo_id, inspected_at);

	CREATE TABLE IF NOT EXISTS audit_logs (
		id TEXT PRIMARY KEY,
		action TEXT NOT NULL,
		entity TEXT NOT NULL,
		entity_id TEXT NOT NULL,
		operator TEXT NOT NULL,
		detail TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_audit_entity ON audit_logs(entity, entity_id);
	`
	_, err := db.conn.Exec(schema)
	return err
}

// now 返回当前时间
func now() time.Time {
	return time.Now().UTC()
}
