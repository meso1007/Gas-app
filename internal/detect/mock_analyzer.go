package detect

import (
	fetcher "gasinsight/internal/fetch"
	"log"
)

// MockAnalyzeNews モックニュース分析（API不要）
func MockAnalyzeNews(article fetcher.NewsArticle) (*AnalyzedNews, error) {
	log.Printf("🧪 モック分析を使用: %s", article.Title)

	// タイトルに基づいて感情を自動判定
	sentiment := "ニュートラル"
	if containsKeyword(article.Title, []string{"上昇", "増加", "高騰", "緊張", "リスク"}) {
		sentiment = "ネガティブ"
	} else if containsKeyword(article.Title, []string{"補助", "軽減", "普及", "好調"}) {
		sentiment = "ポジティブ"
	}

	// モック要約を生成
	summary := "【要約】\n" + article.Content + "\n\n【感情分析】\n" + sentiment + "\n\n【ガソリン価格への影響】\n中"

	return &AnalyzedNews{
		Title:     article.Title,
		Summary:   summary,
		Sentiment: sentiment,
		URL:       article.URL,
		Date:      article.Date,
	}, nil
}

// containsKeyword タイトルにキーワードが含まれているか確認
func containsKeyword(title string, keywords []string) bool {
	for _, keyword := range keywords {
		if len(title) > 0 && len(keyword) > 0 {
			// 簡易的な文字列検索
			for i := 0; i+len(keyword) <= len(title); i++ {
				if title[i:i+len(keyword)] == keyword {
					return true
				}
			}
		}
	}
	return false
}
