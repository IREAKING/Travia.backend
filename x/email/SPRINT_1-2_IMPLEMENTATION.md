# SPRINT 1-2 IMPLEMENTATION GUIDE

## ✅ ĐÃ HOÀN THÀNH

### 1. SQL Queries
- ✅ `db/query/booking.sql` - Booking Management (20+ queries)
- ✅ `db/query/departure.sql` - Departure Management (16+ queries)
- ✅ `db/query/review.sql` - Review Management (18+ queries)
- ✅ `db/query/tour.sql` - Tour CRUD + Search/Filter (18+ queries)

### 2. Handlers
- ✅ `api/handler/booking.go` - Full booking management
- ✅ `api/handler/departure.go` - Full departure management
- ✅ `api/handler/review.go` - Full review management

---

## 🔨 CẦN HOÀN THÀNH

### 3. Generate SQLC Code
```bash
cd /path/to/project
sqlc generate
```

**Lỗi hiện tại:** Schema migration conflict với bảng `thanh_toan`.  
**Fix:** Kiểm tra `db/migration/002_add_payments.sql` và `db/migration/schema.sql` có duplicate table definition không.

### 4. Cập nhật `api/handler/tour.go`

Thêm các handlers sau vào file tour.go:

```go
// CreateTour godoc
// @Summary      Tạo tour mới
// @Description  Tạo tour mới (Admin/Supplier)
// @Tags         tour
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Router       /tour/create [post]
func (s *Server) CreateTour(c *gin.Context) {
    // Implementation using s.z.CreateTour()
}

// UpdateTour godoc
// @Summary      Cập nhật tour
// @Description  Cập nhật thông tin tour
// @Tags         tour
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Router       /tour/{id} [put]
func (s *Server) UpdateTour(c *gin.Context) {
    // Implementation using s.z.UpdateTour()
}

// DeleteTour godoc
// @Summary      Xóa tour
// @Description  Xóa tour
// @Tags         tour
// @Security     BearerAuth
// @Router       /tour/{id} [delete]
func (s *Server) DeleteTour(c *gin.Context) {
    // Implementation using s.z.DeleteTour()
}

// SearchTours godoc
// @Summary      Tìm kiếm tour
// @Description  Tìm kiếm tour theo từ khóa
// @Tags         tour
// @Produce      json
// @Param        q query string true "Search query"
// @Param        limit query int false "Limit"
// @Param        offset query int false "Offset"
// @Router       /tour/search [get]
func (s *Server) SearchTours(c *gin.Context) {
    // Implementation using s.z.SearchTours()
}

// FilterTours godoc
// @Summary      Lọc tour
// @Description  Lọc tour theo nhiều tiêu chí
// @Tags         tour
// @Produce      json
// @Param        category query int false "Category ID"
// @Param        price_min query float false "Min price"
// @Param        price_max query float false "Max price"
// @Param        days_min query int false "Min days"
// @Param        days_max query int false "Max days"
// @Param        rating_min query float false "Min rating"
// @Param        sort_by query string false "Sort by"
// @Router       /tour/filter [get]
func (s *Server) FilterTours(c *gin.Context) {
    // Implementation using s.z.FilterTours()
}

// GetFeaturedTours godoc
// @Summary      Lấy tour nổi bật
// @Description  Lấy danh sách tour nổi bật
// @Tags         tour
// @Produce      json
// @Param        limit query int false "Limit"
// @Router       /tour/featured [get]
func (s *Server) GetFeaturedTours(c *gin.Context) {
    // Implementation using s.z.GetFeaturedTours()
}

// DuplicateTour godoc
// @Summary      Nhân bản tour
// @Description  Tạo bản sao của tour
// @Tags         tour
// @Security     BearerAuth
// @Router       /tour/{id}/duplicate [post]
func (s *Server) DuplicateTour(c *gin.Context) {
    // Implementation using s.z.DuplicateTour()
}

// GetTourImages godoc
// @Summary      Lấy ảnh tour
// @Description  Lấy tất cả ảnh của tour
// @Tags         tour
// @Router       /tour/{id}/images [get]
func (s *Server) GetTourImages(c *gin.Context) {
    // Implementation using s.z.GetTourImages()
}

// AddTourImage godoc
// @Summary      Thêm ảnh tour
// @Description  Upload ảnh cho tour
// @Tags         tour
// @Security     BearerAuth
// @Router       /tour/{id}/images [post]
func (s *Server) AddTourImage(c *gin.Context) {
    // Implementation using s.z.AddTourImage()
}

// SetPrimaryTourImage godoc
// @Summary      Đặt ảnh chính
// @Description  Đặt ảnh làm ảnh chính của tour
// @Tags         tour
// @Security     BearerAuth
// @Router       /tour/{tour_id}/images/{image_id}/primary [put]
func (s *Server) SetPrimaryTourImage(c *gin.Context) {
    // Implementation using s.z.SetPrimaryTourImage()
}

// DeleteTourImage godoc
// @Summary      Xóa ảnh tour
// @Description  Xóa một ảnh của tour
// @Tags         tour
// @Security     BearerAuth
// @Router       /tour/{tour_id}/images/{image_id} [delete]
func (s *Server) DeleteTourImage(c *gin.Context) {
    // Implementation using s.z.DeleteTourImage()
}
```

