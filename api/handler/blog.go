package handler

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"travia.backend/api/helpers"
	"travia.backend/api/utils"
	db "travia.backend/db/sqlc"
)

// ========== PUBLIC BLOG ENDPOINTS ==========

// GetPublishedBlogs godoc
// @Summary Lấy danh sách blog đã công bố
// @Description Lấy danh sách blog đã công bố với pagination
// @Tags Blog
// @Accept json
// @Produce json
// @Param limit query int false "Số lượng blog mỗi trang" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} gin.H
// @Router /blog/posts [get]
func (s *Server) GetPublishedBlogs(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	blogs, err := s.z.GetPublishedBlogs(ctx, db.GetPublishedBlogsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		helpers.InternalServerError(c, "Lỗi khi lấy danh sách blog", err)
		return
	}

	helpers.Ok(c, blogs)
}

// GetBlogBySlug godoc
// @Summary Lấy blog theo slug
// @Description Lấy chi tiết blog theo slug
// @Tags Blog
// @Accept json
// @Produce json
// @Param slug path string true "Slug của blog"
// @Success 200 {object} gin.H
// @Router /blog/posts/:slug [get]
func (s *Server) GetBlogBySlug(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	slug := c.Param("slug")
	blog, err := s.z.GetBlogBySlug(ctx, slug)
	if err != nil {
		if err == sql.ErrNoRows {
			helpers.NotFound(c, "Không tìm thấy blog")
			return
		}
		helpers.InternalServerError(c, "Lỗi khi lấy blog", err)
		return
	}

	// Tăng lượt xem
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.z.IncrementBlogViews(ctx, blog.ID)
	}()

	helpers.Ok(c, blog)
}

// GetFeaturedBlogs godoc
// @Summary Lấy danh sách blog nổi bật
// @Description Lấy danh sách blog nổi bật
// @Tags Blog
// @Accept json
// @Produce json
// @Param limit query int false "Số lượng blog" default(5)
// @Success 200 {object} gin.H
// @Router /blog/featured [get]
func (s *Server) GetFeaturedBlogs(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))
	blogs, err := s.z.GetFeaturedBlogs(ctx, int32(limit))
	if err != nil {
		helpers.InternalServerError(c, "Lỗi khi lấy blog nổi bật", err)
		return
	}

	helpers.Ok(c, blogs)
}

