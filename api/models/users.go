package models

type CreateUser struct {
	FullName string  `json:"full_name"`
	Email    string  `json:"email"`
	Password string  `json:"password"`
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
type GetUpdateUser struct {
	ID          string  `json:"id"`
	FullName    string  `json:"full_name"`
	Email       string  `json:"email"`
	Phone       *string `json:"phone"`
	NgayCapNhat string  `json:"ngay_cap_nhat"`
}