### 5. Cập nhật `api/handler/router.go`

Thêm routes sau vào `SetupRoutes()`:

```go
// ==================== BOOKING ROUTES ====================
booking := api.Group("/booking")
{
    // Public
    booking.GET("/check-availability", s.CheckAvailability)
    
    // Protected - User
    bookingAuth := booking.Group("")
    bookingAuth.Use(middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret))
    {
        bookingAuth.POST("/create", s.CreateBooking)
        bookingAuth.GET("/my-bookings", s.GetMyBookings)
        bookingAuth.GET("/:id", s.GetBookingByID)
        bookingAuth.PUT("/:id/cancel", s.CancelBooking)
        bookingAuth.GET("/:id/passengers", s.GetBookingPassengers)
        bookingAuth.POST("/:id/passengers", s.AddPassenger)
    }
    
    // Protected - Admin
    bookingAdmin := booking.Group("")
    bookingAdmin.Use(
        middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
        middleware.RequireRoles("quan_tri"),
    )
    {
        bookingAdmin.GET("/all", s.GetAllBookings)
        bookingAdmin.GET("/by-status", s.GetBookingsByStatusAdmin)
        bookingAdmin.PUT("/:id/status", s.UpdateBookingStatus)
    }
}

// ==================== DEPARTURE ROUTES ====================
departure := api.Group("/departure")
{
    // Public
    departure.GET("/:id", s.GetDepartureByID)
    departure.GET("/tour/:tour_id", s.GetDeparturesByTour)
    departure.GET("/available", s.GetAvailableDepartures)
    departure.GET("/upcoming", s.GetUpcomingDepartures)
    
    // Protected - Admin/Supplier
    departureWrite := departure.Group("")
    departureWrite.Use(
        middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
        middleware.RequireRoles("quan_tri", "nha_cung_cap"),
    )
    {
        departureWrite.POST("/create", s.CreateDeparture)
        departureWrite.PUT("/:id", s.UpdateDeparture)
        departureWrite.PUT("/:id/cancel", s.CancelDeparture)
        departureWrite.DELETE("/:id", s.DeleteDeparture)
    }
}

// ==================== REVIEW ROUTES ====================
review := api.Group("/review")
{
    // Public
    review.GET("/tour/:tour_id", s.GetReviewsByTour)
    review.GET("/recent", s.GetRecentReviews)
    review.GET("/top-rated-tours", s.GetTopRatedTours)
    
    // Protected - User
    reviewAuth := review.Group("")
    reviewAuth.Use(middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret))
    {
        reviewAuth.POST("/create", s.CreateReview)
        reviewAuth.GET("/my-reviews", s.GetMyReviews)
        reviewAuth.PUT("/:id", s.UpdateReview)
        reviewAuth.DELETE("/:id", s.DeleteReview)
    }
    
    // Protected - Admin
    reviewAdmin := review.Group("")
    reviewAdmin.Use(
        middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
        middleware.RequireRoles("quan_tri"),
    )
    {
        reviewAdmin.PUT("/:id/moderate", s.ModerateReview)
    }
}

// ==================== UPDATE TOUR ROUTES ====================
// Thêm vào tour group hiện có:
tour := api.Group("/tour")
{
    // Public GET routes (existing + new)
    tour.GET("/getAllTourCategory", ...)
    tour.GET("/getAllTour", ...)
    tour.GET("/getTourDetailByID/:id", ...)
    tour.GET("/search", s.SearchTours)              // NEW
    tour.GET("/filter", s.FilterTours)              // NEW
    tour.GET("/featured", s.GetFeaturedTours)       // NEW
    tour.GET("/:id/images", s.GetTourImages)        // NEW
    
    // Protected write routes
    tourWrite := tour.Group("")
    tourWrite.Use(
        middleware.AuthMiddleware(s.config.ServerConfig.ApiSecret),
        middleware.RequireRoles("quan_tri", "nha_cung_cap"),
    )
    {
        tourWrite.POST("/create", s.CreateTour)                    // NEW
        tourWrite.PUT("/:id", s.UpdateTour)                        // NEW
        tourWrite.DELETE("/:id", s.DeleteTour)                     // NEW
        tourWrite.POST("/:id/duplicate", s.DuplicateTour)          // NEW
        tourWrite.POST("/:id/images", s.AddTourImage)              // NEW
        tourWrite.PUT("/:tour_id/images/:image_id/primary",        // NEW
            s.SetPrimaryTourImage)
        tourWrite.DELETE("/:tour_id/images/:image_id",             // NEW
            s.DeleteTourImage)
    }
}
```

### 6. Email Notification Integration

Tạo `api/helpers/email_notification.go`:

