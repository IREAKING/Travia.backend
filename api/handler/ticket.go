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
	db "travia.backend/db/sqlc"
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

	bookingCode := fmt.Sprintf("BK-%d", booking.ID)

	/* ================== PDF INIT ================== */

	// Tìm đường dẫn font
	_, filename, _, _ := runtime.Caller(0)
	handlerDir := filepath.Dir(filename)
	backendDir := filepath.Dir(filepath.Dir(handlerDir))
	fontDir := filepath.Join(backendDir, "font")
	fontRegularPath := filepath.Join(fontDir, "BeVietnamPro-Regular.ttf")
	fontBoldPath := filepath.Join(fontDir, "BeVietnamPro-Bold.ttf")
	fontMediumPath := filepath.Join(fontDir, "BeVietnamPro-Medium.ttf")

	// Convert sang absolute path
	absFontRegular, _ := filepath.Abs(fontRegularPath)
	absFontBold, _ := filepath.Abs(fontBoldPath)
	absFontMedium, _ := filepath.Abs(fontMediumPath)

	// Tạo PDF A4 (210x297mm) - mỗi hành khách 1 trang
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{
		PageSize: gopdf.Rect{
			W: 210,
			H: 297,
		},
		Unit: gopdf.UnitMM,
	})

	// Thêm font
	fontName := "BeVietnamPro"
	if err := pdf.AddTTFFont(fontName, absFontRegular); err != nil {
		c.JSON(500, gin.H{"error": "Không thể load font Regular", "path": absFontRegular, "details": err.Error()})
		return
	}
	if err := pdf.AddTTFFontWithOption(fontName+"Bold", absFontBold, gopdf.TtfOption{
		UseKerning: true,
	}); err != nil {
		c.JSON(500, gin.H{"error": "Không thể load font Bold", "path": absFontBold, "details": err.Error()})
		return
	}
	if err := pdf.AddTTFFont(fontName+"Medium", absFontMedium); err != nil {
		// Fallback nếu không có Medium, dùng Bold
		fontName += "Bold"
	}

	/* ================== TẠO VÉ CHO TỪNG HÀNH KHÁCH ================== */

	for idx, ticket := range tickets {
		// Thêm trang mới cho mỗi hành khách (trừ hành khách đầu tiên)
		if idx > 0 {
			pdf.AddPage()
		} else {
			pdf.AddPage()
		}

		drawTicketPage(&pdf, ticket, booking, bookingCode, fontName, idx+1, len(tickets))
	}

	/* ================== EXPORT ================== */

	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		c.JSON(500, gin.H{"error": "Không thể tạo PDF", "details": err.Error()})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header(
		"Content-Disposition",
		fmt.Sprintf("attachment; filename=Ve-Dien-Tu-%s.pdf", bookingCode),
	)
	c.Data(http.StatusOK, "application/pdf", buf.Bytes())
}

/* ================== VẼ TRANG VÉ CHO 1 HÀNH KHÁCH ================== */

