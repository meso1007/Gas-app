package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// ExchangeRateData 為替レートデータ
type ExchangeRateData struct {
	Date   string
	USDJPY float64
	EURJPY float64
	GBPJPY float64
	CNYJPY float64
}

// ExchangeRateFetcher 為替レートフェッチャー
type ExchangeRateFetcher struct {
	httpClient *HTTPClient
	baseURL    string
}

// NewExchangeRateFetcher 為替レートフェッチャーを作成
func NewExchangeRateFetcher() *ExchangeRateFetcher {
	return &ExchangeRateFetcher{
		httpClient: NewHTTPClient(10 * time.Second),
		baseURL:    "https://api.exchangerate-api.com/v4/latest/JPY",
	}
}

// Fetch 為替レートを取得
func (e *ExchangeRateFetcher) Fetch(ctx context.Context) (*ExchangeRateData, error) {
	log.Println("🌍 為替レートを取得中...")

	// APIからデータ取得
	htmlContent, err := e.httpClient.Get(ctx, e.baseURL)
	if err != nil {
		return nil, fmt.Errorf("API取得エラー: %w", err)
	}

	// JSONをパース
	var apiResponse struct {
		Date  string             `json:"date"`
		Rates map[string]float64 `json:"rates"`
	}

	if err := json.Unmarshal([]byte(htmlContent), &apiResponse); err != nil {
		return nil, fmt.Errorf("JSONパースエラー: %w", err)
	}

	// JPYベースなので、逆数を計算（1 USD = X JPY）
	data := &ExchangeRateData{
		Date: time.Now().Format("2006-01-02"),
	}

	// JPY建てレートに変換
	if usd, ok := apiResponse.Rates["USD"]; ok && usd > 0 {
		data.USDJPY = 1.0 / usd
	}
	if eur, ok := apiResponse.Rates["EUR"]; ok && eur > 0 {
		data.EURJPY = 1.0 / eur
	}
	if gbp, ok := apiResponse.Rates["GBP"]; ok && gbp > 0 {
		data.GBPJPY = 1.0 / gbp
	}
	if cny, ok := apiResponse.Rates["CNY"]; ok && cny > 0 {
		data.CNYJPY = 1.0 / cny
	}

	log.Printf("✅ USD/JPY: %.2f円", data.USDJPY)
	log.Printf("✅ EUR/JPY: %.2f円", data.EURJPY)
	log.Printf("✅ GBP/JPY: %.2f円", data.GBPJPY)
	log.Printf("✅ CNY/JPY: %.2f円", data.CNYJPY)

	return data, nil
}

// MockExchangeRateFetcher モック用フェッチャー
type MockExchangeRateFetcher struct{}

func NewMockExchangeRateFetcher() *MockExchangeRateFetcher {
	return &MockExchangeRateFetcher{}
}

func (m *MockExchangeRateFetcher) Fetch(ctx context.Context) (*ExchangeRateData, error) {
	log.Println("🧪 モック為替データを使用")
	return &ExchangeRateData{
		Date:   time.Now().Format("2006-01-02"),
		USDJPY: 150.25,
		EURJPY: 163.80,
		GBPJPY: 190.50,
		CNYJPY: 20.85,
	}, nil
}
