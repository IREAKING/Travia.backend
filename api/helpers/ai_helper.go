package helpers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
)

// GenerateTextWithOpenAI generates text using OpenAI ChatGPT
func GenerateTextWithOpenAI(apiKey string, prompt string, systemPrompt string) (string, error) {
	if apiKey == "" {
		return "", fmt.Errorf("OpenAI API key is not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Tạo OpenAI client
	client := openai.NewClient(apiKey)

	// Xây dựng messages với system prompt và user prompt
	messages := []openai.ChatCompletionMessage{}

	if systemPrompt != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		})
	}

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: prompt,
	})

	// Gọi API với model GPT-3.5-turbo (nhanh và rẻ) hoặc GPT-4 (chất lượng cao hơn)
	req := openai.ChatCompletionRequest{
		Model:       openai.GPT3Dot5Turbo,
		Messages:    messages,
		Temperature: 0.7,
		MaxTokens:   1000,
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("AI trả về kết quả rỗng")
	}

	return resp.Choices[0].Message.Content, nil
}

// GenerateChatbotResponse generates chatbot response with context from chat history and tours
func GenerateChatbotResponse(apiKey string, userMessage string, chatHistory []string, toursList string) (string, error) {
	if apiKey == "" {
		return "", fmt.Errorf("OpenAI API key is not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Tạo OpenAI client
	client := openai.NewClient(apiKey)

	// System prompt
	systemPrompt := `Bạn là trợ lý AI thân thiện và chuyên nghiệp của công ty du lịch Travia.
Nhiệm vụ của bạn là:
- Trả lời câu hỏi của khách hàng về tours, địa điểm, giá cả, lịch trình
- Tư vấn tour phù hợp với nhu cầu khách hàng
- Cung cấp thông tin chính xác về các tour có sẵn
- Thân thiện, nhiệt tình và chuyên nghiệp
- Trả lời bằng tiếng Việt
- Nếu không có thông tin, hãy thành thật và đề xuất liên hệ bộ phận hỗ trợ

Lưu ý:
- Sử dụng thông tin tours từ danh sách có sẵn nếu có
- Tham khảo lịch sử chat để hiểu ngữ cảnh
- Trả lời ngắn gọn, súc tích nhưng đầy đủ thông tin`

	// Xây dựng messages array
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		},
	}

	// Thêm thông tin tours vào system message nếu có
	if toursList != "" {
		messages[0].Content += fmt.Sprintf("\n\nDanh sách tours có sẵn:\n%s", toursList)
	}

	// Thêm chat history (chỉ lấy 10 câu gần nhất để không quá dài)
	startIdx := 0
	if len(chatHistory) > 10 {
		startIdx = len(chatHistory) - 10
	}

	// Parse chat history và thêm vào messages
	// Format: "Q: ...\nA: ..."
	for i := startIdx; i < len(chatHistory); i++ {
		historyItem := chatHistory[i]
		// Tách Q và A
		if len(historyItem) > 3 && historyItem[:2] == "Q:" {
			parts := strings.SplitN(historyItem, "\nA: ", 2)
			if len(parts) == 2 {
				question := strings.TrimPrefix(parts[0], "Q: ")
				answer := parts[1]
				messages = append(messages, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleUser,
					Content: question,
				})
				messages = append(messages, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleAssistant,
					Content: answer,
				})
			}
		}
	}

	// Thêm câu hỏi hiện tại
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: userMessage,
	})

	// Gọi API
	req := openai.ChatCompletionRequest{
		Model:       openai.GPT3Dot5Turbo,
		Messages:    messages,
		Temperature: 0.7,
		MaxTokens:   1000,
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("AI trả về kết quả rỗng")
	}

	return resp.Choices[0].Message.Content, nil
}

// CreateEmbedding tạo embedding vector cho text sử dụng OpenAI
// Trả về vector 1536 chiều (theo chuẩn OpenAI text-embedding-3-small)
func CreateEmbedding(apiKey string, text string) ([]float32, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key is not configured")
	}

	if text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := openai.NewClient(apiKey)

	// Sử dụng text-embedding-3-small (1536 dimensions) hoặc text-embedding-ada-002
	req := openai.EmbeddingRequest{
		Input: []string{text},
		Model: openai.SmallEmbedding3, // text-embedding-3-small (1536 dimensions)
	}

	resp, err := client.CreateEmbeddings(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data returned")
	}

	return resp.Data[0].Embedding, nil
}

// GenerateTourRecommendation tạo gợi ý tour dựa trên sở thích và lịch sử của người dùng
func GenerateTourRecommendation(apiKey string, userPreferences string, viewHistory string, availableTours string) (string, error) {
	if apiKey == "" {
		return "", fmt.Errorf("OpenAI API key is not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := openai.NewClient(apiKey)

	systemPrompt := `Bạn là chuyên gia tư vấn du lịch AI của công ty Travia.
Nhiệm vụ của bạn là phân tích sở thích và lịch sử xem tour của khách hàng để đưa ra gợi ý tour phù hợp nhất.

Hướng dẫn:
- Phân tích sở thích khách hàng (danh mục tour yêu thích, điểm đến quan tâm)
- Xem xét lịch sử xem tour (tour nào khách xem lâu, quan tâm nhiều)
- So sánh với danh sách tour có sẵn
- Đưa ra 3-5 gợi ý tour phù hợp nhất với lý do cụ thể
- Trả lời bằng tiếng Việt, thân thiện và chuyên nghiệp
- Nếu không đủ thông tin, đề xuất tour phổ biến hoặc tour nổi bật`

	userPrompt := fmt.Sprintf(`Dựa trên thông tin sau, hãy đưa ra gợi ý tour phù hợp:

SỞ THÍCH KHÁCH HÀNG:
%s

LỊCH SỬ XEM TOUR:
%s

DANH SÁCH TOUR CÓ SẴN:
%s

Hãy phân tích và đưa ra gợi ý tour tốt nhất.`, userPreferences, viewHistory, availableTours)

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: userPrompt,
		},
	}

	req := openai.ChatCompletionRequest{
		Model:       openai.GPT3Dot5Turbo,
		Messages:    messages,
		Temperature: 0.7,
		MaxTokens:   1500,
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to generate recommendation: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("AI trả về kết quả rỗng")
	}

	return resp.Choices[0].Message.Content, nil
}