func drawTicketPage(pdf *gopdf.GoPdf, ticket db.GetTicketsByBookingIDRow, booking db.GetBookingByIdRow, bookingCode string, fontName string, passengerNum int, totalPassengers int) {
	/* ================== BACKGROUND ================== */
	// Nền trắng sạch
	pdf.SetFillColor(255, 255, 255)
	pdf.RectFromUpperLeft(0, 0, 210, 297)

	/* ================== BORDER ================== */
	// Border đơn giản, tinh tế
	pdf.SetStrokeColor(220, 220, 220)
	pdf.SetLineWidth(1)
	pdf.RectFromUpperLeft(10, 10, 190, 277)

	/* ================== HEADER ================== */
	// Tính toán để căn giữa
	pageWidth := 210.0
	contentWidth := 170.0
	startX := (pageWidth - contentWidth) / 2

	// Header background - màu xám đậm, đẹp hơn
	pdf.SetFillColor(30, 41, 59)
	pdf.RectFromUpperLeft(startX, 20, contentWidth, 45)

	// Logo/Title area - căn giữa, lớn và rõ ràng
	pdf.SetFont(fontName+"Bold", "", 28)
	pdf.SetTextColor(255, 255, 255)
	titleText := "VÉ ĐIỆN TỬ"
	titleWidth := getTextWidth(pdf, titleText)
	pdf.SetX(startX + (contentWidth-titleWidth)/2)
	pdf.SetY(35)
	pdf.Cell(nil, titleText)

	// Booking code - căn giữa, rõ ràng
	pdf.SetFont(fontName+"Medium", "", 12)
	pdf.SetTextColor(255, 255, 255)
	codeText := "Mã: " + bookingCode
	codeWidth := getTextWidth(pdf, codeText)
	pdf.SetX(startX + (contentWidth-codeWidth)/2)
	pdf.SetY(50)
	pdf.Cell(nil, codeText)

	// Số thứ tự hành khách - góc phải trên
	pdf.SetFont(fontName, "", 10)
	pdf.SetTextColor(200, 200, 200)
	passengerText := fmt.Sprintf("%d/%d", passengerNum, totalPassengers)
	passengerWidth := getTextWidth(pdf, passengerText)
	pdf.SetX(startX + contentWidth - passengerWidth - 12)
	pdf.SetY(35)
	pdf.Cell(nil, passengerText)

	/* ================== TOUR INFORMATION ================== */
	y := 80.0

	// Card background cho tour name
	pdf.SetFillColor(248, 250, 252)
	pdf.RectFromUpperLeft(startX, y-5, contentWidth, 25)

	// Tour name - lớn và nổi bật, căn giữa chính xác
	tourName := strings.ToUpper(ticket.TenTour)
	if len(tourName) > 65 {
		tourName = tourName[:62] + "..."
	}
	pdf.SetFont(fontName+"Bold", "", 18)
	pdf.SetTextColor(15, 23, 42)
	// Căn giữa: tính center point và offset text
	centerX := startX + contentWidth/2
	// Ước lượng width: font 18pt ≈ 1.08mm per character cho Vietnamese
	estimatedWidth := float64(len(tourName)) * 1.08
	pdf.SetX(centerX - estimatedWidth/2)
	pdf.SetY(y + 8)
	pdf.Cell(nil, tourName)

	// Line separator đơn giản
	pdf.SetStrokeColor(226, 232, 240)
	pdf.SetLineWidth(1)
	pdf.Line(startX, y+20, startX+contentWidth, y+20)

	/* ================== PASSENGER INFORMATION ================== */
	y += 35.0

	// Card background cho thông tin hành khách - tăng height để không bị che
	pdf.SetFillColor(255, 255, 255)
	pdf.SetStrokeColor(226, 232, 240)
	pdf.SetLineWidth(1)
	pdf.RectFromUpperLeft(startX, y, contentWidth, 85)

	// Section title - rõ ràng
	sectionTitle := "THÔNG TIN HÀNH KHÁCH"
	pdf.SetFont(fontName+"Bold", "", 14)
	pdf.SetTextColor(15, 23, 42)
	pdf.SetX(startX + 10)
	pdf.SetY(y + 8)
	pdf.Cell(nil, sectionTitle)

	// Line dưới title
	pdf.SetStrokeColor(226, 232, 240)
	pdf.SetLineWidth(1)
	pdf.Line(startX+10, y+15, startX+contentWidth-10, y+15)

	y += 25.0

	// Helper function để vẽ field với label và value rõ ràng
	drawInfoField := func(label, value string, x, y float64, isImportant bool) {
		// Label - nhỏ, màu xám
		pdf.SetFont(fontName, "", 9)
		pdf.SetTextColor(100, 116, 139)
		pdf.SetX(x)
		pdf.SetY(y)
		pdf.Cell(nil, label)

		// Value - lớn hơn, màu đen, rõ ràng
		if isImportant {
			pdf.SetFont(fontName+"Bold", "", 16)
			pdf.SetTextColor(15, 23, 42)
		} else {
			pdf.SetFont(fontName+"Medium", "", 12)
			pdf.SetTextColor(30, 41, 59)
		}
		pdf.SetX(x)
		pdf.SetY(y + 8)
		// Wrap text nếu quá dài
		if len(value) > 40 {
			value = value[:37] + "..."
		}
		pdf.Cell(nil, value)
	}

	// Họ tên - quan trọng nhất, lớn và rõ, căn giữa
	nameX := startX + (contentWidth-140)/2
	drawInfoField("Họ và tên", ticket.HoTen, nameX, y, true)

	// Grid layout cho các thông tin khác - 2 cột rõ ràng
	y += 30.0
	leftColX := startX + 15
	rightColX := startX + contentWidth/2 + 10
	fieldSpacing := 24.0 // Tăng spacing để không bị che

	// Ngày sinh
	birthDate := "N/A"
	if ticket.NgaySinh.Valid {
		birthDate = ticket.NgaySinh.Time.Format("02/01/2006")
	}
	drawInfoField("Ngày sinh", birthDate, leftColX, y, false)

	// Số giấy tờ tùy thân
	soGiayTo := "N/A"
	if ticket.SoGiayToTuyThanh != nil && *ticket.SoGiayToTuyThanh != "" {
		soGiayTo = *ticket.SoGiayToTuyThanh
	}
	drawInfoField("Số giấy tờ", soGiayTo, rightColX, y, false)

	y += fieldSpacing

	// Loại khách
	loaiKhach := "Người lớn"
	if ticket.LoaiKhach != nil {
		if *ticket.LoaiKhach == "tre_em" {
			loaiKhach = "Trẻ em"
		} else if *ticket.LoaiKhach == "em_be" {
			loaiKhach = "Em bé"
		}
	}
	drawInfoField("Loại khách", loaiKhach, leftColX, y, false)

	// Quốc tịch
	quocTich := "Việt Nam"
	if ticket.QuocTich != nil && *ticket.QuocTich != "" {
		quocTich = *ticket.QuocTich
	}
	drawInfoField("Quốc tịch", quocTich, rightColX, y, false)

	/* ================== TOUR DETAILS ================== */
	y += 90.0

	// Card background cho thông tin tour
	pdf.SetFillColor(255, 255, 255)
	pdf.SetStrokeColor(226, 232, 240)
	pdf.SetLineWidth(1)
	pdf.RectFromUpperLeft(startX, y, contentWidth, 70)

	// Section title - rõ ràng
	sectionTitle = "THÔNG TIN TOUR"
	pdf.SetFont(fontName+"Bold", "", 14)
	pdf.SetTextColor(15, 23, 42)
	pdf.SetX(startX + 10)
	pdf.SetY(y + 8)
	pdf.Cell(nil, sectionTitle)

	// Line dưới title
	pdf.SetStrokeColor(226, 232, 240)
	pdf.SetLineWidth(1)
	pdf.Line(startX+10, y+15, startX+contentWidth-10, y+15)

	y += 25.0

	// Ngày khởi hành
	ngayKhoiHanh := formatDate(ticket.NgayKhoiHanh)
	drawInfoField("Ngày khởi hành", ngayKhoiHanh, leftColX, y, false)

	// Ngày kết thúc
	ngayKetThuc := formatDate(ticket.NgayKetThuc)
	drawInfoField("Ngày kết thúc", ngayKetThuc, rightColX, y, false)

	y += fieldSpacing

	// Số người
	soNguoi := "N/A"
	if booking.SoNguoiLon != nil {
		soNguoi = fmt.Sprintf("%d người lớn", *booking.SoNguoiLon)
		if booking.SoTreEm != nil && *booking.SoTreEm > 0 {
			soNguoi += fmt.Sprintf(", %d trẻ em", *booking.SoTreEm)
		}
	}
	drawInfoField("Số khách", soNguoi, leftColX, y, false)

	// Trạng thái với badge style
	trangThai := "CHỜ XÁC NHẬN"
	if booking.TrangThai.Valid {
		trangThai = strings.ReplaceAll(strings.ToUpper(string(booking.TrangThai.TrangThaiDatCho)), "_", " ")
	}

	// Label
	pdf.SetFont(fontName, "", 9)
	pdf.SetTextColor(100, 116, 139)
	pdf.SetX(rightColX)
	pdf.SetY(y)
	pdf.Cell(nil, "Trạng thái")

	// Badge background
	var r, g, b uint8 = 251, 191, 36 // Vàng mặc định
	if strings.Contains(trangThai, "THANH TOAN") || strings.Contains(trangThai, "XAC NHAN") {
		r, g, b = 34, 197, 94 // Xanh lá
	} else if strings.Contains(trangThai, "HUY") {
		r, g, b = 239, 68, 68 // Đỏ
	}

	pdf.SetFillColor(r, g, b)
	pdf.RectFromUpperLeft(rightColX, y+8, 60, 12)

	// Status text
	pdf.SetFont(fontName+"Bold", "", 10)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetX(rightColX + 5)
	pdf.SetY(y + 10)
	pdf.Cell(nil, trangThai)

	/* ================== PRICE SECTION ================== */
	y = 195.0

	// Box cho giá tiền - căn giữa, nổi bật
	priceBoxWidth := 150.0
	priceBoxX := startX + (contentWidth-priceBoxWidth)/2
	priceBoxHeight := 60.0

	// Background với border rõ ràng
	pdf.SetFillColor(248, 250, 252)
	pdf.SetStrokeColor(30, 41, 59)
	pdf.SetLineWidth(2)
	pdf.RectFromUpperLeft(priceBoxX, y, priceBoxWidth, priceBoxHeight)

	// Label - căn giữa
	pdf.SetFont(fontName+"Medium", "", 11)
	pdf.SetTextColor(100, 116, 139)
	labelText := "GIÁ VÉ"
	labelWidth := getTextWidth(pdf, labelText)
	pdf.SetX(priceBoxX + (priceBoxWidth-labelWidth)/2)
	pdf.SetY(y + 10)
	pdf.Cell(nil, labelText)

	// Tính giá vé cho hành khách này
	var ticketPrice float64
	if ticket.TongTien.Valid {
		// Giá vé = tổng tiền / số hành khách
		totalValue, _ := ticket.TongTien.Float64Value()
		totalPassengers := int32(0)
		if booking.SoNguoiLon != nil {
			totalPassengers += *booking.SoNguoiLon
		}
		if booking.SoTreEm != nil {
			totalPassengers += *booking.SoTreEm
		}
		if totalPassengers > 0 {
			ticketPrice = totalValue.Float64 / float64(totalPassengers)
		}
	}

	// Amount - rất lớn và rõ ràng, căn giữa
	currency := "VND"
	if ticket.DonViTienTe != nil {
		currency = *ticket.DonViTienTe
	}
	priceText := formatCurrency(ticketPrice, currency)

	pdf.SetFont(fontName+"Bold", "", 30)
	pdf.SetTextColor(15, 23, 42) // Màu đen đậm
	priceWidth := getTextWidth(pdf, priceText)
	pdf.SetX(priceBoxX + (priceBoxWidth-priceWidth)/2)
	pdf.SetY(y + 28)
	pdf.Cell(nil, priceText)

	// Note về tổng tiền - căn giữa, nhỏ
	pdf.SetFont(fontName, "", 8)
	pdf.SetTextColor(148, 163, 184)
	totalValue, _ := ticket.TongTien.Float64Value()
	totalText := formatCurrency(totalValue.Float64, currency)
	noteText := fmt.Sprintf("Tổng: %s", totalText)
	noteWidth := getTextWidth(pdf, noteText)
	pdf.SetX(priceBoxX + (priceBoxWidth-noteWidth)/2)
	pdf.SetY(y + 50)
	pdf.Cell(nil, noteText)

	/* ================== FOOTER ================== */
	y = 265.0

	// Line separator đơn giản
	pdf.SetStrokeColor(226, 232, 240)
	pdf.SetLineWidth(1)
	pdf.Line(startX, y, startX+contentWidth, y)

	// Footer text - căn giữa, rõ ràng
	footerText := "Vé điện tử này có giá trị pháp lý như vé giấy"
	pdf.SetFont(fontName, "", 9)
	pdf.SetTextColor(100, 116, 139)
	footerWidth := getTextWidth(pdf, footerText)
	pdf.SetX(startX + (contentWidth-footerWidth)/2)
	pdf.SetY(y + 8)
	pdf.Cell(nil, footerText)

	// Contact info - căn giữa
	contactText := "support@travia.vn | Hotline: 1900-xxxx"
	pdf.SetFont(fontName, "", 8)
	pdf.SetTextColor(148, 163, 184)
	contactWidth := getTextWidth(pdf, contactText)
	pdf.SetX(startX + (contentWidth-contactWidth)/2)
	pdf.SetY(y + 16)
	pdf.Cell(nil, contactText)

	// Print date - góc phải dưới
	printDate := time.Now().Format("02/01/2006 15:04")
	dateText := "In: " + printDate
	pdf.SetFont(fontName, "", 8)
	pdf.SetTextColor(148, 163, 184)
	pdf.SetX(startX + contentWidth - getTextWidth(pdf, dateText) - 5)
	pdf.SetY(y + 16)
	pdf.Cell(nil, dateText)
}

