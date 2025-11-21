package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	fetcher "gasinsight/internal/fetch"

	"golang.org/x/net/html"
)

func main() {
	log.Println("🔍 スクレイピングデバッグモード")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// gogo.gsのHTMLを取得
	client := fetcher.NewHTTPClient(15 * time.Second)

	log.Println("\n━━━ gogo.gs をチェック ━━━")
	checkGogoGS(ctx, client)

	log.Println("\n━━━ 経済産業省 をチェック ━━━")
	checkMETI(ctx, client)
}

func checkGogoGS(ctx context.Context, client *fetcher.HTTPClient) {
	htmlContent, err := client.Get(ctx, "https://gogo.gs/")
	if err != nil {
		log.Printf("❌ 取得失敗: %v", err)
		return
	}

	log.Printf("✅ HTML取得成功（%d bytes）", len(htmlContent))

	// HTMLの最初の2000文字を表示
	if len(htmlContent) > 2000 {
		fmt.Println(htmlContent[:2000])
	} else {
		fmt.Println(htmlContent)
	}

	// HTMLをパース
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		log.Printf("❌ パースエラー: %v", err)
		return
	}

	// "円"を含むテキストを全て抽出
	log.Println("\n💴 「円」を含むテキスト:")
	findPriceTexts(doc)

	// 価格らしい数字を全て抽出
	log.Println("\n🔢 価格らしい数字:")
	findPriceNumbers(htmlContent)
}

func checkMETI(ctx context.Context, client *fetcher.HTTPClient) {
	url := "https://www.enecho.meti.go.jp/statistics/petroleum_and_lpgas/pl007/results.html"
	htmlContent, err := client.Get(ctx, url)
	if err != nil {
		log.Printf("❌ 取得失敗: %v", err)
		return
	}

	log.Printf("✅ HTML取得成功（%d bytes）", len(htmlContent))

	// 価格らしい数字を全て抽出
	log.Println("\n🔢 価格らしい数字:")
	findPriceNumbers(htmlContent)
}

func findPriceTexts(n *html.Node) {
	if n.Type == html.TextNode {
		text := strings.TrimSpace(n.Data)
		if strings.Contains(text, "円") && len(text) > 0 && len(text) < 50 {
			fmt.Printf("  - %s\n", text)
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		findPriceTexts(c)
	}
}

func findPriceNumbers(htmlContent string) {
	lines := strings.Split(htmlContent, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 150-180円の範囲の数字を探す
		if strings.Contains(line, "15") || strings.Contains(line, "16") ||
			strings.Contains(line, "17") || strings.Contains(line, "18") {
			if len(line) < 200 && (strings.Contains(line, "円") ||
				strings.Contains(line, "price") || strings.Contains(line, "yen")) {
				fmt.Printf("  - %s\n", line)
			}
		}
	}
}
