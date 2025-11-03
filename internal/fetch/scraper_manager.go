package fetcher

import (
	"context"
	"fmt"
	"log"
)

// ScraperManager スクレイパーマネージャー
type ScraperManager struct {
	scrapers []PriceScraper
}

// PriceScraper スクレイパーのインターフェース
type PriceScraper interface {
	Scrape(ctx context.Context) (*GasPriceData, error)
}

// NewScraperManager スクレイパーマネージャーを作成
func NewScraperManager() *ScraperManager {
	return &ScraperManager{
		scrapers: []PriceScraper{
			NewGogoGSScraper(), // gogo.gsのみ使用
		},
	}
}

// ScrapeWithFallback フォールバック機能付きスクレイピング
func (sm *ScraperManager) ScrapeWithFallback(ctx context.Context, useMock bool) (*GasPriceData, error) {
	log.Println("🚀 スクレイピング開始...")

	// 各スクレイパーを順番に試す
	for i, scraper := range sm.scrapers {
		log.Printf("📡 スクレイパー[%d]を試行中...", i+1)
		data, err := scraper.Scrape(ctx)
		if err == nil && data != nil {
			log.Println("✅ スクレイピング成功")
			return data, nil
		}
		log.Printf("⚠️  スクレイピング失敗: %v", err)
	}

	// 全て失敗した場合、モックにフォールバック
	if useMock {
		log.Println("🧪 フォールバック: モックデータを使用")
		mockFetcher := NewMockGasPriceFetcher()
		return mockFetcher.FetchLatestPrice()
	}

	return nil, fmt.Errorf("全てのスクレイパーが失敗しました")
}

// ScrapeAll 全スクレイパーを実行（将来的に複数ソース対応）
func (sm *ScraperManager) ScrapeAll(ctx context.Context) ([]*GasPriceData, error) {
	var results []*GasPriceData

	for _, scraper := range sm.scrapers {
		data, err := scraper.Scrape(ctx)
		if err == nil && data != nil {
			results = append(results, data)
		}
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("全てのスクレイパーが失敗しました")
	}

	return results, nil
}
