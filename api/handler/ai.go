package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"travia.backend/api/helpers"
	"travia.backend/api/utils"
	db "travia.backend/db/sqlc"
)

// Chatbot godoc
// @Summary Chat với AI chatbot
// @Description Chat với AI chatbot để được tư vấn về tours. Hỗ trợ lưu lịch sử chat.
// @Tags AI
// @Accept json
// @Produce json
// @Param message body object true "Chat message" SchemaExample({"message":"string","session_id":"string","include_tours":false})
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /ai/chatbot [post]
func (s *Server) Chatbot(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	var req struct {
		Message      string `json:"message"`
		SessionID    string `json:"session_id"`
		IncludeTours bool   `json:"include_tours"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"message": "Dữ liệu yêu cầu không hợp lệ: " + err.Error(),
		})
		return
	}

	if req.Message == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing message",
			"message": "Vui lòng nhập câu hỏi",
		})
		return
	}

	// Generate session ID nếu chưa có
	if req.SessionID == "" {
		req.SessionID = fmt.Sprintf("session_%d", time.Now().UnixNano())
	}

	// Lấy user ID nếu đã đăng nhập
	var userID *string
	claims, exists := c.Get("claims")
	if exists {
		if jwtClaims, ok := claims.(*utils.JwtClams); ok {
			userIDStr := jwtClaims.Id.String()
			userID = &userIDStr
		}
	}

	// Lấy lịch sử chat gần đây (10 câu gần nhất)
	var chatHistory []string
	var userUUID pgtype.UUID

	if userID != nil {
		// Lấy từ user ID
		if err := userUUID.Scan(*userID); err != nil {
			userUUID = pgtype.UUID{} // Reset nếu scan fail
		}
	}

	// Lấy chat history từ database
	recentChats, err := s.z.GetRecentChatHistory(ctx, db.GetRecentChatHistoryParams{
		Column1: userUUID,
		MaPhien: req.SessionID,
		Limit:   10,
	})

	if err == nil {
		// Xây dựng chat history từ database
		for _, chat := range recentChats {
			chatHistory = append(chatHistory, fmt.Sprintf("Q: %s\nA: %s", chat.CauHoi, chat.CauTraLoi))
		}
	}

	// Lấy danh sách tours nếu cần
	toursList := ""
	if req.IncludeTours {
		tours, err := s.z.GetAllTour(ctx, db.GetAllTourParams{
			Limit:  10,
			Offset: 0,
		})
		if err == nil {
			for i, tour := range tours {
				if i >= 5 { // Chỉ lấy 5 tours đầu
					break
				}
				var giaStr string
				if tour.GiaNguoiLon.Valid {
					floatVal, _ := tour.GiaNguoiLon.Float64Value()
					if floatVal.Valid {
						giaStr = fmt.Sprintf("%.0f", floatVal.Float64)
					} else {
						giaStr = "N/A"
					}
				} else {
					giaStr = "N/A"
				}
				donViTienTe := "VND"
				if tour.DonViTienTe != nil {
					donViTienTe = *tour.DonViTienTe
				}
				toursList += fmt.Sprintf("- %s (%d ngày %d đêm, giá: %s %s)\n",
					tour.TieuDe, tour.SoNgay, tour.SoDem, giaStr, donViTienTe)
			}
		}
	}

	// Generate response từ AI
	response, err := helpers.GenerateChatbotResponse(
		s.config.OpenAIConfig.APIKey,
		req.Message,
		chatHistory,
		toursList,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate response",
			"message": "Không thể tạo phản hồi: " + err.Error(),
		})
		return
	}

	// Lưu vào database (userUUID đã được set ở trên)
	if userID == nil {
		userUUID = pgtype.UUID{} // NULL nếu không có user
	}

	_, err = s.z.CreateChatHistory(ctx, db.CreateChatHistoryParams{
		NguoiDungID: userUUID,
		MaPhien:     req.SessionID,
		CauHoi:      req.Message,
		CauTraLoi:   response,
	})
	if err != nil {
		// Log error nhưng không fail request
		fmt.Printf("Failed to save chat history: %v\n", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Phản hồi thành công",
		"session_id": req.SessionID,
		"data": gin.H{
			"response": response,
		},
	})
}
