package utils

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// GenerateSlug tạo slug từ tiêu đề tiếng Việt
func GenerateSlug(title string) string {
	// Normalize Unicode (chuyển đổi các ký tự có dấu về dạng NFD)
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	normalized, _, _ := transform.String(t, title)

	// Chuyển thành chữ thường
	slug := strings.ToLower(normalized)

	// Thay thế khoảng trắng và ký tự đặc biệt bằng dấu gạch ngang
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")

	// Loại bỏ các ký tự không phải chữ, số, hoặc dấu gạch ngang
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	slug = result.String()

	// Loại bỏ nhiều dấu gạch ngang liên tiếp
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	// Loại bỏ dấu gạch ngang ở đầu và cuối
	slug = strings.Trim(slug, "-")

	return slug
}
