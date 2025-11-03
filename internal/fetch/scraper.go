package fetcher

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// HTTPClient HTTPクライアント
type HTTPClient struct {
	client *http.Client
}

// NewHTTPClient 新しいHTTPクライアントを作成
func NewHTTPClient(timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:    10,
				IdleConnTimeout: 30 * time.Second,
			},
		},
	}
}

// Get HTTPリクエストを実行
func (h *HTTPClient) Get(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("リクエスト作成エラー: %w", err)
	}

	// ヘッダー設定（ブラウザのふりをする）
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "ja,en-US;q=0.9,en;q=0.8")

	log.Printf("🌐 HTTP GET: %s", url)
	resp, err := h.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("リクエスト実行エラー: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTPエラー: status=%d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("レスポンス読み込みエラー: %w", err)
	}

	return string(body), nil
}

// ParsePrice 価格文字列から数値を抽出（例: "168.5円" -> 168.5）
func ParsePrice(priceStr string) (float64, error) {
	// 数字とドットのみ抽出
	re := regexp.MustCompile(`(\d+\.?\d*)`)
	matches := re.FindStringSubmatch(priceStr)

	if len(matches) < 2 {
		return 0, fmt.Errorf("価格の抽出に失敗: %s", priceStr)
	}

	price, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, fmt.Errorf("数値変換エラー: %w", err)
	}

	return price, nil
}

// FindNodeByText テキストを含むノードを検索
func FindNodeByText(n *html.Node, targetText string) *html.Node {
	if n.Type == html.TextNode {
		if strings.Contains(strings.TrimSpace(n.Data), targetText) {
			return n.Parent
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result := FindNodeByText(c, targetText); result != nil {
			return result
		}
	}

	return nil
}

// GetNodeText ノードのテキストを取得
func GetNodeText(n *html.Node) string {
	if n == nil {
		return ""
	}

	if n.Type == html.TextNode {
		return strings.TrimSpace(n.Data)
	}

	var text strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text.WriteString(GetNodeText(c))
	}

	return strings.TrimSpace(text.String())
}

// FindNodesByTag タグ名でノードを検索
func FindNodesByTag(n *html.Node, tag string) []*html.Node {
	var nodes []*html.Node

	var traverse func(*html.Node)
	traverse = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == tag {
			nodes = append(nodes, node)
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(n)
	return nodes
}

// FindNodeByClass クラス名でノードを検索
func FindNodeByClass(n *html.Node, className string) *html.Node {
	if n.Type == html.ElementNode {
		for _, attr := range n.Attr {
			if attr.Key == "class" && strings.Contains(attr.Val, className) {
				return n
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result := FindNodeByClass(c, className); result != nil {
			return result
		}
	}

	return nil
}