// SearchBlogs godoc
// @Summary Tìm kiếm blog
// @Description Tìm kiếm blog theo từ khóa
// @Tags Blog
// @Accept json
// @Produce json
// @Param q query string true "Từ khóa tìm kiếm"
// @Param limit query int false "Số lượng blog" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} gin.H
// @Router /blog/search [get]
func (s *Server) SearchBlogs(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	query := c.Query("q")
	if query == "" {
		helpers.BadRequest(c, "Từ khóa tìm kiếm không được để trống", nil)
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	blogs, err := s.z.SearchBlogs(ctx, db.SearchBlogsParams{
		PlaintoTsquery: query,
		Limit:          int32(limit),
		Offset:         int32(offset),
	})
	if err != nil {
		helpers.InternalServerError(c, "Lỗi khi tìm kiếm blog", err)
		return
	}

	helpers.Ok(c, blogs)
}

// GetBlogsByCategory godoc
// @Summary Lấy blog theo danh mục
// @Description Lấy danh sách blog theo danh mục
// @Tags Blog
// @Accept json
// @Produce json
// @Param category path string true "Danh mục"
// @Param limit query int false "Số lượng blog" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} gin.H
// @Router /blog/category/:category [get]
func (s *Server) GetBlogsByCategory(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	category := c.Param("category")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	categoryPtr := &category
	blogs, err := s.z.GetBlogsByCategory(ctx, db.GetBlogsByCategoryParams{
		DanhMuc: categoryPtr,
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		helpers.InternalServerError(c, "Lỗi khi lấy blog theo danh mục", err)
		return
	}

	helpers.Ok(c, blogs)
}

// IncrementBlogViews godoc
// @Summary Tăng lượt xem blog
// @Description Tăng lượt xem blog
// @Tags Blog
// @Accept json
// @Produce json
// @Param id path int true "ID blog"
// @Success 200 {object} gin.H
// @Router /blog/posts/:id/view [post]
func (s *Server) IncrementBlogViews(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		helpers.BadRequest(c, "ID không hợp lệ", nil)
		return
	}

	err = s.z.IncrementBlogViews(ctx, int32(id))
	if err != nil {
		helpers.InternalServerError(c, "Lỗi khi tăng lượt xem", err)
		return
	}

	helpers.Ok(c, nil)
}

// ========== ADMIN BLOG ENDPOINTS ==========

// GetAllBlogsForAdmin godoc
// @Summary Lấy tất cả blog cho admin
// @Description Lấy tất cả blog với filter và pagination
// @Tags Blog Admin
// @Accept json
// @Produce json
// @Param limit query int false "Số lượng blog" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} gin.H
// @Security ApiKeyAuth
// @Router /admin/blog [get]
func (s *Server) GetAllBlogsForAdmin(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	blogs, err := s.z.GetAllBlogs(ctx, db.GetAllBlogsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		helpers.InternalServerError(c, "Lỗi khi lấy danh sách blog", err)
		return
	}

	helpers.Ok(c, blogs)
}

// GetBlogByIDForAdmin godoc
// @Summary Lấy blog theo ID cho admin
// @Description Lấy chi tiết blog theo ID
// @Tags Blog Admin
// @Accept json
// @Produce json
// @Param id path int true "ID blog"
// @Success 200 {object} gin.H
// @Security ApiKeyAuth
// @Router /admin/blog/:id [get]
func (s *Server) GetBlogByIDForAdmin(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		helpers.BadRequest(c, "ID không hợp lệ", nil)
		return
	}

	blog, err := s.z.GetBlogByID(ctx, int32(id))
	if err != nil {
		if err == sql.ErrNoRows {
			helpers.NotFound(c, "Không tìm thấy blog")
			return
		}
		helpers.InternalServerError(c, "Lỗi khi lấy blog", err)
		return
	}

	helpers.Ok(c, blog)
}

// CreateBlog godoc
// @Summary Tạo blog mới
// @Description Tạo blog mới
// @Tags Blog Admin
// @Accept json
// @Produce json
// @Param blog body CreateBlogRequest true "Thông tin blog"
// @Success 200 {object} gin.H
// @Security ApiKeyAuth
// @Router /admin/blog [post]
func (s *Server) CreateBlog(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var req CreateBlogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.BadRequest(c, "Dữ liệu không hợp lệ", err)
		return
	}

	// Lấy user ID từ JWT
	jwtClaims, exists := c.Get("jwtClaims")
	if !exists {
		helpers.Unauthorized(c, "Không có quyền truy cập", nil)
		return
	}
	claims := jwtClaims.(*utils.JwtClams)
	authorID := claims.Id

	// Tạo slug từ tiêu đề nếu chưa có
	slug := req.Slug
	if slug == "" {
		slug = utils.GenerateSlug(req.TieuDe)
	}

	// Parse ngay_dang nếu có
	var ngayDang pgtype.Timestamp
	if req.NgayDang != "" {
		t, err := time.Parse("2006-01-02T15:04:05Z07:00", req.NgayDang)
		if err != nil {
			// Thử format khác
			t, err = time.Parse("2006-01-02", req.NgayDang)
			if err != nil {
				helpers.BadRequest(c, "Định dạng ngày không hợp lệ", err)
				return
			}
		}
		ngayDang = pgtype.Timestamp{Time: t, Valid: true}
	}

	// Parse các trường optional
	var tomTat *string
	if req.TomTat != "" {
		tomTat = &req.TomTat
	}

	var anhDaiDien *string
	if req.AnhDaiDien != "" {
		anhDaiDien = &req.AnhDaiDien
	}

	var danhMuc *string
	if req.DanhMuc != "" {
		danhMuc = &req.DanhMuc
	}

	var trangThai *string
	if req.TrangThai != "" {
		trangThai = &req.TrangThai
	}

	noiBat := &req.NoiBat

	blog, err := s.z.CreateBlog(ctx, db.CreateBlogParams{
		TieuDe:     req.TieuDe,
		Slug:       slug,
		TomTat:     tomTat,
		NoiDung:    req.NoiDung,
		AnhDaiDien: anhDaiDien,
		TacGiaID:   authorID,
		DanhMuc:    danhMuc,
		TuKhoa:     req.TuKhoa,
		TrangThai:  trangThai,
		NoiBat:     noiBat,
		NgayDang:   ngayDang,
	})
	if err != nil {
		helpers.InternalServerError(c, "Lỗi khi tạo blog", err)
		return
	}

	helpers.Ok(c, gin.H{
		"message": "Tạo blog thành công",
		"data":    blog,
	})
}

