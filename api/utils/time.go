package utils

import (
	"strings"
	"time"
)

func OnlyTine(s string) string {
	// Bỏ khoảng trắng thừa (nếu có)
	s = strings.TrimSpace(s)

	// Nếu người dùng nhập dạng "HH:MM" → thêm ":00"
	if len(s) == 5 && strings.Count(s, ":") == 1 {
		s = s + ":00"
	}

	// Nếu người dùng nhập dạng "H:MM" → thêm số 0 đầu + giây
	if len(s) == 4 && strings.Count(s, ":") == 1 {
		s = "0" + s + ":00"
	}

	// Nếu người dùng nhập dạng "H:MM:SS" → thêm 0 đầu giờ
	if len(s) == 7 && s[1] == ':' {
		s = "0" + s
	}

	// Dùng time.Parse để parse linh hoạt
	// t, err := time.Parse(time.TimeOnly, s)
	// if err != nil {
	// 	fmt.Println("lỗi parse:", err)
	// 	return nil
	// }

	return s
}
func StringToDate(s string) time.Time {
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return time.Time{}
	}
	return t
}