/* ================== DECORATIVE FUNCTIONS ================== */

// Vẽ hoa văn nền (watermark pattern)
func drawDecorativePattern(pdf *gopdf.GoPdf) {
	// Pattern nhạt ở góc - dùng hình vuông nhỏ thay vì circle
	pdf.SetFillColor(240, 245, 250)

	// Vẽ các hình vuông nhỏ làm pattern
	for i := 0; i < 20; i++ {
		x := float64(10 + (i%5)*40)
		y := float64(50 + (i/5)*50)
		pdf.RectFromUpperLeft(x-1, y-1, 2, 2)
	}

	// Vẽ pattern ở góc dưới
	for i := 0; i < 15; i++ {
		x := float64(20 + (i%5)*35)
		y := float64(250 + (i/5)*30)
		pdf.RectFromUpperLeft(x-0.75, y-0.75, 1.5, 1.5)
	}
}

// Vẽ border decorative với hoa văn
func drawDecorativeBorder(pdf *gopdf.GoPdf) {
	// Border chính - màu xám đậm
	pdf.SetStrokeColor(51, 65, 85)
	pdf.SetLineWidth(2)
	pdf.RectFromUpperLeft(8, 8, 194, 281)

	// Border trong
	pdf.SetStrokeColor(200, 200, 200)
	pdf.SetLineWidth(0.5)
	pdf.RectFromUpperLeft(10, 10, 190, 277)

	// Corner decorations
	cornerSize := 15.0
	cornerThickness := 2.0

	// Góc trên trái
	pdf.SetStrokeColor(51, 65, 85)
	pdf.SetLineWidth(cornerThickness)
	pdf.Line(8, 8, 8+cornerSize, 8)
	pdf.Line(8, 8, 8, 8+cornerSize)

	// Góc trên phải
	pdf.Line(202-cornerSize, 8, 202, 8)
	pdf.Line(202, 8, 202, 8+cornerSize)

	// Góc dưới trái
	pdf.Line(8, 289-cornerSize, 8, 289)
	pdf.Line(8, 289, 8+cornerSize, 289)

	// Góc dưới phải
	pdf.Line(202-cornerSize, 289, 202, 289)
	pdf.Line(202, 289-cornerSize, 202, 289)
}