// UpdateBlog godoc
// @Summary Cập nhật blog
// @Description Cập nhật blog
// @Tags Blog Admin
// @Accept json
// @Produce json
// @Param id path int true "ID blog"
// @Param blog body UpdateBlogRequest true "Thông tin blog"
// @Success 200 {object} gin.H
// @Security ApiKeyAuth
// @Router /admin/blog/:id [put]
func (s *Server) UpdateBlog(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		helpers.BadRequest(c, "ID không hợp lệ", nil)
		return
	}

	var req UpdateBlogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.BadRequest(c, "Dữ liệu không hợp lệ", err)
		return
	}

	// Lấy blog hiện tại để giữ giá trị cũ nếu không cập nhật
	currentBlog, err := s.z.GetBlogByID(ctx, int32(id))
	if err != nil {
		if err == sql.ErrNoRows {
			helpers.NotFound(c, "Không tìm thấy blog")
			return
		}
		helpers.InternalServerError(c, "Lỗi khi lấy blog", err)
		return
	}

	// Parse các trường, giữ giá trị cũ nếu không có giá trị mới
	tieuDe := currentBlog.TieuDe
	if req.TieuDe != nil {
		tieuDe = *req.TieuDe
	}

	slug := currentBlog.Slug
	if req.Slug != nil {
		slug = *req.Slug
	} else if req.TieuDe != nil {
		slug = utils.GenerateSlug(*req.TieuDe)
	}

	var tomTat *string
	if req.TomTat != nil {
		tomTat = req.TomTat
	} else if currentBlog.TomTat != nil {
		tomTat = currentBlog.TomTat
	}

	noiDung := currentBlog.NoiDung
	if req.NoiDung != nil {
		noiDung = *req.NoiDung
	}

	var anhDaiDien *string
	if req.AnhDaiDien != nil {
		anhDaiDien = req.AnhDaiDien
	} else if currentBlog.AnhDaiDien != nil {
		anhDaiDien = currentBlog.AnhDaiDien
	}

	var danhMuc *string
	if req.DanhMuc != nil {
		danhMuc = req.DanhMuc
	} else if currentBlog.DanhMuc != nil {
		danhMuc = currentBlog.DanhMuc
	}

	tuKhoa := currentBlog.TuKhoa
	if req.TuKhoa != nil {
		tuKhoa = *req.TuKhoa
	}

	var trangThai *string
	if req.TrangThai != nil {
		trangThai = req.TrangThai
	} else if currentBlog.TrangThai != nil {
		trangThai = currentBlog.TrangThai
	}

	var noiBat *bool
	if req.NoiBat != nil {
		noiBat = req.NoiBat
	} else if currentBlog.NoiBat != nil {
		noiBat = currentBlog.NoiBat
	}

	ngayDang := currentBlog.NgayDang
	if req.NgayDang != nil && *req.NgayDang != "" {
		t, err := time.Parse("2006-01-02T15:04:05Z07:00", *req.NgayDang)
		if err != nil {
			t, err = time.Parse("2006-01-02", *req.NgayDang)
			if err != nil {
				helpers.BadRequest(c, "Định dạng ngày không hợp lệ", err)
				return
			}
		}
		ngayDang = pgtype.Timestamp{Time: t, Valid: true}
	}

	blog, err := s.z.UpdateBlog(ctx, db.UpdateBlogParams{
		ID:         int32(id),
		TieuDe:     tieuDe,
		Slug:       slug,
		TomTat:     tomTat,
		NoiDung:    noiDung,
		AnhDaiDien: anhDaiDien,
		DanhMuc:    danhMuc,
		TuKhoa:     tuKhoa,
		TrangThai:  trangThai,
		NoiBat:     noiBat,
		NgayDang:   ngayDang,
	})
	if err != nil {
		helpers.InternalServerError(c, "Lỗi khi cập nhật blog", err)
		return
	}

	helpers.Ok(c, gin.H{
		"message": "Cập nhật blog thành công",
		"data":    blog,
	})
}

