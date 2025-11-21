package detect

import (
	"context"
	fetcher "gasinsight/internal/fetch"
	"os"

	"github.com/sashabaranov/go-openai"
)

type AnalyzedNews struct {
	Title       string
	Summary     string
	Sentiment   string
	ImpactLevel string // ガソリン価格への影響（大・中・小・なし）
	URL         string
	Date        string
}

func AnalyzeNewsWithOpenAI(article fetcher.NewsArticle) (*AnalyzedNews, error) {
	prompt := "以下のニュースを要約して、感情をポジティブ・ニュートラル・ネガティブで判定してください：\n\n" + article.Content

	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	resp, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: "gpt-4",
		Messages: []openai.ChatCompletionMessage{
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return nil, err
	}

	// レスポンスの最初の選択肢のメッセージを取得
	content := ""
	if len(resp.Choices) > 0 {
		content = resp.Choices[0].Message.Content
	}

	return &AnalyzedNews{
		Title:     article.Title,
		Summary:   content,
		Sentiment: "Neutral", // 必要に応じて解析して上書き
		URL:       article.URL,
		Date:      article.Date,
	}, nil
}
