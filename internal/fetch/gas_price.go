package fetcher

import (
	"log"
	"time"
)

type GasPriceData struct {
	Date         string
	RegularPrice float64
	PremiumPrice float64
	DieselPrice  float64
	Region       string
}

type MockGasPriceFetcher struct{}

func NewMockGasPriceFetcher() *MockGasPriceFetcher {
	return &MockGasPriceFetcher{}
}

func (m *MockGasPriceFetcher) FetchLatestPrice() (*GasPriceData, error) {
	now := time.Now().Format("2006-01-02")
	log.Println("🧪 モックデータを使用してガソリン価格を取得")
	return &GasPriceData{
		Date:         now,
		RegularPrice: 180,
		PremiumPrice: 179.2,
		DieselPrice:  148.8,
		Region:       "全国平均",
	}, nil
}
