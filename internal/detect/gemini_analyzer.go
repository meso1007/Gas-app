package detect

import (
	"context"
	"fmt"
	fetcher "gasinsight/internal/fetch"
	"log"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// AnalyzeNewsWithGemini Gemini APIを使ってニュースを分析
func AnalyzeNewsWithGemini(article fetcher.NewsArticle) (*AnalyzedNews, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY環境変数が設定されていません")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("Gemini クライアント作成エラー: %w", err)
	}
	defer client.Close()

	// Gemini 2.0 Flash Lite を使用（高速・軽量）
	model := client.GenerativeModel("gemini-2.0-flash-lite")

	// プロンプトを構築
	prompt := fmt.Sprintf(`以下のニュース記事を分析してください。

タイトル: %s
内容: %s

以下の形式で回答してください：
【要約】
（3行以内で要約）

【感情分析】
（ポジティブ/ニュートラル/ネガティブ のいずれか1つのみ）

【ガソリン価格への影響】
（大/中/小/なし のいずれか1つ）`, article.Title, article.Content)

	log.Printf("🤖 Gemini APIでニュース分析中...")
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		if strings.Contains(err.Error(), "429") {
			return nil, fmt.Errorf("Gemini APIレート制限超過 (429): しばらく待ってから再試行してください。詳細: %w", err)
		}
		return nil, fmt.Errorf("Gemini API呼び出しエラー: %w", err)
	}

	// レスポンスからテキストを取得
	var summary string
	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("Gemini APIからの応答が空または不正です")
	}
	if text, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
		summary = string(text)
	} else {
		return nil, fmt.Errorf("Gemini APIからの応答が予期しない形式です")
	}

	// 感情分析の抽出（簡易版）
	sentiment := "ニュートラル"
	summaryLower := strings.ToLower(summary)
	if strings.Contains(summaryLower, "ポジティブ") || strings.Contains(summaryLower, "positive") {
		sentiment = "ポジティブ"
	} else if strings.Contains(summaryLower, "ネガティブ") || strings.Contains(summaryLower, "negative") {
		sentiment = "ネガティブ"
	}

	// ガソリン価格への影響を抽出
	impact := "なし"
	if strings.Contains(summary, "大") {
		impact = "大"
	} else if strings.Contains(summary, "中") {
		impact = "中"
	} else if strings.Contains(summary, "小") {
		impact = "小"
	}

	log.Printf("✅ 分析完了: %s", sentiment)

	return &AnalyzedNews{
		Title:       article.Title,
		URL:         article.URL,
		Date:        article.Date, // PublishedAt -> Date
		Summary:     summary,
		Sentiment:   sentiment,
		ImpactLevel: impact,
	}, nil
}

// AnalyzePriceChange はガソリン価格の変動とニュース記事を受け取り、変動要因を分析します
func AnalyzePriceChange(ctx context.Context, priceDiff int, oldPrice, newPrice int, newsList []fetcher.NewsArticle) (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY is not set")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return "", fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	// Gemini 2.0 Flash Lite を使用
	model := client.GenerativeModel("gemini-2.0-flash-lite")

	// ニュースリストをテキスト化
	var newsText string
	for i, n := range newsList {
		newsText += fmt.Sprintf("%d. %s\n   (URL: %s)\n", i+1, n.Title, n.URL)
	}

	// プロンプト作成
	prompt := fmt.Sprintf(`
あなたはエネルギー市場のアナリストです。
日本のガソリン価格が以下のように変動しました。
提供されたニュース記事の中から、この価格変動の要因として考えられるものを特定し、その理由を解説してください。

【価格変動データ】
- 変動前: %d円
- 変動後: %d円
- 変動幅: %+d円

【本日のニュース】
%s

【分析依頼】
1. この価格変動に最も影響を与えたと思われるニュースを1つ以上挙げてください。
2. なぜそのニュースが価格に影響したのか、因果関係を論理的に説明してください。
3. もし関連するニュースがない場合は、「関連するニュースは見当たりませんでした」と回答してください。

回答は日本語で、一般のドライバーにも分かりやすく簡潔にお願いします。
`, oldPrice, newPrice, priceDiff, newsText)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("Gemini API error: %w", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response from Gemini")
	}

	if text, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
		return string(text), nil
	}

	return "", fmt.Errorf("unexpected response format")
}