// DeleteBlog godoc
// @Summary Xóa blog
// @Description Xóa blog
// @Tags Blog Admin
// @Accept json
// @Produce json
// @Param id path int true "ID blog"
// @Success 200 {object} gin.H
// @Security ApiKeyAuth
// @Router /admin/blog/:id [delete]
func (s *Server) DeleteBlog(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		helpers.BadRequest(c, "ID không hợp lệ", nil)
		return
	}

	err = s.z.DeleteBlog(ctx, int32(id))
	if err != nil {
		helpers.InternalServerError(c, "Lỗi khi xóa blog", err)
		return
	}

	helpers.Ok(c, gin.H{
		"message": "Xóa blog thành công",
	})
}

// GetBlogStats godoc
// @Summary Lấy thống kê blog
// @Description Lấy thống kê blog
// @Tags Blog Admin
// @Accept json
// @Produce json
// @Success 200 {object} gin.H
// @Security ApiKeyAuth
// @Router /admin/blog/stats [get]
func (s *Server) GetBlogStats(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	stats, err := s.z.GetBlogStats(ctx)
	if err != nil {
		helpers.InternalServerError(c, "Lỗi khi lấy thống kê blog", err)
		return
	}

	helpers.Ok(c, gin.H{
		"message": "Lấy thống kê blog thành công",
		"data":    stats,
	})
}

// ========== AI BLOG ENDPOINTS ==========

// GenerateBlogContent godoc
// @Summary Tạo nội dung blog bằng AI
// @Description Tạo tiêu đề, tóm tắt và nội dung blog bằng AI
// @Tags Blog Admin AI
// @Accept json
// @Produce json
// @Param request body GenerateBlogContentRequest true "Thông tin tạo blog"
// @Success 200 {object} gin.H
// @Security ApiKeyAuth
// @Router /admin/blog/ai/generate [post]
func (s *Server) GenerateBlogContent(c *gin.Context) {
	var req GenerateBlogContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.BadRequest(c, "Dữ liệu không hợp lệ", err)
		return
	}

	if req.Topic == "" {
		helpers.BadRequest(c, "Chủ đề không được để trống", nil)
		return
	}

	// Lấy OpenAI API key từ config
	apiKey := s.config.OpenAIConfig.APIKey
	if apiKey == "" {
		helpers.InternalServerError(c, "OpenAI API key chưa được cấu hình", nil)
		return
	}

	// Gọi AI helper để tạo nội dung
	title, summary, content, err := helpers.GenerateBlogContent(
		apiKey,
		req.Topic,
		req.BlogType,
		req.AdditionalContext,
	)
	if err != nil {
		helpers.InternalServerError(c, "Lỗi khi tạo nội dung blog bằng AI", err)
		return
	}

	// Lấy thông tin tokens từ OpenAI response (cần parse từ response)
	// Tạm thời trả về kết quả
	helpers.Ok(c, gin.H{
		"message": "Tạo nội dung blog thành công",
		"data": gin.H{
			"title":   title,
			"summary": summary,
			"content": content,
		},
	})
}

// GenerateBlogTitleSuggestions godoc
// @Summary Tạo gợi ý tiêu đề blog bằng AI
// @Description Tạo nhiều gợi ý tiêu đề blog SEO-friendly
// @Tags Blog Admin AI
// @Accept json
// @Produce json
// @Param request body GenerateTitleSuggestionsRequest true "Thông tin tạo tiêu đề"
// @Success 200 {object} gin.H
// @Security ApiKeyAuth
// @Router /admin/blog/ai/titles [post]
func (s *Server) GenerateBlogTitleSuggestions(c *gin.Context) {
	var req GenerateTitleSuggestionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.BadRequest(c, "Dữ liệu không hợp lệ", err)
		return
	}

	if req.Topic == "" {
		helpers.BadRequest(c, "Chủ đề không được để trống", nil)
		return
	}

	// Lấy OpenAI API key từ config
	apiKey := s.config.OpenAIConfig.APIKey
	if apiKey == "" {
		helpers.InternalServerError(c, "OpenAI API key chưa được cấu hình", nil)
		return
	}

	count := req.Count
	if count <= 0 || count > 10 {
		count = 5
	}

	// Gọi AI helper để tạo gợi ý tiêu đề
	titles, err := helpers.GenerateBlogTitleSuggestions(apiKey, req.Topic, count)
	if err != nil {
		helpers.InternalServerError(c, "Lỗi khi tạo gợi ý tiêu đề", err)
		return
	}

	helpers.Ok(c, gin.H{
		"message": "Tạo gợi ý tiêu đề thành công",
		"data": gin.H{
			"titles": titles,
		},
	})
}

