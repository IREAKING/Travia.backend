package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"travia.backend/api/models"
	"travia.backend/api/utils"
	db "travia.backend/db/sqlc"
)

// GetReviewByTourId godoc
// @Summary Lấy đánh giá tour
// @Description Lấy đánh giá tour theo ID tour
// @Tags Review
// @Accept json
// @Produce json
// @Param id path int true "ID tour"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /review/tour/{id} [get]
func (s *Server) GetReviewByTourId(c *gin.Context) {
    id, _ := strconv.Atoi(c.Param("id"))
    
    // SQLc lúc này trả về trực tiếp 1 object review (không phải slice)
    review, err := s.z.GetReviewByTourId(c, int32(id))
    if err != nil {
        // Nếu không tìm thấy record nào, SQLc sẽ trả về lỗi sql.ErrNoRows
        c.JSON(http.StatusNotFound, gin.H{"error": "Không có đánh giá nào cho tour này"})
        return
    }

    // Giải mã JSON để xử lý triệt để vấn đề Base64 và mảng hình ảnh
    var details interface{}
    _ = json.Unmarshal(review.ThongTinDanhGia, &details)

    c.JSON(http.StatusOK, gin.H{
        "data": gin.H{
            "thong_tin_danh_gia": details,
            "tong_so_danh_gia":   review.TongSoDanhGia,
            "diem_trung_binh":    review.DiemTrungBinh,
            "so_luong_5_sao":     review.SoLuong5Sao,
            "so_luong_4_sao":     review.SoLuong4Sao,
            "so_luong_3_sao":     review.SoLuong3Sao,
            "so_luong_2_sao":     review.SoLuong2Sao,
            "so_luong_1_sao":     review.SoLuong1Sao,
        },
    })
}

// CreateReview godoc
// @Summary Tạo đánh giá tour
// @Description Tạo đánh giá tour khi đã hoàn thành tour (booking có trạng thái hoan_thanh)
// @Tags Review
// @Accept json
// @Produce json
// @Param request body models.CreateReviewRequest true "Thông tin đánh giá"
// @Success 201 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 403 {object} gin.H
// @Failure 404 {object} gin.H
// @Failure 409 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /review/create [post]
// @Security BearerAuth
func (s *Server) CreateReview(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Lấy thông tin user từ JWT token
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "User not authenticated",
			"message": "Người dùng chưa đăng nhập",
		})
		return
	}
	jwtClaims, ok := claims.(*utils.JwtClams)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Invalid authentication claims",
			"message": "Thông tin xác thực không hợp lệ",
		})
		return
	}
	userUUID := jwtClaims.Id

	// Parse request body
	var req models.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"message": "Dữ liệu yêu cầu không hợp lệ: " + err.Error(),
		})
		return
	}

	// Kiểm tra booking có thể đánh giá: trang_thai_dat_cho = 'da_thanh_toan' VÀ trang_thai_khoi_hanh = 'hoan_thanh'
	bookingCheck, err := s.z.CheckBookingCompletedAndNotReviewed(ctx, db.CheckBookingCompletedAndNotReviewedParams{
		ID:          req.DatChoID,
		NguoiDungID: userUUID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Booking not found or cannot be reviewed",
			"message": "Không tìm thấy booking hoặc booking chưa đủ điều kiện để đánh giá (cần: đã thanh toán và tour đã hoàn thành)",
		})
		return
	}

	// Kiểm tra đã có review chưa (double check)
	if bookingCheck.DaCoReview {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Review already exists",
			"message": "Bạn đã đánh giá tour này rồi",
		})
		return
	}

	// Tạo review
	review, err := s.z.CreateReview(ctx, db.CreateReviewParams{
		TourID:         bookingCheck.TourID,
		NguoiDungID:    userUUID,
		DatChoID:       req.DatChoID,
		DiemDanhGia:    req.DiemDanhGia,
		TieuDe:         req.TieuDe,
		NoiDung:        req.NoiDung,
		HinhAnhDinhKem: req.HinhAnhDinhKem,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create review",
			"message": "Lỗi khi tạo đánh giá: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Đánh giá đã được tạo thành công",
		"data":    review,
	})
}

// CheckReviewStatus godoc
// @Summary Kiểm tra trạng thái đánh giá của booking
// @Description Kiểm tra xem booking đã có đánh giá chưa
// @Tags Review
// @Accept json
// @Produce json
// @Param dat_cho_id path int true "ID của booking"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /review/check/{dat_cho_id} [get]
// @Security BearerAuth
func (s *Server) CheckReviewStatus(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Lấy thông tin user từ JWT token
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "User not authenticated",
			"message": "Người dùng chưa đăng nhập",
		})
		return
	}
	jwtClaims, ok := claims.(*utils.JwtClams)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Invalid authentication claims",
			"message": "Thông tin xác thực không hợp lệ",
		})
		return
	}
	userUUID := jwtClaims.Id

	// Parse booking ID
	datChoID, err := strconv.Atoi(c.Param("dat_cho_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid booking ID",
			"message": "ID booking không hợp lệ",
		})
		return
	}

	// Kiểm tra đã có review chưa
	hasReview, err := s.z.CheckReviewExists(ctx, db.CheckReviewExistsParams{
		DatChoID:    int32(datChoID),
		NguoiDungID: userUUID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to check review status",
			"message": "Lỗi khi kiểm tra trạng thái đánh giá",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"dat_cho_id": datChoID,
			"has_review": hasReview,
		},
	})
}
