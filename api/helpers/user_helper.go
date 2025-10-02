package helpers

import (
	"regexp"
	"strings"
)

// nhận từ client
type UserInput struct {
	FirsName string `json:"first_name"`
	LastName string `json:"last_name"`
	Email    string `json:"emaiil"`
	Password string `json:"password"`
	Phone    string `json:"phone"`
}

// trả về
type User struct {
	Id        string `json:"id"`
	FullName  string `json:"full_name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	Phone     string `json:"phone"`
	CreatedAt string `json:"created_at"`
}

func AddFullName(firstname, lastname string) string {
	return strings.TrimSpace(firstname + " " + lastname)
}

// kiểm ta email có hợp lệ không
func ValidateEmail(email string) bool {
	regex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return regex.MatchString(email)
}

// ValidatePassword kiểm tra password đủ mạnh (tối thiểu 8 ký tự, có số & chữ cái)
func ValidatePassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	return hasLetter && hasNumber
}

// FormatUser format dữ liệu user để trả về API (ẩn password, token,...)
func FormatUser(id, firstName, lastName, email, phone, role, createdAt string) User {
	return User{
		Id:        id,
		FullName:  AddFullName(firstName, lastName),
		Email:     email,
		Phone:     phone,
		Role:      role,
		CreatedAt: createdAt,
	}
}