```go
package helpers

import (
    "fmt"
    "net/smtp"
)

func SendBookingConfirmation(toEmail, bookingID, tourName string) error {
    from := "noreply@travia.com"
    password := "your-smtp-password"
    
    to := []string{toEmail}
    
    smtpHost := "smtp.gmail.com"
    smtpPort := "587"
    
    message := []byte(fmt.Sprintf(
        "Subject: Booking Confirmation - %s\n\n"+
        "Your booking ID: %s has been confirmed.\n"+
        "Tour: %s\n\n"+
        "Thank you for booking with Travia!",
        tourName, bookingID, tourName,
    ))
    
    auth := smtp.PlainAuth("", from, password, smtpHost)
    
    err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message)
    if err != nil {
        return err
    }
    return nil
}
```

Gọi trong `CreateBooking`:
```go
// After successful booking creation
go SendBookingConfirmation(userEmail, booking.ID, tourName)
```

### 7. Update Swagger Documentation

```bash
swag init
```

---

## 📝 API ENDPOINTS SUMMARY

### Booking (14 endpoints)
- `POST /booking/create` - Tạo booking
- `GET /booking/my-bookings` - Danh sách booking của tôi
- `GET /booking/:id` - Chi tiết booking
- `PUT /booking/:id/cancel` - Hủy booking
- `GET /booking/:id/passengers` - Danh sách hành khách
- `POST /booking/:id/passengers` - Thêm hành khách
- `GET /booking/check-availability` - Kiểm tra chỗ trống
- `GET /booking/all` - Tất cả booking (Admin)
- `GET /booking/by-status` - Booking theo status (Admin)
- `PUT /booking/:id/status` - Cập nhật status (Admin)

### Departure (9 endpoints)
- `POST /departure/create` - Tạo lịch khởi hành
- `GET /departure/:id` - Chi tiết khởi hành
- `GET /departure/tour/:tour_id` - Khởi hành theo tour
- `GET /departure/available` - Khởi hành còn chỗ
- `GET /departure/upcoming` - Khởi hành sắp tới
- `PUT /departure/:id` - Cập nhật khởi hành
- `PUT /departure/:id/cancel` - Hủy khởi hành
- `DELETE /departure/:id` - Xóa khởi hành

### Review (10 endpoints)
- `POST /review/create` - Tạo đánh giá
- `GET /review/tour/:tour_id` - Đánh giá theo tour
- `GET /review/my-reviews` - Đánh giá của tôi
- `PUT /review/:id` - Cập nhật đánh giá
- `DELETE /review/:id` - Xóa đánh giá
- `PUT /review/:id/moderate` - Kiểm duyệt (Admin)
- `GET /review/recent` - Đánh giá mới nhất
- `GET /review/top-rated-tours` - Tour đánh giá cao

### Tour CRUD (15 endpoints)
- `POST /tour/create` - Tạo tour
- `PUT /tour/:id` - Cập nhật tour
- `DELETE /tour/:id` - Xóa tour
- `GET /tour/search` - Tìm kiếm tour
- `GET /tour/filter` - Lọc tour
- `GET /tour/featured` - Tour nổi bật
- `POST /tour/:id/duplicate` - Nhân bản tour
- `GET /tour/:id/images` - Danh sách ảnh
- `POST /tour/:id/images` - Thêm ảnh
- `PUT /tour/:tour_id/images/:image_id/primary` - Đặt ảnh chính
- `DELETE /tour/:tour_id/images/:image_id` - Xóa ảnh

**TỔNG: ~48 endpoints mới**

---

## 🧪 TESTING

### Test Booking Flow
```bash
# 1. Check availability
curl http://localhost:8080/api/booking/check-availability?departure_id=1

# 2. Create booking
curl -X POST http://localhost:8080/api/booking/create \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "khoi_hanh_id": 1,
    "so_nguoi_lon": 2,
    "so_tre_em": 1,
    "tong_tien": 15000000,
    "don_vi_tien_te": "VND"
  }'

# 3. Add passengers
curl -X POST http://localhost:8080/api/booking/1/passengers \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "ho_ten": "Nguyen Van A",
    "ngay_sinh": "1990-01-01",
    "loai_khach": "nguoi_lon",
    "gioi_tinh": "nam"
  }'

# 4. Get my bookings
curl http://localhost:8080/api/booking/my-bookings \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Test Review Flow
```bash
# 1. Create review (after tour completed)
curl -X POST http://localhost:8080/api/review/create \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tour_id": 1,
    "dat_cho_id": 1,
    "rating": 5,
    "tieu_de": "Tour tuyệt vời",
    "noi_dung": "Rất hài lòng với chuyến đi..."
  }'

# 2. Get reviews by tour
curl http://localhost:8080/api/review/tour/1
```

---

## ✅ CHECKLIST

- [ ] Fix sqlc generation error
- [ ] Add tour CRUD handlers to tour.go
- [ ] Update router.go with all new routes
- [ ] Implement email notification
- [ ] Test all booking endpoints
- [ ] Test all departure endpoints
- [ ] Test all review endpoints
- [ ] Test tour CRUD endpoints
- [ ] Update Swagger documentation
- [ ] Update README.md with new endpoints

---

## 🚀 NEXT STEPS (Sprint 3-4)

1. Itinerary Management (lịch trình tour chi tiết)
2. User Profile enhancements (change password, wishlist)
3. Advanced search với full-text search
4. Notification system (in-app + push)
5. Discount/Promotion management

