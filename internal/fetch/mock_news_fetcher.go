package fetcher

import (
	"log"
	"time"
)

// MockNewsFetcher モックニュース取得
type MockNewsFetcher struct{}

func NewMockNewsFetcher() *MockNewsFetcher {
	return &MockNewsFetcher{}
}

func (m *MockNewsFetcher) FetchTopNews(query string) ([]NewsArticle, error) {
	log.Println("🧪 モックニュースを使用")

	now := time.Now().Format("2006-01-02T15:04:05Z")

	mockNews := []NewsArticle{
		{
			Title:   "原油価格が3週連続で上昇、ガソリン価格への影響は？",
			Content: "国際原油価格が3週連続で上昇している。OPECプラスの減産継続決定により、供給懸念が強まっている。専門家は「年末にかけてガソリン価格が1リットルあたり5円程度上昇する可能性がある」と指摘。消費者への影響が懸念される。",
			URL:     "https://example.com/news/oil-price-rise",
			Date:    now,
		},
		{
			Title:   "電気自動車の普及でガソリン需要が減少傾向",
			Content: "環境意識の高まりと政府の補助金政策により、電気自動車（EV）の販売が好調だ。自動車業界アナリストによると、2025年には国内のEV比率が20%を突破する見込み。長期的にはガソリン需要の減少が予想される。",
			URL:     "https://example.com/news/ev-adoption",
			Date:    now,
		},
		{
			Title:   "円安進行で輸入コスト増、エネルギー価格に影響",
			Content: "為替市場で円安が進行しており、1ドル=150円台を記録。原油などのエネルギー資源は輸入に頼っているため、円安はガソリン価格の押し上げ要因となる。経済産業省は「価格動向を注視する」と表明。",
			URL:     "https://example.com/news/yen-weakness",
			Date:    now,
		},
		{
			Title:   "政府が燃料補助金の延長を検討、家計負担軽減へ",
			Content: "岸田政権は、ガソリン価格高騰対策として実施している燃料補助金制度の延長を検討している。現在の補助金により、小売価格は1リットルあたり約10円抑制されている。延長期間は3ヶ月程度となる見通し。",
			URL:     "https://example.com/news/fuel-subsidy",
			Date:    now,
		},
		{
			Title:   "中東情勢の緊張でエネルギー市場が揺れる",
			Content: "中東地域での地政学的リスクの高まりにより、原油先物価格が急騰。市場関係者は「供給途絶のリスクが意識されている」と分析。日本は中東からの原油輸入が多く、価格変動の影響を受けやすい。",
			URL:     "https://example.com/news/middle-east-tension",
			Date:    now,
		},
	}

	log.Printf("✅ %d件のモックニュースを生成", len(mockNews))
	return mockNews, nil
}
