package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

type RequestLog struct {
	ID             int64
	Timestamp      time.Time
	Model          string
	Provider       string
	PromptTokens   int
	CompletionTokens int
	TotalTokens    int
	LatencyMs      int64
	StatusCode     int
	ErrorMessage   string
	ClientIP       string
}

func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// Enable WAL mode for better concurrent read performance
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("set WAL mode: %w", err)
	}

	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return &Store{db: db}, nil
}

func migrate(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS request_logs (
		id              INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp       DATETIME DEFAULT CURRENT_TIMESTAMP,
		model           TEXT NOT NULL,
		provider        TEXT NOT NULL,
		prompt_tokens   INTEGER DEFAULT 0,
		completion_tokens INTEGER DEFAULT 0,
		total_tokens    INTEGER DEFAULT 0,
		latency_ms      INTEGER DEFAULT 0,
		status_code     INTEGER DEFAULT 200,
		error_message   TEXT DEFAULT '',
		client_ip       TEXT DEFAULT ''
	);

	CREATE INDEX IF NOT EXISTS idx_request_logs_timestamp ON request_logs(timestamp);
	CREATE INDEX IF NOT EXISTS idx_request_logs_model ON request_logs(model);
	CREATE INDEX IF NOT EXISTS idx_request_logs_provider ON request_logs(provider);
	`
	_, err := db.Exec(schema)
	return err
}

func (s *Store) LogRequest(log RequestLog) error {
	_, err := s.db.Exec(`
		INSERT INTO request_logs (model, provider, prompt_tokens, completion_tokens, total_tokens, latency_ms, status_code, error_message, client_ip)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		log.Model, log.Provider, log.PromptTokens, log.CompletionTokens,
		log.TotalTokens, log.LatencyMs, log.StatusCode, log.ErrorMessage, log.ClientIP,
	)
	return err
}

type Stats struct {
	TotalRequests    int     `json:"total_requests"`
	TotalTokens      int     `json:"total_tokens"`
	AvgLatencyMs     float64 `json:"avg_latency_ms"`
	ErrorCount       int     `json:"error_count"`
	ModelBreakdown   map[string]int `json:"model_breakdown"`
	ProviderBreakdown map[string]int `json:"provider_breakdown"`
}

func (s *Store) GetStats(since time.Duration) (*Stats, error) {
	cutoff := time.Now().Add(-since)

	stats := &Stats{
		ModelBreakdown:    make(map[string]int),
		ProviderBreakdown: make(map[string]int),
	}

	// Totals
	err := s.db.QueryRow(`
		SELECT COUNT(*), COALESCE(SUM(total_tokens), 0), COALESCE(AVG(latency_ms), 0),
		       COALESCE(SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END), 0)
		FROM request_logs WHERE timestamp >= ?`, cutoff,
	).Scan(&stats.TotalRequests, &stats.TotalTokens, &stats.AvgLatencyMs, &stats.ErrorCount)
	if err != nil {
		return nil, err
	}

	// Model breakdown
	rows, err := s.db.Query(`
		SELECT model, COUNT(*) FROM request_logs WHERE timestamp >= ? GROUP BY model`, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var model string
		var count int
		if err := rows.Scan(&model, &count); err != nil {
			return nil, err
		}
		stats.ModelBreakdown[model] = count
	}

	// Provider breakdown
	rows2, err := s.db.Query(`
		SELECT provider, COUNT(*) FROM request_logs WHERE timestamp >= ? GROUP BY provider`, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows2.Close()
	for rows2.Next() {
		var provider string
		var count int
		if err := rows2.Scan(&provider, &count); err != nil {
			return nil, err
		}
		stats.ProviderBreakdown[provider] = count
	}

	return stats, nil
}

func (s *Store) GetRecentLogs(limit int) ([]RequestLog, error) {
	rows, err := s.db.Query(`
		SELECT id, timestamp, model, provider, prompt_tokens, completion_tokens,
		       total_tokens, latency_ms, status_code, error_message, client_ip
		FROM request_logs ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []RequestLog
	for rows.Next() {
		var l RequestLog
		if err := rows.Scan(&l.ID, &l.Timestamp, &l.Model, &l.Provider,
			&l.PromptTokens, &l.CompletionTokens, &l.TotalTokens,
			&l.LatencyMs, &l.StatusCode, &l.ErrorMessage, &l.ClientIP); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}