// Vẽ corner decoration cho header
func drawCornerDecoration(pdf *gopdf.GoPdf, x, y, w, h float64) {
	// Vẽ các đường trang trí ở góc
	pdf.SetStrokeColor(255, 255, 255)
	pdf.SetLineWidth(1)

	// Góc trên trái
	pdf.Line(x, y, x+10, y)
	pdf.Line(x, y, x, y+10)

	// Góc trên phải
	pdf.Line(x+w-10, y, x+w, y)
	pdf.Line(x+w, y, x+w, y+10)

	// Góc dưới trái
	pdf.Line(x, y+h-10, x, y+h)
	pdf.Line(x, y+h, x+10, y+h)

	// Góc dưới phải
	pdf.Line(x+w-10, y+h, x+w, y+h)
	pdf.Line(x+w, y+h-10, x+w, y+h)
}

// Vẽ icon tour (decorative)
func drawTourIcon(pdf *gopdf.GoPdf, x, y float64) {
	// Vẽ icon đơn giản bằng các hình học - màu xám đậm
	pdf.SetFillColor(51, 65, 85)

	// Vẽ hình vuông nhỏ (thay vì circle)
	pdf.RectFromUpperLeft(x-2, y-2, 4, 4)

	// Vẽ các đường trang trí
	pdf.SetStrokeColor(51, 65, 85)
	pdf.SetLineWidth(1)
	pdf.Line(x-5, y, x+5, y)
	pdf.Line(x, y-5, x, y+5)
}

