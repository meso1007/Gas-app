package database

import (
	"database/sql"
	"fmt"
	"log"

	model "gasinsight/internal/model"
)

// CreateExchangeRateTable 為替レートテーブルを作成
func (s *SQLiteClient) CreateExchangeRateTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS exchange_rates (
		id TEXT PRIMARY KEY,
		date TEXT NOT NULL,
		usd_jpy REAL NOT NULL,
		eur_jpy REAL NOT NULL,
		gbp_jpy REAL NOT NULL,
		cny_jpy REAL NOT NULL,
		source TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_exchange_rates_date ON exchange_rates(date);
	`

	if _, err := s.db.Exec(query); err != nil {
		return fmt.Errorf("為替レートテーブル作成エラー: %w", err)
	}

	log.Println("✅ 為替レートテーブルを作成しました")
	return nil
}

// SaveExchangeRate 為替レートを保存
func (s *SQLiteClient) SaveExchangeRate(rate *model.ExchangeRate) error {
	query := `
		INSERT OR REPLACE INTO exchange_rates 
		(id, date, usd_jpy, eur_jpy, gbp_jpy, cny_jpy, source, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.Exec(query,
		rate.ID,
		rate.Date,
		rate.USDJPY,
		rate.EURJPY,
		rate.GBPJPY,
		rate.CNYJPY,
		rate.Source,
		rate.CreatedAt,
		rate.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("為替レート保存エラー: %w", err)
	}

	log.Printf("✅ 為替レートを保存: %s", rate.Date)
	return nil
}

// GetAllExchangeRates 全ての為替レートを取得
func (s *SQLiteClient) GetAllExchangeRates() ([]*model.ExchangeRate, error) {
	query := `
		SELECT id, date, usd_jpy, eur_jpy, gbp_jpy, cny_jpy, source, created_at, updated_at
		FROM exchange_rates
		ORDER BY date DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rates []*model.ExchangeRate
	for rows.Next() {
		var rate model.ExchangeRate
		err := rows.Scan(
			&rate.ID,
			&rate.Date,
			&rate.USDJPY,
			&rate.EURJPY,
			&rate.GBPJPY,
			&rate.CNYJPY,
			&rate.Source,
			&rate.CreatedAt,
			&rate.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		rates = append(rates, &rate)
	}

	return rates, nil
}

// GetLatestExchangeRate 最新の為替レートを取得
func (s *SQLiteClient) GetLatestExchangeRate() (*model.ExchangeRate, error) {
	query := `
		SELECT id, date, usd_jpy, eur_jpy, gbp_jpy, cny_jpy, source, created_at, updated_at
		FROM exchange_rates
		ORDER BY date DESC
		LIMIT 1`

	var rate model.ExchangeRate
	err := s.db.QueryRow(query).Scan(
		&rate.ID,
		&rate.Date,
		&rate.USDJPY,
		&rate.EURJPY,
		&rate.GBPJPY,
		&rate.CNYJPY,
		&rate.Source,
		&rate.CreatedAt,
		&rate.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("為替データが見つかりません")
	}
	if err != nil {
		return nil, err
	}

	return &rate, nil
}

// GetExchangeRateByDate 特定日付の為替レートを取得
func (s *SQLiteClient) GetExchangeRateByDate(date string) (*model.ExchangeRate, error) {
	query := `
		SELECT id, date, usd_jpy, eur_jpy, gbp_jpy, cny_jpy, source, created_at, updated_at
		FROM exchange_rates
		WHERE date = ?`

	var rate model.ExchangeRate
	err := s.db.QueryRow(query, date).Scan(
		&rate.ID,
		&rate.Date,
		&rate.USDJPY,
		&rate.EURJPY,
		&rate.GBPJPY,
		&rate.CNYJPY,
		&rate.Source,
		&rate.CreatedAt,
		&rate.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("指定日付の為替データが見つかりません: %s", date)
	}
	if err != nil {
		return nil, err
	}

	return &rate, nil
}
