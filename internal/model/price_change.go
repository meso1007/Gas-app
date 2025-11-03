package models

import "time"

// PriceChange 価格変動データ
type PriceChange struct {
	ID            string  `json:"id"`              // プライマリキー
	Date          string  `json:"date"`            // 日付
	PriceType     string  `json:"price_type"`      // 価格タイプ（regular/premium/diesel）
	PreviousPrice float64 `json:"previous_price"`  // 前回価格
	CurrentPrice  float64 `json:"current_price"`   // 現在価格
	ChangeAmount  float64 `json:"change_amount"`   // 変動額
	ChangePercent float64 `json:"change_percent"`  // 変動率(%)
	IsAlert       bool    `json:"is_alert"`        // アラート対象か
	CreatedAt     int64   `json:"created_at"`      // 作成タイムスタンプ
}

// NewPriceChange 新しいPriceChangeインスタンスを作成
func NewPriceChange(date, priceType string, prevPrice, currPrice float64) *PriceChange {
	changeAmount := currPrice - prevPrice
	changePercent := 0.0
	if prevPrice > 0 {
		changePercent = (changeAmount / prevPrice) * 100
	}

	// 5%以上の変動をアラートとする
	isAlert := changePercent >= 5.0 || changePercent <= -5.0

	return &PriceChange{
		ID:            date + "_" + priceType,
		Date:          date,
		PriceType:     priceType,
		PreviousPrice: prevPrice,
		CurrentPrice:  currPrice,
		ChangeAmount:  changeAmount,
		ChangePercent: changePercent,
		IsAlert:       isAlert,
		CreatedAt:     time.Now().Unix(),
	}
}

// ExchangeRateChange 為替レート変動データ
type ExchangeRateChange struct {
	ID            string  `json:"id"`
	Date          string  `json:"date"`
	Currency      string  `json:"currency"`       // 通貨（USD/EUR/GBP/CNY）
	PreviousRate  float64 `json:"previous_rate"`  // 前回レート
	CurrentRate   float64 `json:"current_rate"`   // 現在レート
	ChangeAmount  float64 `json:"change_amount"`  // 変動額
	ChangePercent float64 `json:"change_percent"` // 変動率(%)
	IsAlert       bool    `json:"is_alert"`
	CreatedAt     int64   `json:"created_at"`
}

// NewExchangeRateChange 新しいExchangeRateChangeインスタンスを作成
func NewExchangeRateChange(date, currency string, prevRate, currRate float64) *ExchangeRateChange {
	changeAmount := currRate - prevRate
	changePercent := 0.0
	if prevRate > 0 {
		changePercent = (changeAmount / prevRate) * 100
	}

	isAlert := changePercent >= 3.0 || changePercent <= -3.0

	return &ExchangeRateChange{
		ID:            date + "_" + currency,
		Date:          date,
		Currency:      currency,
		PreviousRate:  prevRate,
		CurrentRate:   currRate,
		ChangeAmount:  changeAmount,
		ChangePercent: changePercent,
		IsAlert:       isAlert,
		CreatedAt:     time.Now().Unix(),
	}
}
