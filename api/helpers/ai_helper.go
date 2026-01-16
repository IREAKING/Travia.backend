package helpers

import (
	"context"
	"fmt"
	"regexp"
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

// GenerateBlogContent tạo nội dung blog với sự hỗ trợ của AI
// Hỗ trợ tạo tiêu đề, tóm tắt, và nội dung blog về du lịch
func GenerateBlogContent(apiKey string, topic string, blogType string, additionalContext string) (title string, summary string, content string, err error) {
	if apiKey == "" {
		return "", "", "", fmt.Errorf("OpenAI API key is not configured")
	}

	if topic == "" {
		return "", "", "", fmt.Errorf("topic cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client := openai.NewClient(apiKey)

	// System prompt cho viết blog du lịch
	systemPrompt := `Bạn là một chuyên gia viết blog du lịch chuyên nghiệp của công ty Travia.
Nhiệm vụ của bạn là tạo ra các bài viết blog chất lượng cao về du lịch, địa điểm, kinh nghiệm du lịch.

Yêu cầu:
- Viết bằng tiếng Việt, tự nhiên và hấp dẫn
- Nội dung phải chính xác, hữu ích và có giá trị cho người đọc
- Sử dụng ngôn ngữ thân thiện, dễ hiểu
- Cấu trúc bài viết rõ ràng với các đoạn văn hợp lý
- Thêm các mẹo, kinh nghiệm thực tế khi phù hợp
- Tạo tiêu đề hấp dẫn, tóm tắt ngắn gọn (100-150 từ), và nội dung đầy đủ (800-1500 từ)

Các loại blog bạn có thể viết:
- kinh_nghiem: Kinh nghiệm du lịch, tips, hướng dẫn
- dia_diem: Giới thiệu địa điểm, điểm đến du lịch
- huong_dan: Hướng dẫn chi tiết về tour, lịch trình
- tin_tuc: Tin tức du lịch, sự kiện, cập nhật
- review: Đánh giá tour, khách sạn, dịch vụ

Format trả về (JSON):
{
  "title": "Tiêu đề bài viết",
  "summary": "Tóm tắt ngắn gọn 100-150 từ",
  "content": "Nội dung đầy đủ của bài viết, chia thành các đoạn văn rõ ràng"
}`

	// Xây dựng user prompt
	userPrompt := fmt.Sprintf(`Hãy tạo một bài viết blog về chủ đề: "%s"

Loại blog: %s

%s

Hãy tạo:
1. Một tiêu đề hấp dẫn và SEO-friendly
2. Một đoạn tóm tắt ngắn gọn (100-150 từ) để thu hút người đọc
3. Nội dung đầy đủ (800-1500 từ) với cấu trúc rõ ràng, chia thành các phần hợp lý

Trả về kết quả dưới dạng JSON với format đã chỉ định.`, topic, blogType, additionalContext)

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
		Temperature: 0.8, // Tăng tính sáng tạo cho blog
		MaxTokens:   3000, // Tăng tokens cho nội dung dài
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate blog content: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", "", "", fmt.Errorf("AI trả về kết quả rỗng")
	}

	// Parse JSON response
	responseText := resp.Choices[0].Message.Content
	// Loại bỏ markdown code blocks nếu có
	responseText = strings.TrimSpace(responseText)
	if strings.HasPrefix(responseText, "```json") {
		responseText = strings.TrimPrefix(responseText, "```json")
		responseText = strings.TrimSuffix(responseText, "```")
	} else if strings.HasPrefix(responseText, "```") {
		responseText = strings.TrimPrefix(responseText, "```")
		responseText = strings.TrimSuffix(responseText, "```")
	}
	responseText = strings.TrimSpace(responseText)

	// Parse JSON (đơn giản hóa - có thể dùng json.Unmarshal nếu cần chính xác hơn)
	// Tìm các trường title, summary, content
	title = extractJSONField(responseText, "title")
	summary = extractJSONField(responseText, "summary")
	content = extractJSONField(responseText, "content")

	if title == "" || summary == "" || content == "" {
		// Nếu không parse được JSON, trả về toàn bộ response làm content
		// và tự tạo title, summary
		content = responseText
		title = topic
		// Lấy 150 ký tự đầu làm summary
		if len(content) > 150 {
			summary = content[:150] + "..."
		} else {
			summary = content
		}
	}

	return title, summary, content, nil
}

// extractJSONField trích xuất giá trị của một field từ JSON string (đơn giản hóa)
func extractJSONField(jsonStr, fieldName string) string {
	// Tìm field trong JSON
	startIdx := strings.Index(jsonStr, fmt.Sprintf(`"%s":`, fieldName))
	if startIdx == -1 {
		return ""
	}

	// Tìm dấu nháy kép sau dấu hai chấm
	valueStart := strings.Index(jsonStr[startIdx:], `"`)
	if valueStart == -1 {
		return ""
	}
	valueStart += startIdx + 1

	// Tìm dấu nháy kép kết thúc
	valueEnd := strings.Index(jsonStr[valueStart:], `"`)
	if valueEnd == -1 {
		return ""
	}

	value := jsonStr[valueStart : valueStart+valueEnd]
	// Unescape JSON strings
	value = strings.ReplaceAll(value, `\"`, `"`)
	value = strings.ReplaceAll(value, `\\`, `\`)
	value = strings.ReplaceAll(value, `\n`, "\n")
	value = strings.ReplaceAll(value, `\t`, "\t")

	return value
}

// GenerateBlogTitleSuggestions tạo các gợi ý tiêu đề blog
func GenerateBlogTitleSuggestions(apiKey string, topic string, count int) ([]string, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key is not configured")
	}

	if topic == "" {
		return nil, fmt.Errorf("topic cannot be empty")
	}

	if count <= 0 || count > 10 {
		count = 5 // Default 5 suggestions
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := openai.NewClient(apiKey)

	systemPrompt := `Bạn là chuyên gia tạo tiêu đề blog du lịch.
Nhiệm vụ của bạn là tạo ra các tiêu đề hấp dẫn, SEO-friendly cho blog du lịch.

Yêu cầu:
- Tiêu đề phải hấp dẫn, thu hút người đọc
- Tối ưu SEO với từ khóa phù hợp
- Độ dài 50-70 ký tự (tối ưu cho SEO)
- Viết bằng tiếng Việt
- Mỗi tiêu đề phải khác biệt và độc đáo`

	userPrompt := fmt.Sprintf(`Hãy tạo %d tiêu đề blog hấp dẫn về chủ đề: "%s"

Trả về danh sách tiêu đề, mỗi tiêu đề trên một dòng.`, count, topic)

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
		Temperature: 0.9, // Tăng tính sáng tạo
		MaxTokens:   500,
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate title suggestions: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("AI trả về kết quả rỗng")
	}

	// Parse response thành danh sách tiêu đề
	responseText := resp.Choices[0].Message.Content
	lines := strings.Split(responseText, "\n")
	
	var titles []string
	numberRegex := regexp.MustCompile(`^\d+\.`)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Loại bỏ số thứ tự nếu có (1., 2., - , etc.)
		line = strings.TrimPrefix(line, "-")
		line = strings.TrimSpace(line)
		line = numberRegex.ReplaceAllString(line, "")
		line = strings.TrimSpace(line)
		if line != "" && len(line) > 10 {
			titles = append(titles, line)
		}
	}

	// Giới hạn số lượng
	if len(titles) > count {
		titles = titles[:count]
	}

	return titles, nil
}
