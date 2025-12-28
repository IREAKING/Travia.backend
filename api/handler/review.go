package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (s *Server) GetReviewByTourId(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tour ID"})
		return
	}
	reviews, err := s.z.GetReviewByTourId(c, int32(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": reviews})
}
