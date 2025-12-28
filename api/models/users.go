package models

type CreateUser struct {
	FullName string  `json:"full_name" binding:"required"`
	Email    string  `json:"email" binding:"required,email"`
	Password string  `json:"password" binding:"required,min=8"`
	Phone    *string `json:"phone"`
}
type PendingUser struct {
	FullName string  `json:"full_name"`
	Email    string  `json:"email"`
	Password string  `json:"password"`
	Phone    *string `json:"phone"`
	Otp      string  `json:"otp"`
}
type OTP struct {
	Email string `json:"email"`
	Otp   string `json:"otp"`
}
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type GetUser struct {
	ID                       string  `json:"id"`
	FullName                 string  `json:"full_name"`
	Email                    string  `json:"email"`
	Phone                    *string `json:"phone"`
	TongDatCho               int64   `json:"tong_dat_cho"`
	TongDatChoDaThanhToan    int64   `json:"tong_dat_cho_da_thanh_toan"`
	TongDatChoDangChoXacNhan int64   `json:"tong_dat_cho_dang_cho_xac_nhan"`
	NgayCapNhat              string  `json:"ngay_cap_nhat"`
	NgayTao                  string  `json:"ngay_tao"`
}
type UpdateUser struct {
	FullName    string  `json:"full_name"`
	Email       string  `json:"email"`
	Phone       *string `json:"phone"`
	NgayCapNhat string  `json:"ngay_cap_nhat"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type AddPassengersParams struct {
	DatChoID         int32   `json:"dat_cho_id"`
	HoTen            string  `json:"ho_ten"`
	NgaySinh         string  `json:"ngay_sinh"`
	LoaiKhach        *string `json:"loai_khach"`
	GioiTinh         *string `json:"gioi_tinh"`
	SoGiayToTuyThanh *string `json:"so_giay_to_tuy_thanh"`
	QuocTich         *string `json:"quoc_tich"`
	GhiChu           *string `json:"ghi_chu"`
}
