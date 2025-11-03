package models

import "time"

// ExchangeRate 為替レートのモデル
type ExchangeRate struct {
	ID        string  `json:"id"`         // プライマリキー: YYYY-MM-DD
	Date      string  `json:"date"`       // 日付
	USDJPY    float64 `json:"usd_jpy"`    // 米ドル/円
	EURJPY    float64 `json:"eur_jpy"`    // ユーロ/円
	GBPJPY    float64 `json:"gbp_jpy"`    // 英ポンド/円
	CNYJPY    float64 `json:"cny_jpy"`    // 中国元/円
	Source    string  `json:"source"`     // データソース
	CreatedAt int64   `json:"created_at"` // 作成タイムスタンプ
	UpdatedAt int64   `json:"updated_at"` // 更新タイムスタンプ
}

// NewExchangeRate 新しいExchangeRateインスタンスを作成
func NewExchangeRate(date string, usdJpy, eurJpy, gbpJpy, cnyJpy float64) *ExchangeRate {
	now := time.Now().Unix()
	return &ExchangeRate{
		ID:        date,
		Date:      date,
		USDJPY:    usdJpy,
		EURJPY:    eurJpy,
		GBPJPY:    gbpJpy,
		CNYJPY:    cnyJpy,
		Source:    "exchangerate-api.com",
		CreatedAt: now,
		UpdatedAt: now,
	}
}