// CreateBlogWithAI godoc
// @Summary Tạo blog với AI và lưu vào database
// @Description Tạo blog với AI, lưu vào database và lưu lịch sử AI
// @Tags Blog Admin AI
// @Accept json
// @Produce json
// @Param request body CreateBlogWithAIRequest true "Thông tin tạo blog với AI"
// @Success 200 {object} gin.H
// @Security ApiKeyAuth
// @Router /admin/blog/ai/create [post]
func (s *Server) CreateBlogWithAI(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	var req CreateBlogWithAIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.BadRequest(c, "Dữ liệu không hợp lệ", err)
		return
	}

	if req.Topic == "" {
		helpers.BadRequest(c, "Chủ đề không được để trống", nil)
		return
	}

	// Lấy user ID từ JWT
	jwtClaims, exists := c.Get("jwtClaims")
	if !exists {
		helpers.Unauthorized(c, "Không có quyền truy cập", nil)
		return
	}
	claims := jwtClaims.(*utils.JwtClams)
	authorID := claims.Id

	// Lấy OpenAI API key từ config
	apiKey := s.config.OpenAIConfig.APIKey
	if apiKey == "" {
		helpers.InternalServerError(c, "OpenAI API key chưa được cấu hình", nil)
		return
	}

	// Gọi AI helper để tạo nội dung
	title, summary, content, err := helpers.GenerateBlogContent(
		apiKey,
		req.Topic,
		req.BlogType,
		req.AdditionalContext,
	)
	if err != nil {
		helpers.InternalServerError(c, "Lỗi khi tạo nội dung blog bằng AI", err)
		return
	}

	// Tạo slug từ tiêu đề
	slug := utils.GenerateSlug(title)

	// Parse các trường optional
	var tomTat *string
	if summary != "" {
		tomTat = &summary
	}

	var anhDaiDien *string
	if req.AnhDaiDien != "" {
		anhDaiDien = &req.AnhDaiDien
	}

	var danhMuc *string
	if req.DanhMuc != "" {
		danhMuc = &req.DanhMuc
	}

	var trangThai *string
	if req.TrangThai != "" {
		trangThai = &req.TrangThai
	}

	noiBat := &req.NoiBat

	// Parse ngay_dang nếu có
	var ngayDang pgtype.Timestamp
	if req.NgayDang != "" {
		t, err := time.Parse("2006-01-02T15:04:05Z07:00", req.NgayDang)
		if err != nil {
			t, err = time.Parse("2006-01-02", req.NgayDang)
			if err != nil {
				helpers.BadRequest(c, "Định dạng ngày không hợp lệ", err)
				return
			}
		}
		ngayDang = pgtype.Timestamp{Time: t, Valid: true}
	}

	// Tạo blog trong database
	blog, err := s.z.CreateBlog(ctx, db.CreateBlogParams{
		TieuDe:     title,
		Slug:       slug,
		TomTat:     tomTat,
		NoiDung:    content,
		AnhDaiDien: anhDaiDien,
		TacGiaID:   authorID,
		DanhMuc:    danhMuc,
		TuKhoa:     req.TuKhoa,
		TrangThai:  trangThai,
		NoiBat:     noiBat,
		NgayDang:   ngayDang,
	})
	if err != nil {
		helpers.InternalServerError(c, "Lỗi khi tạo blog", err)
		return
	}

	// Lưu lịch sử AI (tạm thời không có thông tin tokens từ response)
	// Cần cải thiện để lấy thông tin tokens từ OpenAI response
	prompt := fmt.Sprintf("Tạo blog về chủ đề: %s, loại: %s", req.Topic, req.BlogType)
	if req.AdditionalContext != "" {
		prompt += fmt.Sprintf(", thông tin thêm: %s", req.AdditionalContext)
	}

	phanHoiAi := fmt.Sprintf("Title: %s\nSummary: %s\nContent: %s", title, summary, content)
	moHinhAi := "gpt-3.5-turbo"
	_, err = s.z.CreateBlogAIHistory(ctx, db.CreateBlogAIHistoryParams{
		BlogID:       &blog.ID,
		Prompt:       prompt,
		PhanHoiAi:    &phanHoiAi,
		MoHinhAi:     &moHinhAi,
		SoLuongToken: nil, // Cần cải thiện để lấy từ response
	})
	if err != nil {
		// Log lỗi nhưng không fail request
		fmt.Printf("Lỗi khi lưu lịch sử AI: %v\n", err)
	}

	helpers.Ok(c, gin.H{
		"message": "Tạo blog với AI thành công",
		"data":    blog,
	})
}

