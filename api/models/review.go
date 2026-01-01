package models

import (
	"encoding/json"
	"time"
)

// CreateReviewRequest chứa dữ liệu để tạo đánh giá tour
type CreateReviewRequest struct {
	DatChoID       int32    `json:"dat_cho_id" binding:"required"`                // ID của booking đã hoàn thành
	DiemDanhGia    int32    `json:"diem_danh_gia" binding:"required,min=1,max=5"` // Điểm đánh giá từ 1-5
	TieuDe         *string  `json:"tieu_de"`                                      // Tiêu đề đánh giá (optional)
	NoiDung        *string  `json:"noi_dung"`                                     // Nội dung đánh giá (optional)
	HinhAnhDinhKem []string `json:"hinh_anh_dinh_kem"`                            // Mảng URL ảnh đính kèm (optional)
}

type ReviewDetail struct {
	ID             int32           `json:"id"`
	TieuDe         string          `json:"tieu_de"`
	DiemDanhGia    int32           `json:"diem_danh_gia"`
	NoiDung        string          `json:"noi_dung"`
	HinhAnhDinhKem json.RawMessage `json:"hinh_anh_dinh_kem"` // Giữ nguyên JSON của mảng ảnh
	NgayTao        time.Time       `json:"ngay_tao"`
	HoTen          string          `json:"ho_ten"`
}
