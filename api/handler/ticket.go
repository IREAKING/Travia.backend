package handler

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/signintech/gopdf"
)

/* ================== HANDLER ================== */

func (s *Server) PrintTicket(c *gin.Context) {
	idStr := c.Param("dat_cho_id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "ID không hợp lệ"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	booking, err := s.z.GetBookingById(ctx, int32(id))
	if err != nil {
		c.JSON(404, gin.H{"error": "Không tìm thấy booking"})
		return
	}

	tickets, err := s.z.GetTicketsByBookingID(ctx, int32(id))
	if err != nil || len(tickets) == 0 {
		c.JSON(404, gin.H{"error": "Không có vé"})
		return
	}

	ticket := tickets[0]
	bookingCode := fmt.Sprintf("BK-%d", booking.ID)

	/* ================== PDF INIT ================== */

	// Tìm đường dẫn font
	_, filename, _, _ := runtime.Caller(0)
	handlerDir := filepath.Dir(filename)
	backendDir := filepath.Dir(filepath.Dir(handlerDir))
	fontDir := filepath.Join(backendDir, "font")
	fontRegularPath := filepath.Join(fontDir, "BeVietnamPro-Regular.ttf")
	fontBoldPath := filepath.Join(fontDir, "BeVietnamPro-Bold.ttf")

	// Convert sang absolute path
	absFontRegular, _ := filepath.Abs(fontRegularPath)
	absFontBold, _ := filepath.Abs(fontBoldPath)

	// Tạo PDF A5 (148x210mm)
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{
		PageSize: gopdf.Rect{
			W: 148,
			H: 210,
		},
		Unit: gopdf.UnitMM,
	})
	pdf.AddPage()

	// Thêm font
	fontName := "BeVietnamPro"
	if err := pdf.AddTTFFont(fontName, absFontRegular); err != nil {
		// Fallback nếu không tìm thấy font
		c.JSON(500, gin.H{"error": "Không thể load font Regular", "path": absFontRegular, "details": err.Error()})
		return
	}
	if err := pdf.AddTTFFontWithOption(fontName+"Bold", absFontBold, gopdf.TtfOption{
		UseKerning: true,
	}); err != nil {
		c.JSON(500, gin.H{"error": "Không thể load font Bold", "path": absFontBold, "details": err.Error()})
		return
	}

	/* ================== BACKGROUND ================== */

	// Nền tối
	pdf.SetFillColor(2, 6, 23)
	pdf.RectFromUpperLeft(0, 0, 148, 210)

	/* ================== HEADER ================== */

	// Header background
	pdf.SetFillColor(37, 99, 235)
	pdf.RectFromUpperLeft(10, 15, 128, 25)

	// Title
	pdf.SetFont(fontName+"Bold", "", 18)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetX(10)
	pdf.SetY(20)
	pdf.Cell(nil, "VÉ ĐIỆN TỬ")

	// Booking code
	pdf.SetFont(fontName, "", 10)
	pdf.SetX(10)
	pdf.SetY(30)
	pdf.Cell(nil, "MÃ ĐẶT CHỖ • "+bookingCode)

	/* ================== TOUR NAME ================== */

	pdf.SetFont(fontName+"Bold", "", 16)
	pdf.SetTextColor(56, 189, 248)
	pdf.SetX(10)
	pdf.SetY(50)
	pdf.Cell(nil, strings.ToUpper(ticket.TenTour))

	/* ================== INFO ================== */

	y := 75.0
	drawField := func(label, value string, x, y float64, isGreen bool) {
		// Label
		pdf.SetFont(fontName, "", 9)
		pdf.SetTextColor(148, 163, 184)
		pdf.SetX(x)
		pdf.SetY(y)
		pdf.Cell(nil, label)

		// Value
		if isGreen {
			pdf.SetTextColor(34, 197, 94)
		} else {
			pdf.SetTextColor(248, 250, 252)
		}
		pdf.SetFont(fontName+"Bold", "", 11)
		pdf.SetX(x)
		pdf.SetY(y + 6)
		pdf.Cell(nil, value)
	}

	drawField("KHÁCH HÀNG", ticket.HoTen, 15, y, false)
	drawField("NGÀY KHỞI HÀNH", formatDate(ticket.NgayKhoiHanh), 80, y, false)

	y += 20
	drawField("SỐ KHÁCH", fmt.Sprintf("%d người", booking.SoNguoiLon), 15, y, false)

	trangThai := "CHỜ XÁC NHẬN"
	if booking.TrangThai.Valid {
		trangThai = strings.ReplaceAll(strings.ToUpper(string(booking.TrangThai.TrangThaiDatCho)), "_", " ")
	}
	drawField("TRẠNG THÁI", "● "+trangThai, 80, y, true)

	/* ================== FOOTER ================== */

	// Line
	pdf.SetStrokeColor(51, 65, 85)
	pdf.Line(10, 165, 138, 165)

	// Total label
	pdf.SetFont(fontName, "", 9)
	pdf.SetTextColor(148, 163, 184)
	pdf.SetX(15)
	pdf.SetY(170)
	pdf.Cell(nil, "TỔNG THANH TOÁN")

	// Total amount
	var total float64
	if ticket.TongTien.Valid {
		v, _ := ticket.TongTien.Float64Value()
		total = v.Float64
	}

	pdf.SetFont(fontName+"Bold", "", 22)
	pdf.SetTextColor(244, 114, 182)
	pdf.SetX(15)
	pdf.SetY(178)
	pdf.Cell(nil, formatCurrency(total, *ticket.DonViTienTe))

	/* ================== EXPORT ================== */

	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		c.JSON(500, gin.H{"error": "Không thể tạo PDF", "details": err.Error()})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header(
		"Content-Disposition",
		fmt.Sprintf("attachment; filename=Ticket-%s.pdf", bookingCode),
	)
	c.Data(http.StatusOK, "application/pdf", buf.Bytes())
}

/* ================== HELPERS ================== */

func formatDate(d pgtype.Date) string {
	if !d.Valid {
		return "N/A"
	}
	return d.Time.Format("02/01/2006")
}

func formatCurrency(amount float64, currency string) string {
	s := fmt.Sprintf("%.0f", amount)
	var parts []string
	for i := len(s); i > 0; i -= 3 {
		start := i - 3
		if start < 0 {
			start = 0
		}
		parts = append([]string{s[start:i]}, parts...)
	}
	return strings.Join(parts, ".") + " ₫"
}