// Vẽ decorative line với pattern
func drawDecorativeLine(pdf *gopdf.GoPdf, x, y, width float64) {
	// Line chính
	pdf.SetStrokeColor(226, 232, 240)
	pdf.SetLineWidth(1)
	pdf.Line(x, y, x+width, y)

	// Pattern dots - dùng hình vuông nhỏ
	pdf.SetFillColor(200, 200, 200)
	dotSpacing := width / 20
	for i := 0; i < 20; i++ {
		dotX := x + float64(i)*dotSpacing
		pdf.RectFromUpperLeft(dotX-0.5, y-0.5, 1, 1)
	}
}

// Vẽ decorative underline
func drawDecorativeUnderline(pdf *gopdf.GoPdf, x, y, width float64) {
	// Line chính - màu xám đậm
	pdf.SetStrokeColor(51, 65, 85)
	pdf.SetLineWidth(2)
	pdf.Line(x, y, x+width, y)

	// Thêm các điểm nhỏ - dùng hình vuông
	pdf.SetFillColor(51, 65, 85)
	for i := 0; i < 5; i++ {
		dotX := x + float64(i)*width/4
		pdf.RectFromUpperLeft(dotX-1, y-1, 2, 2)
	}
}

// Vẽ decorative box với pattern
func drawDecorativeBox(pdf *gopdf.GoPdf, x, y, w, h float64) {
	// Border chính - màu xám đậm
	pdf.SetStrokeColor(51, 65, 85)
	pdf.SetLineWidth(2)
	pdf.RectFromUpperLeft(x, y, w, h)

	// Border trong
	pdf.SetStrokeColor(226, 232, 240)
	pdf.SetLineWidth(0.5)
	pdf.RectFromUpperLeft(x+2, y+2, w-4, h-4)

	// Corner decorations
	cornerSize := 8.0
	pdf.SetStrokeColor(51, 65, 85)
	pdf.SetLineWidth(1.5)

	// Góc trên trái
	pdf.Line(x, y, x+cornerSize, y)
	pdf.Line(x, y, x, y+cornerSize)

	// Góc trên phải
	pdf.Line(x+w-cornerSize, y, x+w, y)
	pdf.Line(x+w, y, x+w, y+cornerSize)

	// Góc dưới trái
	pdf.Line(x, y+h-cornerSize, x, y+h)
	pdf.Line(x, y+h, x+cornerSize, y+h)

	// Góc dưới phải
	pdf.Line(x+w-cornerSize, y+h, x+w, y+h)
	pdf.Line(x+w, y+h-cornerSize, x+w, y+h)
}

