package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
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

// CreateContactResponse godoc
// @Summary Tạo phản hồi cho liên hệ (Admin only)
// @Description Admin phản hồi lại liên hệ từ khách hàng
// @Tags Contact
// @Accept json
// @Produce json
// @Param id path int true "ID liên hệ"
// @Param request body models.ResponseContact_Swagger true "Nội dung phản hồi"
// @Success 201 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 403 {object} gin.H
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Security BearerAuth
// @Router /contact/:id/response [post]
func (s *Server) CreateContactResponse(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Get user from JWT
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	jwtClaims, ok := claims.(*utils.JwtClams)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authentication claims"})
		return
	}

	contactID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"message": "ID không hợp lệ",
		})
		return
	}

	// Check if contact exists
	contact, err := s.z.GetContactByID(ctx, int32(contactID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Contact not found",
			"message": "Không tìm thấy liên hệ",
		})
		return
	}

	var req struct {
		NoiDung string `json:"noi_dung" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("CreateContactResponse - Bind error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"message": "Dữ liệu yêu cầu không hợp lệ: " + err.Error(),
			"details": err.Error(),
		})
		return
	}

	if req.NoiDung == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing required field",
			"message": "Vui lòng nhập nội dung phản hồi",
		})
		return
	}

	fmt.Printf("CreateContactResponse - ContactID: %d, NoiDung length: %d\n", contactID, len(req.NoiDung))

	// Create response
	response, err := s.z.CreateContactResponse(ctx, db.CreateContactResponseParams{
		LienHeID:       int32(contactID),
		NguoiPhanHoiID: jwtClaims.Id,
		NoiDung:        req.NoiDung,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create response",
			"message": "Không thể tạo phản hồi: " + err.Error(),
		})
		return
	}

	// Update contact status to 'da_phan_hoi'
	_, err = s.z.UpdateContactStatus(ctx, db.UpdateContactStatusParams{
		ID:        int32(contactID),
		TrangThai: stringPtr("da_phan_hoi"),
		DaDoc:     contact.DaDoc,
	})
	if err != nil {
		// Log error but don't fail the response creation
		fmt.Printf("Warning: Failed to update contact status: %v\n", err)
	}

	// Create notification if contact has user_id
	if contact.NguoiDungID.Valid {
		notification, err := s.z.CreateNotificationForContactResponse(ctx, db.CreateNotificationForContactResponseParams{
			LienHeID:  int32(contactID),
			PhanHoiID: response.ID,
		})
		if err != nil {
			// Check if it's a "no rows" error (contact might not have user_id or JOIN failed)
			if errors.Is(err, pgx.ErrNoRows) {
				// This is OK - contact might not have a registered user
				fmt.Printf("Info: No notification created (contact may not have registered user): %v\n", err)
			} else {
				// Other errors - log but don't fail
				fmt.Printf("Warning: Failed to create notification: %v\n", err)
			}
		} else {
			fmt.Printf("Info: Notification created successfully: ID %d\n", notification.ID)
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Phản hồi đã được gửi thành công",
		"data":    response,
		"success": true,
	})
}

// GetContactResponses godoc
// @Summary Lấy danh sách phản hồi của liên hệ (Admin only)
// @Description Lấy tất cả phản hồi của một liên hệ
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
// @Router /contact/:id/responses [get]
// @Security BearerAuth
func (s *Server) GetContactResponses(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	contactID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"message": "ID không hợp lệ",
		})
		return
	}

	// Check if contact exists
	_, err = s.z.GetContactByID(ctx, int32(contactID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Contact not found",
			"message": "Không tìm thấy liên hệ",
		})
		return
	}

	responses, err := s.z.GetContactResponses(ctx, int32(contactID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get responses",
			"message": "Không thể lấy danh sách phản hồi: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": responses,
	})
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
