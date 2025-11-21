package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

type NewsFetcher struct {
	apiKey string
	client *http.Client
}

type NewsArticle struct {
	Title   string
	Content string
	URL     string
	Date    string
}

func NewNewsFetcher(apiKey string) *NewsFetcher {
	return &NewsFetcher{
		apiKey: apiKey,
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

func (n *NewsFetcher) FetchTopNews(query string) ([]NewsArticle, error) {
	// URLパラメータを適切にエンコード
	baseURL := "https://newsapi.org/v2/everything"
	params := url.Values{}
	params.Add("q", query)
	params.Add("sortBy", "publishedAt")
	params.Add("pageSize", "3")
	params.Add("apiKey", n.apiKey)

	fullURL := baseURL + "?" + params.Encode()

	log.Printf("🌐 NewsAPI リクエスト中...")
	log.Printf("   クエリ: %s", query)
	resp, err := n.client.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("HTTPリクエストエラー: %w", err)
	}
	defer resp.Body.Close()

	// レスポンスボディを読み取り
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("レスポンス読み取りエラー: %w", err)
	}

	// ステータスコードをチェック
	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ HTTPエラー詳細: %s", string(body))
		return nil, fmt.Errorf("HTTPエラー: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var result struct {
		Status   string `json:"status"`
		Code     string `json:"code"`
		Message  string `json:"message"`
		Articles []struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			URL         string `json:"url"`
			PublishedAt string `json:"publishedAt"`
		} `json:"articles"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("❌ JSON解析エラー: %s", string(body))
		return nil, fmt.Errorf("JSONデコードエラー: %w", err)
	}

	// APIエラーチェック
	if result.Status == "error" {
		return nil, fmt.Errorf("NewsAPIエラー: [%s] %s", result.Code, result.Message)
	}

	log.Printf("✅ %d件のニュースを取得", len(result.Articles))

	articles := []NewsArticle{}
	for _, a := range result.Articles {
		articles = append(articles, NewsArticle{
			Title:   a.Title,
			Content: a.Description,
			URL:     a.URL,
			Date:    a.PublishedAt,
		})
	}

	return articles, nil
}