// Helper để tính width của text (ước lượng)
func getTextWidth(pdf *gopdf.GoPdf, text string) float64 {
	// Estimate width based on character count
	// Average character width in mm for Vietnamese characters: ~0.6mm per character for size 10
	// For Vietnamese text, multiply by 1.1 for wider characters
	charWidth := 0.6 // mm per character for size 10

	if len(text) > 0 {
		// Estimate based on text length
		// Adjust multiplier based on typical font sizes used in the document
		return float64(len(text)) * charWidth * 1.1
	}
	return 0
}

/* ================== HELPERS ================== */

func formatDate(d pgtype.Date) string {
	if !d.Valid {
		return "N/A"
	}
	return d.Time.Format("02/01/2006")
}

func formatCurrency(amount float64, currency string) string {
	// Format số với dấu chấm phân cách hàng nghìn
	s := fmt.Sprintf("%.0f", amount)
	var parts []string
	for i := len(s); i > 0; i -= 3 {
		start := i - 3
		if start < 0 {
			start = 0
		}
		parts = append([]string{s[start:i]}, parts...)
	}
	formatted := strings.Join(parts, ".")

	// Thêm đơn vị tiền tệ
	if currency == "VND" || currency == "" {
		return formatted + " ₫"
	}
	return formatted + " " + currency
}
