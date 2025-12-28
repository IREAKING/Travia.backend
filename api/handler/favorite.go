package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"travia.backend/api/utils"
	db "travia.backend/db/sqlc"
)
// Thêm tour yêu thích
// @Summary Thêm tour yêu thích
// @Description Thêm tour yêu thích
// @Tags Favorite
// @Accept json
// @Produce json
// @Param favorite body db.CreateFavoriteTourParams true "Favorite"
// @Security ApiKeyAuth
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /favorite/ [post]
func (s *Server) CreateFavoriteTour(c *gin.Context) {
	var req db.CreateFavoriteTourParams
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
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
	userUUID := jwtClaims.Id
	favoriteTour, err := s.z.CreateFavoriteTour(c.Request.Context(), db.CreateFavoriteTourParams{
		NguoiDungID: userUUID,
		TourID:      req.TourID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create favorite tour"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Favorite tour created successfully", "favoriteTour": favoriteTour})
}

// Xóa tour yêu thích
// @Summary Xóa tour yêu thích
// @Description Xóa tour yêu thích
// @Tags Favorite
// @Accept json
// @Produce json
// @Param favorite body db.DeleteFavoriteTourParams true "Favorite"
// @Security ApiKeyAuth
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /favorite/ [delete]
func (s *Server) DeleteFavoriteTour(c *gin.Context) {
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
	userUUID := jwtClaims.Id
	var req db.DeleteFavoriteTourParams
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = s.z.DeleteFavoriteTour(c.Request.Context(), db.DeleteFavoriteTourParams{
		NguoiDungID: userUUID,
		TourID:      req.TourID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete favorite tour"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Favorite tour deleted successfully"})
}

// Lấy danh sách tour yêu thích
// @Summary Lấy danh sách tour yêu thích
// @Description Lấy danh sách tour yêu thích
// @Tags Favorite
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /favorite/ [get]
func (s *Server) GetFavoriteTours(c *gin.Context) {
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
	userUUID := jwtClaims.Id
	favoriteTours, err := s.z.GetFavoriteTours(c.Request.Context(), userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get favorite tours"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Favorite tours fetched successfully", "favoriteTours": favoriteTours})
}
