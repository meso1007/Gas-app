package fetcher

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// GogoGSScraper gogo.gsのスクレイパー
type GogoGSScraper struct {
	httpClient *HTTPClient
	baseURL    string
}

// NewGogoGSScraper gogo.gsスクレイパーを作成
func NewGogoGSScraper() *GogoGSScraper {
	return &GogoGSScraper{
		httpClient: NewHTTPClient(15 * time.Second),
		baseURL:    "https://gogo.gs/",
	}
}

// Scrape ガソリン価格をスクレイピング
func (g *GogoGSScraper) Scrape(ctx context.Context) (*GasPriceData, error) {
	log.Println("🔍 gogo.gsから価格情報を取得中...")

	// HTMLを取得
	htmlContent, err := g.httpClient.Get(ctx, g.baseURL)
	if err != nil {
		return nil, fmt.Errorf("HTML取得エラー: %w", err)
	}

	// HTMLをパース
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("HTMLパースエラー: %w", err)
	}

	// <div class="price">XXX</div> を抽出
	prices := g.extractPrices(doc)

	if len(prices) < 3 {
		return nil, fmt.Errorf("価格情報が不足しています（取得: %d件）", len(prices))
	}

	priceData := &GasPriceData{
		Date:         time.Now().Format("2006-01-02"),
		RegularPrice: prices[0], // 最初がレギュラー
		PremiumPrice: prices[1], // 2番目がハイオク
		DieselPrice:  prices[2], // 3番目が軽油
		Region:       "全国平均（gogo.gs）",
	}

	log.Printf("✅ レギュラー: %.2f円", priceData.RegularPrice)
	log.Printf("✅ ハイオク: %.2f円", priceData.PremiumPrice)
	log.Printf("✅ 軽油: %.2f円", priceData.DieselPrice)

	return priceData, nil
}

// extractPrices <div class="price">XXX</div> から価格を抽出
func (g *GogoGSScraper) extractPrices(n *html.Node) []float64 {
	var prices []float64

	var traverse func(*html.Node)
	traverse = func(node *html.Node) {
		// <div class="price"> を探す
		if node.Type == html.ElementNode && node.Data == "div" {
			for _, attr := range node.Attr {
				if attr.Key == "class" && attr.Val == "price" {
					// このdivのテキストを取得
					text := GetNodeText(node)
					if price, err := ParsePrice(text); err == nil {
						// 妥当な価格範囲かチェック（100円〜300円）
						if price >= 100 && price <= 300 {
							prices = append(prices, price)
							log.Printf("  📍 価格発見: %.2f円", price)
						}
					}
				}
			}
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(n)
	return prices
}
