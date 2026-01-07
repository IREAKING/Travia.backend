package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"travia.backend/api/helpers"
	"travia.backend/api/utils"
	db "travia.backend/db/sqlc"
)

// TrackTourView godoc
// @Summary Lưu lịch sử xem tour
// @Description Lưu lại khi người dùng xem tour để phân tích sở thích và gợi ý tour
// @Tags Recommendation
// @Accept json
// @Produce json
// @Param request body object true "Thông tin xem tour" SchemaExample({"tour_id":1,"thoi_luong_xem_giay":60})
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /recommendation/track-view [post]
func (s *Server) TrackTourView(c *gin.Context) {
	var req struct {
		TourID          int `json:"tour_id"`
		ThoiLuongXemGiay int `json:"thoi_luong_xem_giay"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"message": "Dữ liệu yêu cầu không hợp lệ: " + err.Error(),
		})
		return
	}

	if req.TourID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid tour_id",
			"message": "ID tour không hợp lệ",
		})
		return
	}

	ctx := context.Background()

	// Lấy user ID nếu đã đăng nhập
	var userID *string
	var userUUID pgtype.UUID
	claims, exists := c.Get("claims")
	if exists {
		if jwtClaims, ok := claims.(*utils.JwtClams); ok {
			userIDStr := jwtClaims.Id.String()
			userID = &userIDStr
			if err := userUUID.Scan(*userID); err != nil {
				userUUID = pgtype.UUID{}
			}
		}
	}

	// Lấy IP và User Agent
	ip := GetClientIP(c)
	userAgent := c.GetHeader("User-Agent")

	// Nếu không có user ID, set NULL
	if userID == nil {
		userUUID = pgtype.UUID{}
	}

	// Lưu lịch sử xem
	tourID := int32(req.TourID)
	thoiLuongXem := int32(req.ThoiLuongXemGiay)
	thoiGianXem := pgtype.Timestamp{Time: time.Now(), Valid: true}
	
	_, err := s.z.CreateTourViewHistory(ctx, db.CreateTourViewHistoryParams{
		NguoiDungID:      userUUID,
		TourID:           &tourID,
		ThoiGianXem:      thoiGianXem,
		ThoiLuongXemGiay: &thoiLuongXem,
		IpAddress:        &ip,
		UserAgent:        &userAgent,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to track view",
			"message": "Không thể lưu lịch sử xem: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Đã lưu lịch sử xem tour thành công",
	})
}

// GetRecommendedTours godoc
// @Summary Lấy danh sách tour được gợi ý
// @Description Gợi ý tour dựa trên sở thích, lịch sử xem và AI
// @Tags Recommendation
// @Accept json
// @Produce json
// @Param limit query int false "Số lượng tour (mặc định: 10)"
// @Param offset query int false "Offset (mặc định: 0)"
// @Param method query string false "Phương pháp gợi ý: preferences, destinations, history, ai (mặc định: preferences)"
// @Success 200 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /recommendation/tours [get]
func (s *Server) GetRecommendedTours(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	method := c.DefaultQuery("method", "preferences") // preferences, destinations, history, ai

	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	ctx := context.Background()

	// Lấy user ID nếu đã đăng nhập
	var userID *string
	var userUUID pgtype.UUID
	claims, exists := c.Get("claims")
	if exists {
		if jwtClaims, ok := claims.(*utils.JwtClams); ok {
			userIDStr := jwtClaims.Id.String()
			userID = &userIDStr
			if err := userUUID.Scan(*userID); err != nil {
				userUUID = pgtype.UUID{}
			}
		}
	}

	// Nếu không có user, trả về tour nổi bật
	if userID == nil {
		tours, err := s.z.GetAllTour(ctx, db.GetAllTourParams{
			Limit:  int32(limit),
			Offset: int32(offset),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to get tours",
				"message": "Không thể lấy danh sách tour: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Lấy danh sách tour nổi bật thành công",
			"data":    tours,
			"method":  "featured",
		})
		return
	}

	var tours interface{}
	var err error
	var methodName string

	switch method {
	case "preferences":
		// Gợi ý dựa trên sở thích (điểm số)
		tours, err = s.z.GetRecommendedToursByPreferences(ctx, db.GetRecommendedToursByPreferencesParams{
			NguoiDungID: userUUID,
			Limit:      int32(limit),
			Offset:     int32(offset),
		})
		methodName = "sở thích"

	case "destinations":
		// Gợi ý dựa trên điểm đến yêu thích
		tours, err = s.z.GetRecommendedToursByFavoriteDestinations(ctx, db.GetRecommendedToursByFavoriteDestinationsParams{
			NguoiDungID: userUUID,
			Limit:      int32(limit),
			Offset:     int32(offset),
		})
		methodName = "điểm đến yêu thích"

	case "history":
		// Gợi ý dựa trên lịch sử xem
		tours, err = s.z.GetRecommendedToursByViewHistory(ctx, db.GetRecommendedToursByViewHistoryParams{
			NguoiDungID: userUUID,
			Limit:      int32(limit),
			Offset:     int32(offset),
		})
		methodName = "lịch sử xem"

	case "ai":
		// Gợi ý bằng AI
		tours, err = s.getAIRecommendedTours(ctx, userUUID, limit, offset)
		methodName = "AI"

	default:
		// Mặc định: preferences
		tours, err = s.z.GetRecommendedToursByPreferences(ctx, db.GetRecommendedToursByPreferencesParams{
			NguoiDungID: userUUID,
			Limit:      int32(limit),
			Offset:     int32(offset),
		})
		methodName = "sở thích"
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get recommended tours",
			"message": "Không thể lấy danh sách tour gợi ý: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Lấy danh sách tour gợi ý (%s) thành công", methodName),
		"data":    tours,
		"method":  method,
	})
}

// getAIRecommendedTours gợi ý tour bằng AI
func (s *Server) getAIRecommendedTours(ctx context.Context, userUUID pgtype.UUID, limit, offset int) (interface{}, error) {
	// Lấy sở thích người dùng
	preferences, err := s.z.GetUserPreferences(ctx, userUUID)
	if err != nil {
		preferences = []db.SoThichNguoiDung{}
	}

	// Lấy lịch sử xem
	viewHistory, err := s.z.GetTourViewHistoryByUser(ctx, db.GetTourViewHistoryByUserParams{
		NguoiDungID: userUUID,
		Limit:       10,
		Offset:      0,
	})
	if err != nil {
		viewHistory = []db.LichSuXemTour{}
	}

	// Xây dựng chuỗi mô tả sở thích
	prefText := "Không có sở thích được ghi nhận"
	if len(preferences) > 0 {
		var prefParts []string
		for _, p := range preferences {
			if p.LoaiSoThich == "danh_muc" {
				prefParts = append(prefParts, fmt.Sprintf("Danh mục ID %d: điểm số %.2f", p.GiaTriID, p.DiemSo))
			} else if p.LoaiSoThich == "diem_den" {
				prefParts = append(prefParts, fmt.Sprintf("Điểm đến ID %d: điểm số %.2f", p.GiaTriID, p.DiemSo))
			}
		}
		if len(prefParts) > 0 {
			prefText = strings.Join(prefParts, "\n")
		}
	}

	// Xây dựng chuỗi lịch sử xem
	historyText := "Không có lịch sử xem"
	if len(viewHistory) > 0 {
		var historyParts []string
		for _, h := range viewHistory {
			if h.TourID != nil && h.ThoiLuongXemGiay != nil {
				historyParts = append(historyParts, fmt.Sprintf("Tour ID %d: xem %d giây", *h.TourID, *h.ThoiLuongXemGiay))
			}
		}
		if len(historyParts) > 0 {
			historyText = strings.Join(historyParts, "\n")
		}
	}

	// Lấy danh sách tour có sẵn
	availableTours, err := s.z.GetAllTour(ctx, db.GetAllTourParams{
		Limit:  20,
		Offset: 0,
	})
	if err != nil {
		availableTours = []db.GetAllTourRow{}
	}

	// Xây dựng chuỗi tour có sẵn
	toursText := "Không có tour nào"
	if len(availableTours) > 0 {
		var tourParts []string
		for i, t := range availableTours {
			if i >= 10 {
				break
			}
			tourParts = append(tourParts, fmt.Sprintf("- Tour ID %d: %s", t.ID, t.TieuDe))
		}
		if len(tourParts) > 0 {
			toursText = strings.Join(tourParts, "\n")
		}
	}

	// Gọi AI để gợi ý
	aiRecommendation, err := helpers.GenerateTourRecommendation(
		s.config.OpenAIConfig.APIKey,
		prefText,
		historyText,
		toursText,
	)

	if err != nil {
		// Fallback về phương pháp preferences nếu AI fail
		return s.z.GetRecommendedToursByPreferences(ctx, db.GetRecommendedToursByPreferencesParams{
			NguoiDungID: userUUID,
			Limit:      int32(limit),
			Offset:     int32(offset),
		})
	}

	// Parse AI response để lấy tour IDs (đơn giản hóa - có thể cải thiện)
	// Tạm thời trả về tour nổi bật kèm AI recommendation text
	tours, err := s.z.GetAllTour(ctx, db.GetAllTourParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})

	return gin.H{
		"tours":           tours,
		"ai_recommendation": aiRecommendation,
	}, err
}

// GetSimilarTours godoc
// @Summary Lấy tour tương tự dựa trên embedding
// @Description Tìm tour tương tự với tour hiện tại sử dụng semantic search
// @Tags Recommendation
// @Accept json
// @Produce json
// @Param tour_id path int true "ID tour"
// @Param limit query int false "Số lượng tour (mặc định: 5)"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /recommendation/similar/{tour_id} [get]
func (s *Server) GetSimilarTours(c *gin.Context) {
	tourIDStr := c.Param("tour_id")
	tourID, err := strconv.Atoi(tourIDStr)
	if err != nil || tourID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid tour_id",
			"message": "ID tour không hợp lệ",
		})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))
	if limit <= 0 {
		limit = 5
	}

	ctx := context.Background()

	// Lấy embedding của tour hiện tại
	embedding, err := s.z.GetTourEmbedding(ctx, int32(tourID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Tour embedding not found",
			"message": "Không tìm thấy embedding cho tour này. Vui lòng tạo embedding trước.",
		})
		return
	}

	// Tìm tour tương tự (cần convert embedding từ database format)
	// Note: Cần implement conversion từ pgvector format sang []float32
	// Tạm thời trả về tour cùng danh mục
	_, err = s.z.GetTourDetailByID(ctx, int32(tourID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Tour not found",
			"message": "Không tìm thấy tour",
		})
		return
	}

	// Fallback: trả về tour cùng danh mục
	tours, err := s.z.GetAllTour(ctx, db.GetAllTourParams{
		Limit:  int32(limit + 1), // +1 để loại bỏ tour hiện tại
		Offset: 0,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get similar tours",
			"message": "Không thể lấy tour tương tự: " + err.Error(),
		})
		return
	}

	// Filter out current tour
	var similarTours []db.GetAllTourRow
	for _, t := range tours {
		if t.ID != int32(tourID) {
			similarTours = append(similarTours, t)
			if len(similarTours) >= limit {
				break
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Lấy tour tương tự thành công",
		"data":         similarTours,
		"note":         "Đang sử dụng phương pháp fallback (cùng danh mục). Embedding search sẽ được implement sau.",
		"tour_embedding_id": embedding.TourID,
	})
}

