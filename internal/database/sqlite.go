package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"gasinsight/internal/models"
)

type SQLiteClient struct {
	db *sql.DB
}

func NewSQLiteClient(dbPath string) (*SQLiteClient, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("ディレクトリ作成エラー: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("データベースオープンエラー: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("データベース接続エラー: %w", err)
	}

	client := &SQLiteClient{db: db}
	if err := client.createTables(); err != nil {
		return nil, err
	}

	log.Printf("✅ SQLiteデータベースに接続: %s", dbPath)
	return client, nil
}

func (s *SQLiteClient) createTables() error {
	// ガソリン価格テーブル
	gasPriceQuery := `CREATE TABLE IF NOT EXISTS gas_prices (
		id TEXT PRIMARY KEY,
		date TEXT NOT NULL,
		regular_price REAL NOT NULL,
		premium_price REAL NOT NULL,
		diesel_price REAL NOT NULL,
		region TEXT NOT NULL,
		source TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	);`

	if _, err := s.db.Exec(gasPriceQuery); err != nil {
		return fmt.Errorf("ガソリン価格テーブル作成エラー: %w", err)
	}

	// 為替レートテーブルを作成
	if err := s.CreateExchangeRateTable(); err != nil {
		return err
	}

	log.Println("✅ 全テーブルを作成しました")
	return nil
}

func (s *SQLiteClient) SaveGasPrice(price *models.GasPrice) error {
	query := `INSERT OR REPLACE INTO gas_prices 
		(id, date, regular_price, premium_price, diesel_price, region, source, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.Exec(query, price.ID, price.Date, price.RegularPrice,
		price.PremiumPrice, price.DieselPrice, price.Region, price.Source,
		price.CreatedAt, price.UpdatedAt)

	if err != nil {
		return fmt.Errorf("データ保存エラー: %w", err)
	}

	log.Printf("✅ ガソリン価格を保存: %s", price.Date)
	return nil
}

func (s *SQLiteClient) GetAllGasPrices() ([]*models.GasPrice, error) {
	query := `SELECT id, date, regular_price, premium_price, diesel_price,
		region, source, created_at, updated_at FROM gas_prices ORDER BY date DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prices []*models.GasPrice
	for rows.Next() {
		var p models.GasPrice
		if err := rows.Scan(&p.ID, &p.Date, &p.RegularPrice, &p.PremiumPrice,
			&p.DieselPrice, &p.Region, &p.Source, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		prices = append(prices, &p)
	}

	return prices, nil
}

func (s *SQLiteClient) GetLatestGasPrice() (*models.GasPrice, error) {
	query := `SELECT id, date, regular_price, premium_price, diesel_price,
		region, source, created_at, updated_at FROM gas_prices ORDER BY date DESC LIMIT 1`

	var p models.GasPrice
	err := s.db.QueryRow(query).Scan(&p.ID, &p.Date, &p.RegularPrice,
		&p.PremiumPrice, &p.DieselPrice, &p.Region, &p.Source, &p.CreatedAt, &p.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("データなし")
	}
	return &p, err
}

func (s *SQLiteClient) Close() error {
	if s.db != nil {
		log.Println("📪 データベース接続を閉じます")
		return s.db.Close()
	}
	return nil
}

// GetGasPriceByDate 特定日付のガソリン価格を取得
func (s *SQLiteClient) GetGasPriceByDate(date string) (*models.GasPrice, error) {
	query := `SELECT id, date, regular_price, premium_price, diesel_price,
		region, source, created_at, updated_at FROM gas_prices WHERE date = ?`

	var p models.GasPrice
	err := s.db.QueryRow(query, date).Scan(&p.ID, &p.Date, &p.RegularPrice,
		&p.PremiumPrice, &p.DieselPrice, &p.Region, &p.Source, &p.CreatedAt, &p.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("指定日付のデータが見つかりません: %s", date)
	}
	return &p, err
}