// GetBlogAIHistory godoc
// @Summary Lấy lịch sử AI của blog
// @Description Lấy lịch sử sử dụng AI để tạo blog
// @Tags Blog Admin AI
// @Accept json
// @Produce json
// @Param id path int true "ID blog"
// @Success 200 {object} gin.H
// @Security ApiKeyAuth
// @Router /admin/blog/:id/ai-history [get]
func (s *Server) GetBlogAIHistory(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		helpers.BadRequest(c, "ID không hợp lệ", nil)
		return
	}

	blogID := int32(id)
	history, err := s.z.GetBlogAIHistory(ctx, &blogID)
	if err != nil {
		helpers.InternalServerError(c, "Lỗi khi lấy lịch sử AI", err)
		return
	}

	helpers.Ok(c, gin.H{
		"message": "Lấy lịch sử AI thành công",
		"data":    history,
	})
}

// ========== REQUEST STRUCTS ==========

type CreateBlogRequest struct {
	TieuDe     string   `json:"tieu_de" binding:"required"`
	Slug       string   `json:"slug"`
	TomTat     string   `json:"tom_tat"`
	NoiDung    string   `json:"noi_dung" binding:"required"`
	AnhDaiDien string   `json:"anh_dai_dien"`
	DanhMuc    string   `json:"danh_muc"`
	TuKhoa     []string `json:"tu_khoa"`
	TrangThai  string   `json:"trang_thai" binding:"required"`
	NoiBat     bool     `json:"noi_bat"`
	NgayDang   string   `json:"ngay_dang"`
}

type UpdateBlogRequest struct {
	TieuDe     *string   `json:"tieu_de"`
	Slug       *string   `json:"slug"`
	TomTat     *string   `json:"tom_tat"`
	NoiDung    *string   `json:"noi_dung"`
	AnhDaiDien *string   `json:"anh_dai_dien"`
	DanhMuc    *string   `json:"danh_muc"`
	TuKhoa     *[]string `json:"tu_khoa"`
	TrangThai  *string   `json:"trang_thai"`
	NoiBat     *bool     `json:"noi_bat"`
	NgayDang   *string   `json:"ngay_dang"`
}

type GenerateBlogContentRequest struct {
	Topic             string `json:"topic" binding:"required"`
	BlogType          string `json:"blog_type"` // kinh_nghiem, dia_diem, huong_dan, tin_tuc, review
	AdditionalContext string `json:"additional_context"`
}

type GenerateTitleSuggestionsRequest struct {
	Topic string `json:"topic" binding:"required"`
	Count int    `json:"count"`
}

type CreateBlogWithAIRequest struct {
	Topic             string   `json:"topic" binding:"required"`
	BlogType          string   `json:"blog_type"`
	AdditionalContext string   `json:"additional_context"`
	AnhDaiDien        string   `json:"anh_dai_dien"`
	DanhMuc           string   `json:"danh_muc"`
	TuKhoa            []string `json:"tu_khoa"`
	TrangThai         string   `json:"trang_thai"`
	NoiBat            bool     `json:"noi_bat"`
	NgayDang          string   `json:"ngay_dang"`
}
