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

// GetMyNotifications godoc
// @Summary Lấy thông báo của người dùng
// @Description Lấy danh sách thông báo của người dùng đã đăng nhập
// @Tags Notification
// @Accept json
// @Produce json
// @Param limit query int false "Số lượng kết quả" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /notifications [get]
// @Security BearerAuth
func (s *Server) GetMyNotifications(c *gin.Context) {
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

	// Convert user UUID
	var userUUID pgtype.UUID
	if err := userUUID.Scan(jwtClaims.Id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"message": "ID người dùng không hợp lệ",
		})
		return
	}

	notifications, err := s.z.GetNotificationsByUser(ctx, db.GetNotificationsByUserParams{
		NguoiDungID: userUUID,
		Limit:       int32(limit),
		Offset:      int32(offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get notifications",
			"message": "Không thể lấy danh sách thông báo: " + err.Error(),
		})
		return
	}

	// Get unread count
	unreadCount, err := s.z.CountUnreadNotificationsByUser(ctx, userUUID)
	if err != nil {
		// Log error but don't fail
		unreadCount = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"data":         notifications,
		"unread_count": unreadCount,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
		},
	})
}

// GetUnreadNotifications godoc
// @Summary Lấy thông báo chưa đọc
// @Description Lấy danh sách thông báo chưa đọc của người dùng
// @Tags Notification
// @Accept json
// @Produce json
// @Param limit query int false "Số lượng kết quả" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /notifications/unread [get]
// @Security BearerAuth
func (s *Server) GetUnreadNotifications(c *gin.Context) {
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

	// Convert user UUID
	var userUUID pgtype.UUID
	if err := userUUID.Scan(jwtClaims.Id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"message": "ID người dùng không hợp lệ",
		})
		return
	}

	notifications, err := s.z.GetUnreadNotificationsByUser(ctx, db.GetUnreadNotificationsByUserParams{
		NguoiDungID: userUUID,
		Limit:       int32(limit),
		Offset:      int32(offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get unread notifications",
			"message": "Không thể lấy danh sách thông báo chưa đọc: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": notifications,
	})
}

// MarkNotificationAsRead godoc
// @Summary Đánh dấu thông báo đã đọc
// @Description Đánh dấu một thông báo là đã đọc
// @Tags Notification
// @Accept json
// @Produce json
// @Param id path int true "ID thông báo"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /notifications/:id/read [put]
// @Security BearerAuth
func (s *Server) MarkNotificationAsRead(c *gin.Context) {
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

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"message": "ID không hợp lệ",
		})
		return
	}

	// Check if notification exists and belongs to user
	notification, err := s.z.GetNotificationByID(ctx, int32(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Notification not found",
			"message": "Không tìm thấy thông báo",
		})
		return
	}

	// Verify ownership
	var userUUID pgtype.UUID
	if err := userUUID.Scan(jwtClaims.Id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"message": "ID người dùng không hợp lệ",
		})
		return
	}

	if !notification.NguoiDungID.Valid || notification.NguoiDungID.Bytes != userUUID.Bytes {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Forbidden",
			"message": "Bạn không có quyền truy cập thông báo này",
		})
		return
	}

	updatedNotification, err := s.z.MarkNotificationAsRead(ctx, int32(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to mark notification as read",
			"message": "Không thể đánh dấu thông báo: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Đánh dấu đã đọc thành công",
		"data":    updatedNotification,
	})
}

// MarkAllNotificationsAsRead godoc
// @Summary Đánh dấu tất cả thông báo đã đọc
// @Description Đánh dấu tất cả thông báo của người dùng là đã đọc
// @Tags Notification
// @Accept json
// @Produce json
// @Success 200 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /notifications/read-all [put]
// @Security BearerAuth
func (s *Server) MarkAllNotificationsAsRead(c *gin.Context) {
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

	// Convert user UUID
	var userUUID pgtype.UUID
	if err := userUUID.Scan(jwtClaims.Id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"message": "ID người dùng không hợp lệ",
		})
		return
	}

	err := s.z.MarkAllNotificationsAsRead(ctx, userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to mark all notifications as read",
			"message": "Không thể đánh dấu tất cả thông báo: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Đã đánh dấu tất cả thông báo là đã đọc",
	})
}

// GetNotificationCount godoc
// @Summary Lấy số lượng thông báo chưa đọc
// @Description Lấy số lượng thông báo chưa đọc của người dùng
// @Tags Notification
// @Accept json
// @Produce json
// @Success 200 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /notifications/count [get]
// @Security BearerAuth
func (s *Server) GetNotificationCount(c *gin.Context) {
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

	// Convert user UUID
	var userUUID pgtype.UUID
	if err := userUUID.Scan(jwtClaims.Id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"message": "ID người dùng không hợp lệ",
		})
		return
	}

	count, err := s.z.CountUnreadNotificationsByUser(ctx, userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to count notifications",
			"message": "Không thể đếm số lượng thông báo: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"unread_count": count,
	})
}

