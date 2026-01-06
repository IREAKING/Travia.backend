package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"travia.backend/api/utils"
	db "travia.backend/db/sqlc"
)

// CreateContact godoc
// @Summary Tạo liên hệ mới
// @Description Tạo liên hệ mới từ form contact (có thể có hoặc không có user đăng nhập)
// @Tags Contact
// @Accept json
// @Produce json
// @Param request body db.CreateContactParams true "Thông tin liên hệ"
// @Success 201 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /contact [post]
func (s *Server) CreateContact(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var req db.CreateContactParams
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"message": "Dữ liệu yêu cầu không hợp lệ: " + err.Error(),
		})
		return
	}

	// Validation
	if req.HoTen == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing required field",
			"message": "Vui lòng nhập họ tên",
		})
		return
	}

	if req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing required field",
			"message": "Vui lòng nhập email",
		})
		return
	}

	if req.TieuDe == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing required field",
			"message": "Vui lòng nhập tiêu đề",
		})
		return
	}

	if req.NoiDung == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing required field",
			"message": "Vui lòng nhập nội dung",
		})
		return
	}

	// Nếu có user đăng nhập, lấy user_id từ JWT token
	claims, exists := c.Get("claims")
	if exists {
		jwtClaims, ok := claims.(*utils.JwtClams)
		if ok {
			// Convert string UUID to pgtype.UUID
			var userUUID pgtype.UUID
			if err := userUUID.Scan(jwtClaims.Id); err == nil {
				req.NguoiDungID = userUUID
			}
		}
	}

	// Tạo liên hệ
	contact, err := s.z.CreateContact(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create contact",
			"message": "Không thể tạo liên hệ: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Gửi liên hệ thành công",
		"data":    contact,
	})
}

// GetAllContacts godoc
// @Summary Lấy danh sách liên hệ (Admin only)
// @Description Lấy danh sách tất cả liên hệ với phân trang
// @Tags Contact
// @Accept json
// @Produce json
// @Param limit query int false "Số lượng kết quả" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 403 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /contact [get]
// @Security BearerAuth
func (s *Server) GetAllContacts(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	contacts, err := s.z.GetAllContacts(ctx, db.GetAllContactsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get contacts",
			"message": "Không thể lấy danh sách liên hệ: " + err.Error(),
		})
		return
	}

	// Get total count
	total, err := s.z.CountContacts(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to count contacts",
			"message": "Không thể đếm số lượng liên hệ: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": contacts,
		"pagination": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// GetContactByID godoc
// @Summary Lấy chi tiết liên hệ (Admin only)
// @Description Lấy thông tin chi tiết của một liên hệ
// @Tags Contact
// @Accept json
// @Produce json
// @Param id path int true "ID liên hệ"
// @Success 200 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 403 {object} gin.H
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /contact/:id [get]
// @Security BearerAuth
func (s *Server) GetContactByID(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"message": "ID không hợp lệ",
		})
		return
	}

	contact, err := s.z.GetContactByID(ctx, int32(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Contact not found",
			"message": "Không tìm thấy liên hệ",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": contact,
	})
}

// UpdateContactStatus godoc
// @Summary Cập nhật trạng thái liên hệ (Admin only)
// @Description Cập nhật trạng thái và trạng thái đọc của liên hệ
// @Tags Contact
// @Accept json
// @Produce json
// @Param id path int true "ID liên hệ"
// @Param request body db.UpdateContactStatusParams true "Thông tin cập nhật"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 403 {object} gin.H
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /contact/:id/status [put]
// @Security BearerAuth
func (s *Server) UpdateContactStatus(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"message": "ID không hợp lệ",
		})
		return
	}

	var req db.UpdateContactStatusParams
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"message": "Dữ liệu yêu cầu không hợp lệ: " + err.Error(),
		})
		return
	}

	req.ID = int32(id)

	// Validate status
	validStatuses := map[string]bool{
		"moi":         true,
		"dang_xu_ly":  true,
		"da_phan_hoi": true,
		"da_dong":     true,
	}
	if !validStatuses[*req.TrangThai] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid status",
			"message": "Trạng thái không hợp lệ",
		})
		return
	}

	contact, err := s.z.UpdateContactStatus(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update contact status",
			"message": "Không thể cập nhật trạng thái: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cập nhật trạng thái thành công",
		"data":    contact,
	})
}

// GetContactsByStatus godoc
// @Summary Lấy liên hệ theo trạng thái (Admin only)
// @Description Lấy danh sách liên hệ theo trạng thái với phân trang
// @Tags Contact
// @Accept json
// @Produce json
// @Param status path string true "Trạng thái (moi, dang_xu_ly, da_phan_hoi, da_dong)"
// @Param limit query int false "Số lượng kết quả" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 403 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /contact/status/:status [get]
// @Security BearerAuth
func (s *Server) GetContactsByStatus(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	status := c.Param("status")

	// Validate status
	validStatuses := map[string]bool{
		"moi":         true,
		"dang_xu_ly":  true,
		"da_phan_hoi": true,
		"da_dong":     true,
	}
	if !validStatuses[status] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid status",
			"message": "Trạng thái không hợp lệ",
		})
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	contacts, err := s.z.GetContactsByStatus(ctx, db.GetContactsByStatusParams{
		TrangThai: &status,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get contacts",
			"message": "Không thể lấy danh sách liên hệ: " + err.Error(),
		})
		return
	}

	// Get total count by status
	total, err := s.z.CountContactsByStatus(ctx, &status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to count contacts",
			"message": "Không thể đếm số lượng liên hệ: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": contacts,
		"pagination": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// GetUnreadContacts godoc
// @Summary Lấy danh sách liên hệ chưa đọc (Admin only)
// @Description Lấy danh sách liên hệ chưa đọc với phân trang
// @Tags Contact
// @Accept json
// @Produce json
// @Param limit query int false "Số lượng kết quả" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 403 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /contact/unread [get]
// @Security BearerAuth
func (s *Server) GetUnreadContacts(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	contacts, err := s.z.GetUnreadContacts(ctx, db.GetUnreadContactsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get unread contacts",
			"message": "Không thể lấy danh sách liên hệ chưa đọc: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": contacts,
	})
}

// MarkContactAsRead godoc
// @Summary Đánh dấu liên hệ đã đọc (Admin only)
// @Description Đánh dấu một liên hệ là đã đọc
// @Tags Contact
// @Accept json
// @Produce json
// @Param id path int true "ID liên hệ"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 403 {object} gin.H
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /contact/:id/read [put]
// @Security BearerAuth
func (s *Server) MarkContactAsRead(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"message": "ID không hợp lệ",
		})
		return
	}

	contact, err := s.z.MarkContactAsRead(ctx, int32(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Contact not found",
			"message": "Không tìm thấy liên hệ",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Đánh dấu đã đọc thành công",
		"data":    contact,
	})
}
