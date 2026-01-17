package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"travia.backend/api/helpers"
	"travia.backend/api/utils"
	db "travia.backend/db/sqlc"
)

// Chatbot godoc
// @Summary Chat với AI chatbot
// @Description Chat với AI chatbot để được tư vấn về tours. Hỗ trợ lưu lịch sử chat.
// @Tags AI
// @Accept json
// @Produce json
// @Param message body object true "Chat message" SchemaExample({"message":"string","session_id":"string","include_tours":true})
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /ai/chatbot [post]
func (s *Server) Chatbot(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	var req struct {
		Message      string `json:"message"`
		SessionID    string `json:"session_id"`
		IncludeTours bool   `json:"include_tours"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"message": "Dữ liệu yêu cầu không hợp lệ: " + err.Error(),
		})
		return
	}

	if req.Message == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing message",
			"message": "Vui lòng nhập câu hỏi",
		})
		return
	}

	// Generate session ID nếu chưa có
	if req.SessionID == "" {
		req.SessionID = fmt.Sprintf("session_%d", time.Now().UnixNano())
	}

	// Lấy user ID nếu đã đăng nhập
	var userID *string
	claims, exists := c.Get("claims")
	if exists {
		if jwtClaims, ok := claims.(*utils.JwtClams); ok {
			userIDStr := jwtClaims.Id.String()
			userID = &userIDStr
		}
	}

	// Lấy lịch sử chat gần đây (10 câu gần nhất)
	var chatHistory []string
	var userUUID pgtype.UUID

	if userID != nil {
		// Lấy từ user ID
		if err := userUUID.Scan(*userID); err != nil {
			userUUID = pgtype.UUID{} // Reset nếu scan fail
		}
	}

	// Lấy chat history từ database
	recentChats, err := s.z.GetRecentChatHistory(ctx, db.GetRecentChatHistoryParams{
		Column1: userUUID,
		MaPhien: req.SessionID,
		Limit:   10,
	})

	if err == nil {
		// Xây dựng chat history từ database
		for _, chat := range recentChats {
			chatHistory = append(chatHistory, fmt.Sprintf("Q: %s\nA: %s", chat.CauHoi, chat.CauTraLoi))
		}
	}

	// Lấy danh sách tours nếu cần (ưu tiên tìm theo câu hỏi để có ngữ cảnh phù hợp)
	toursList := ""
	var candidateTours []db.GetTourContextForAIRow
	dateRange := parseDateRangeFromMessage(req.Message, time.Now())
	if req.IncludeTours {
		searchTours, err := s.z.GetTourContextForAI(ctx, db.GetTourContextForAIParams{
			Limit:           5,
			Offset:          0,
			Keyword:         &req.Message,
			DiemDenTen:      nil,
			SoNgayMin:       nil,
			SoNgayMax:       nil,
			GiaMin:          pgtype.Numeric{},
			GiaMax:          pgtype.Numeric{},
			NgayKhoiHanhTu:  dateRange.Start,
			NgayKhoiHanhDen: dateRange.End,
			ChiCoGiamGia:    nil,
		})
		if err == nil && len(searchTours) > 0 {
			candidateTours = searchTours
		}
	}

	if req.IncludeTours && len(candidateTours) == 0 {
		tours, err := s.z.GetTourContextForAI(ctx, db.GetTourContextForAIParams{
			Limit:           5,
			Offset:          0,
			Keyword:         nil,
			DiemDenTen:      nil,
			SoNgayMin:       nil,
			SoNgayMax:       nil,
			GiaMin:          pgtype.Numeric{},
			GiaMax:          pgtype.Numeric{},
			NgayKhoiHanhTu:  dateRange.Start,
			NgayKhoiHanhDen: dateRange.End,
			ChiCoGiamGia:    nil,
		})
		if err == nil {
			candidateTours = tours
		}
	}

	if len(candidateTours) > 0 {
		for i, tour := range candidateTours {
			if i >= 5 {
				break
			}
			toursList += buildTourSummaryContext(tour)
		}
	}

	// Lấy chi tiết tour nếu người dùng hỏi sâu về lịch trình/chi tiết hoặc nhắc ID tour
	tourDetailContext := ""
	if req.IncludeTours {
		tourID := extractTourID(req.Message)
		if tourID > 0 {
			if detail, err := s.z.GetTourDetailByID(ctx, int32(tourID)); err == nil {
				tourDetailContext = buildTourDetailContext(detail)
			}
		} else if len(candidateTours) > 0 && shouldIncludeTourDetail(req.Message) {
			tourDetailContext = buildTourDetailContextFromAI(candidateTours[0])
		}
	}

	// Generate response từ AI
	response, err := helpers.GenerateChatbotResponse(
		s.config.OpenAIConfig.APIKey,
		req.Message,
		chatHistory,
		toursList,
		tourDetailContext,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate response",
			"message": "Không thể tạo phản hồi: " + err.Error(),
		})
		return
	}

	// Lưu vào database (userUUID đã được set ở trên)
	if userID == nil {
		userUUID = pgtype.UUID{} // NULL nếu không có user
	}

	_, err = s.z.CreateChatHistory(ctx, db.CreateChatHistoryParams{
		NguoiDungID: userUUID,
		MaPhien:     req.SessionID,
		CauHoi:      req.Message,
		CauTraLoi:   response,
	})
	if err != nil {
		// Log error nhưng không fail request
		fmt.Printf("Failed to save chat history: %v\n", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Phản hồi thành công",
		"session_id": req.SessionID,
		"data": gin.H{
			"response": response,
		},
	})
}

type destinationInfo struct {
	Ten     string `json:"ten"`
	Tinh    string `json:"tinh"`
	QuocGia string `json:"quoc_gia"`
}

type itineraryInfo struct {
	NgayThu int32   `json:"ngay_thu"`
	TieuDe  string  `json:"tieu_de"`
	MoTa    *string `json:"mo_ta"`
	DiaDiem *string `json:"dia_diem"`
}

type departureInfo struct {
	NgayKhoiHanh string `json:"ngay_khoi_hanh"`
	NgayKetThuc  string `json:"ngay_ket_thuc"`
	SucChua      int32  `json:"suc_chua"`
	SoChoDaDat   int32  `json:"so_cho_da_dat"`
	TrangThai    string `json:"trang_thai"`
}

func buildTourSummaryContext(tour db.GetTourContextForAIRow) string {
	var b strings.Builder
	currency := "VND"
	if tour.DonViTienTe != nil {
		currency = *tour.DonViTienTe
	}

	baseAdult := formatNumeric(tour.GiaNguoiLon)
	baseChild := formatNumeric(tour.GiaTreEm)
	discountPercent, hasDiscount := numericToFloat64(tour.GiamGiaPhanTram)
	finalAdult := baseAdult
	finalChild := baseChild
	if hasDiscount && discountPercent > 0 {
		if adultVal, ok := numericToFloat64(tour.GiaNguoiLon); ok {
			finalAdult = fmt.Sprintf("%.0f", adultVal*(1-discountPercent/100))
		}
		if childVal, ok := numericToFloat64(tour.GiaTreEm); ok {
			finalChild = fmt.Sprintf("%.0f", childVal*(1-discountPercent/100))
		}
	}

	destinations := formatDestinationsFromBytes(tour.Destinations)
	nextDeparture := formatDateValue(tour.NextDepartureDate)

	b.WriteString(fmt.Sprintf("- %s\n", tour.TieuDe))
	b.WriteString(fmt.Sprintf("  • Thời lượng: %d ngày %d đêm\n", tour.SoNgay, tour.SoDem))
	if destinations != "" {
		b.WriteString(fmt.Sprintf("  • Điểm đến: %s\n", destinations))
	}
	if finalAdult != "" || finalChild != "" {
		b.WriteString(fmt.Sprintf("  • Giá: NL %s %s, TE %s %s\n", finalAdult, currency, finalChild, currency))
	}
	if hasDiscount && discountPercent > 0 {
		b.WriteString(fmt.Sprintf("  • Giảm giá: %.0f%%\n", discountPercent))
	}
	if nextDeparture != "" {
		b.WriteString(fmt.Sprintf("  • Khởi hành gần nhất: %s\n", nextDeparture))
	}
	if tour.AvgRating > 0 {
		b.WriteString(fmt.Sprintf("  • Đánh giá: %.1f/5 (%d)\n", tour.AvgRating, tour.TotalReviews))
	}
	return b.String()
}

func buildTourDetailContextFromAI(tour db.GetTourContextForAIRow) string {
	var b strings.Builder
	currency := "VND"
	if tour.DonViTienTe != nil {
		currency = *tour.DonViTienTe
	}

	b.WriteString("Chi tiết tour được tham chiếu:\n")
	b.WriteString(fmt.Sprintf("- Tên tour: %s\n", tour.TieuDe))
	if tour.DanhMucTen != nil {
		b.WriteString(fmt.Sprintf("- Danh mục: %s\n", *tour.DanhMucTen))
	}
	if tour.NhaCungCapTen != nil {
		b.WriteString(fmt.Sprintf("- Nhà cung cấp: %s\n", *tour.NhaCungCapTen))
	}
	b.WriteString(fmt.Sprintf("- Thời lượng: %d ngày %d đêm\n", tour.SoNgay, tour.SoDem))
	if tour.MoTa != nil && strings.TrimSpace(*tour.MoTa) != "" {
		b.WriteString(fmt.Sprintf("- Mô tả: %s\n", strings.TrimSpace(*tour.MoTa)))
	}

	baseAdult := formatNumeric(tour.GiaNguoiLon)
	baseChild := formatNumeric(tour.GiaTreEm)
	if baseAdult != "" || baseChild != "" {
		b.WriteString(fmt.Sprintf("- Giá gốc: NL %s %s, TE %s %s\n", baseAdult, currency, baseChild, currency))
	}

	if discountPercent, ok := numericToFloat64(tour.GiamGiaPhanTram); ok && discountPercent > 0 {
		b.WriteString(fmt.Sprintf("- Giảm giá hiện tại: %.0f%%\n", discountPercent))
	}

	destinations := parseDestinations(tour.Destinations)
	if len(destinations) > 0 {
		b.WriteString("- Điểm đến: ")
		for i, d := range destinations {
			if i >= 5 {
				b.WriteString("...")
				break
			}
			if i > 0 {
				b.WriteString(", ")
			}
			if d.Ten != "" {
				b.WriteString(d.Ten)
			} else {
				b.WriteString("N/A")
			}
		}
		b.WriteString("\n")
	}

	itinerary := parseItinerary(tour.Itinerary)
	if len(itinerary) > 0 {
		b.WriteString("- Lịch trình (tóm tắt):\n")
		for i, day := range itinerary {
			if i >= 5 {
				b.WriteString("  • ...\n")
				break
			}
			title := day.TieuDe
			if title == "" {
				title = "Không có tiêu đề"
			}
			location := ""
			if day.DiaDiem != nil && *day.DiaDiem != "" {
				location = fmt.Sprintf(" - %s", *day.DiaDiem)
			}
			b.WriteString(fmt.Sprintf("  • Ngày %d: %s%s\n", day.NgayThu, title, location))
		}
	}

	departures := parseDepartures(tour.Departures)
	if len(departures) > 0 {
		b.WriteString("- Lịch khởi hành (gần nhất):\n")
		for i, d := range departures {
			if i >= 3 {
				b.WriteString("  • ...\n")
				break
			}
			start := normalizeDateString(d.NgayKhoiHanh)
			end := normalizeDateString(d.NgayKetThuc)
			remaining := d.SucChua - d.SoChoDaDat
			b.WriteString(fmt.Sprintf("  • %s - %s, còn %d/%d chỗ, trạng thái: %s\n",
				start, end, remaining, d.SucChua, d.TrangThai))
		}
	}

	return b.String()
}

func buildTourDetailContext(tour db.GetTourDetailByIDRow) string {
	var b strings.Builder
	currency := "VND"
	if tour.DonViTienTe != nil {
		currency = *tour.DonViTienTe
	}

	b.WriteString("Chi tiết tour được tham chiếu:\n")
	b.WriteString(fmt.Sprintf("- Tên tour: %s\n", tour.TieuDe))
	if tour.TenDanhMuc != nil {
		b.WriteString(fmt.Sprintf("- Danh mục: %s\n", *tour.TenDanhMuc))
	}
	if tour.TenNhaCungCap != nil {
		b.WriteString(fmt.Sprintf("- Nhà cung cấp: %s\n", *tour.TenNhaCungCap))
	}
	b.WriteString(fmt.Sprintf("- Thời lượng: %d ngày %d đêm\n", tour.SoNgay, tour.SoDem))
	if tour.MoTa != nil && strings.TrimSpace(*tour.MoTa) != "" {
		b.WriteString(fmt.Sprintf("- Mô tả: %s\n", strings.TrimSpace(*tour.MoTa)))
	}

	baseAdult := formatNumeric(tour.GiaNguoiLon)
	baseChild := formatNumeric(tour.GiaTreEm)
	if baseAdult != "" || baseChild != "" {
		b.WriteString(fmt.Sprintf("- Giá gốc: NL %s %s, TE %s %s\n", baseAdult, currency, baseChild, currency))
	}

	discount := formatNumeric(tour.GiamGiaPhanTram)
	if discount != "" {
		b.WriteString(fmt.Sprintf("- Giảm giá hiện tại: %s%%\n", discount))
	}

	if tour.SoNhoNhat != nil || tour.SoLonNhat != nil {
		minSize := "N/A"
		maxSize := "N/A"
		if tour.SoNhoNhat != nil {
			minSize = fmt.Sprintf("%d", *tour.SoNhoNhat)
		}
		if tour.SoLonNhat != nil {
			maxSize = fmt.Sprintf("%d", *tour.SoLonNhat)
		}
		b.WriteString(fmt.Sprintf("- Quy mô nhóm: %s - %s khách\n", minSize, maxSize))
	}

	destinations := parseDestinations(tour.Destinations)
	if len(destinations) > 0 {
		b.WriteString("- Điểm đến: ")
		for i, d := range destinations {
			if i >= 5 {
				b.WriteString("...")
				break
			}
			if i > 0 {
				b.WriteString(", ")
			}
			if d.Ten != "" {
				b.WriteString(d.Ten)
			} else {
				b.WriteString("N/A")
			}
		}
		b.WriteString("\n")
	}

	itinerary := parseItinerary(tour.Itinerary)
	if len(itinerary) > 0 {
		b.WriteString("- Lịch trình (tóm tắt):\n")
		for i, day := range itinerary {
			if i >= 5 {
				b.WriteString("  • ...\n")
				break
			}
			title := day.TieuDe
			if title == "" {
				title = "Không có tiêu đề"
			}
			location := ""
			if day.DiaDiem != nil && *day.DiaDiem != "" {
				location = fmt.Sprintf(" - %s", *day.DiaDiem)
			}
			b.WriteString(fmt.Sprintf("  • Ngày %d: %s%s\n", day.NgayThu, title, location))
		}
	}

	departures := parseDepartures(tour.Departures)
	if len(departures) > 0 {
		b.WriteString("- Lịch khởi hành (gần nhất):\n")
		for i, d := range departures {
			if i >= 3 {
				b.WriteString("  • ...\n")
				break
			}
			start := normalizeDateString(d.NgayKhoiHanh)
			end := normalizeDateString(d.NgayKetThuc)
			remaining := d.SucChua - d.SoChoDaDat
			b.WriteString(fmt.Sprintf("  • %s - %s, còn %d/%d chỗ, trạng thái: %s\n",
				start, end, remaining, d.SucChua, d.TrangThai))
		}
	}

	return b.String()
}

func extractTourID(message string) int32 {
	re := regexp.MustCompile(`(?i)(tour|ma\s*tour|id)\s*[:#-]?\s*(\d+)`)
	matches := re.FindStringSubmatch(message)
	if len(matches) < 3 {
		return 0
	}
	var id int32
	fmt.Sscanf(matches[2], "%d", &id)
	return id
}

func shouldIncludeTourDetail(message string) bool {
	keywords := []string{
		"chi tiết", "lich trinh", "lịch trình", "khoi hanh", "khởi hành",
		"diem den", "điểm đến", "gia", "giá", "ưu đãi", "giam gia", "giảm giá",
	}
	msg := strings.ToLower(message)
	for _, kw := range keywords {
		if strings.Contains(msg, kw) {
			return true
		}
	}
	return false
}

func formatNumeric(value pgtype.Numeric) string {
	if !value.Valid {
		return ""
	}
	floatVal, err := value.Float64Value()
	if err != nil || !floatVal.Valid {
		return ""
	}
	return fmt.Sprintf("%.0f", floatVal.Float64)
}

func numericToFloat64(value pgtype.Numeric) (float64, bool) {
	if !value.Valid {
		return 0, false
	}
	floatVal, err := value.Float64Value()
	if err != nil || !floatVal.Valid {
		return 0, false
	}
	return floatVal.Float64, true
}

func formatInterfaceNumber(value interface{}) string {
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case float64:
		return fmt.Sprintf("%.0f", v)
	case float32:
		return fmt.Sprintf("%.0f", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case int32:
		return fmt.Sprintf("%d", v)
	case int:
		return fmt.Sprintf("%d", v)
	case string:
		return v
	case []byte:
		return string(v)
	case pgtype.Numeric:
		return formatNumeric(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func formatDateValue(value interface{}) string {
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case time.Time:
		return v.Format("2006-01-02")
	case pgtype.Date:
		if v.Valid {
			return v.Time.Format("2006-01-02")
		}
	case string:
		return normalizeDateString(v)
	case []byte:
		return normalizeDateString(string(v))
	default:
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func normalizeDateString(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if idx := strings.Index(value, "T"); idx > 0 {
		return value[:idx]
	}
	return value
}

type dateRange struct {
	Start pgtype.Date
	End   pgtype.Date
}

func parseDateRangeFromMessage(message string, now time.Time) dateRange {
	msg := strings.ToLower(message)
	rng := dateRange{}
	location := now.Location()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)

	switch {
	case strings.Contains(msg, "tuần tới") || strings.Contains(msg, "tuan toi"):
		start := today.AddDate(0, 0, 7)
		end := start.AddDate(0, 0, 6)
		_ = rng.Start.Scan(start)
		_ = rng.End.Scan(end)
	case strings.Contains(msg, "tuần sau") || strings.Contains(msg, "tuan sau"):
		start := today.AddDate(0, 0, 7)
		end := start.AddDate(0, 0, 6)
		_ = rng.Start.Scan(start)
		_ = rng.End.Scan(end)
	case strings.Contains(msg, "7 ngày tới") || strings.Contains(msg, "7 ngay toi"):
		start := today
		end := today.AddDate(0, 0, 7)
		_ = rng.Start.Scan(start)
		_ = rng.End.Scan(end)
	}

	return rng
}

func formatDestinationsFromBytes(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}
	destinations := parseDestinations(raw)
	if len(destinations) == 0 {
		return ""
	}
	var items []string
	for _, dest := range destinations {
		if dest.Ten != "" {
			items = append(items, dest.Ten)
		}
	}
	return strings.Join(items, ", ")
}

func parseDestinations(raw []byte) []destinationInfo {
	if len(raw) == 0 {
		return nil
	}
	var items []destinationInfo
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil
	}
	return items
}

func parseItinerary(raw []byte) []itineraryInfo {
	if len(raw) == 0 {
		return nil
	}
	var items []itineraryInfo
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil
	}
	return items
}

func parseDepartures(raw []byte) []departureInfo {
	if len(raw) == 0 {
		return nil
	}
	var items []departureInfo
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil
	}
	return items
}
