package handler

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"travia.backend/api/utils"
	db "travia.backend/db/sqlc"
)

const (
	// VNPayPaymentExpirationTime là thời gian hết hạn của payment URL VNPay (30 phút)
	// VNPay cho phép tối đa 30 phút cho giao dịch thanh toán
	VNPayPaymentExpirationTime = 30 * time.Minute
)

// CreateVNPayPaymentURL tạo URL thanh toán VNPay
// @Summary Tạo URL thanh toán VNPay
// @Description Tạo URL thanh toán VNPay cho booking
// @Tags Payment
// @Accept json
// @Produce json
// @Param request body CreateVNPayPaymentRequest true "Payment Request"
// @Security ApiKeyAuth
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /payment/vnpay/create [post]
func (s *Server) CreateVNPayPaymentURL(c *gin.Context) {
	// Xác thực user
	claims, ok := c.Get("claims")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	claimsMap, ok := claims.(*utils.JwtClams)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req CreateVNPayPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log request để debug
	fmt.Printf("=== CreateVNPayPaymentURL Request ===\n")
	fmt.Printf("Booking ID: %d\n", req.BookingID)
	fmt.Printf("Return URL from request: %s\n", req.ReturnURL)
	fmt.Printf("=====================================\n")

	// Validate và fix returnURL nếu là "string"
	if req.ReturnURL == "string" || req.ReturnURL == "" {
		fmt.Printf("⚠️  Return URL is empty or 'string', using config default\n")
		req.ReturnURL = "" // Sẽ dùng config default
	}

	// Kiểm tra booking tồn tại và thuộc về user
	booking, err := s.z.GetBookingById(c.Request.Context(), req.BookingID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}

	// Kiểm tra booking thuộc về user
	if booking.NguoiDungID.String() != claimsMap.Id.String() {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	// Kiểm tra booking chưa thanh toán
	if booking.TrangThai.Valid && (booking.TrangThai.TrangThaiDatCho == "da_thanh_toan" || booking.TrangThai.TrangThaiDatCho == "hoan_thanh") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Booking already paid"})
		return
	}

	// Kiểm tra booking status hợp lệ (chỉ cho phép thanh toán khi ở trạng thái chờ xác nhận hoặc đã xác nhận)
	if booking.TrangThai.Valid && booking.TrangThai.TrangThaiDatCho != "cho_xac_nhan" && booking.TrangThai.TrangThaiDatCho != "da_xac_nhan" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Booking không ở trạng thái hợp lệ để thanh toán",
			"trang_thai": booking.TrangThai.TrangThaiDatCho,
		})
		return
	}

	// Get client IP for vnp_IpAddr
	clientIP := c.ClientIP()
	if clientIP == "" || clientIP == "::1" {
		clientIP = "127.0.0.1"
	}

	// Lấy múi giờ VN để so sánh thời gian chính xác
	location, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	now := time.Now().In(location)

	existingTransactions, err := s.z.GetTransactionsByBooking(c.Request.Context(), &booking.ID)
	if err == nil && len(existingTransactions) > 0 {
		for _, tx := range existingTransactions {
			if tx.TrangThai.Valid && tx.TrangThai.TrangThaiThanhToan == "dang_cho_thanh_toan" {
				if tx.NgayTao.Valid {
					// VNPayPaymentExpirationTime nên là 15 * time.Minute
					if now.Before(tx.NgayTao.Time.In(location).Add(15 * time.Minute)) {
						// REUSE: Nếu còn hạn, tạo URL mới với mã cũ nhưng Time mới
						paymentURL, err := s.createVNPayPaymentURL(booking, tx.MaGiaoDichNoiBo, req.ReturnURL, clientIP)
						if err == nil {
							c.JSON(http.StatusOK, gin.H{"payment_url": paymentURL, "transaction_code": tx.MaGiaoDichNoiBo})
							return
						}
					}
				}
			}
		}
	}

	// CREATE NEW: Tạo mã mới hoàn toàn để tránh xung đột "Giao dịch cũ" trên VNPay
	transactionCode := fmt.Sprintf("TRAVIA%d%d", booking.ID, now.Unix())
	paymentURL, err := s.createVNPayPaymentURL(booking, transactionCode, req.ReturnURL, clientIP)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{
				"error":   "Không thể tạo liên kết thanh toán",
				"message": err.Error(),
			})
		return
	}

	// Lưu vào DB (Giữ nguyên logic CreateTransaction cũ của bạn)
	congThanhToanID := "vnpay"
	noiDungChuyenKhoan := fmt.Sprintf("Thanh toan don dat cho #%d qua VNPay", booking.ID)
	bookingID := booking.ID

	_, err = s.z.CreateTransaction(c.Request.Context(), db.CreateTransactionParams{
		DatChoID:           &bookingID,
		NguoiDungID:        booking.NguoiDungID,
		MaGiaoDichNoiBo:    transactionCode,
		CongThanhToanID:    &congThanhToanID,
		SoTien:             booking.TongTien,
		NoiDungChuyenKhoan: &noiDungChuyenKhoan,
	})
	if err != nil {
		fmt.Printf("ERROR: Failed to create transaction: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể tạo giao dịch thanh toán"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"payment_url":      paymentURL,
		"transaction_code": transactionCode,
		"booking_id":       booking.ID,
	})
}

// VNPayCallback xử lý callback từ VNPay (Return URL)
// @Summary VNPay Return Callback
// @Description Xử lý callback khi user quay lại từ VNPay
// @Tags Payment
// @Accept json
// @Produce json
// @Param vnp_Amount query string true "Amount"
// @Param vnp_BankCode query string false "Bank Code"
// @Param vnp_BankTranNo query string false "Bank Transaction No"
// @Param vnp_CardType query string false "Card Type"
// @Param vnp_OrderInfo query string true "Order Info"
// @Param vnp_PayDate query string true "Pay Date"
// @Param vnp_ResponseCode query string true "Response Code"
// @Param vnp_TmnCode query string true "TMN Code"
// @Param vnp_TransactionNo query string true "Transaction No"
// @Param vnp_TransactionStatus query string true "Transaction Status"
// @Param vnp_TxnRef query string true "Transaction Ref"
// @Param vnp_SecureHash query string true "Secure Hash"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Router /payment/vnpay/return [get]
func (s *Server) VNPayCallback(c *gin.Context) {
	// Log để debug
	fmt.Printf("=== VNPay Return URL Callback ===\n")
	fmt.Printf("Request URL: %s\n", c.Request.URL.String())
	fmt.Printf("Query Params: %v\n", c.Request.URL.Query())
	fmt.Printf("Client IP: %s\n", c.ClientIP())
	fmt.Printf("User Agent: %s\n", c.Request.UserAgent())
	fmt.Printf("================================\n")

	// Lấy tất cả query parameters
	queryParams := c.Request.URL.Query()

	// Verify signature
	if !s.verifyVNPaySignature(queryParams) {
		fmt.Printf("ERROR: Invalid signature in Return URL callback\n")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid signature"})
		return
	}
	fmt.Printf("Signature verified successfully\n")

	// Parse response code
	responseCode := queryParams.Get("vnp_ResponseCode")
	transactionStatus := queryParams.Get("vnp_TransactionStatus")
	txnRef := queryParams.Get("vnp_TxnRef")
	transactionNo := queryParams.Get("vnp_TransactionNo")

	// Extract booking ID from transaction code (format: TRAVIA{bookingID}{timestamp})
	// Tìm booking ID từ transaction code
	transaction, err := s.z.GetTransactionByCode(c.Request.Context(), txnRef)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	if transaction.DatChoID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction"})
		return
	}
	bookingID := int(*transaction.DatChoID)

	// Kiểm tra booking tồn tại và ở trạng thái hợp lệ
	booking, err := s.z.GetBookingById(c.Request.Context(), int32(bookingID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}

	// Kiểm tra booking status hợp lệ
	if !booking.TrangThai.Valid || (booking.TrangThai.TrangThaiDatCho != "cho_xac_nhan" && booking.TrangThai.TrangThaiDatCho != "da_xac_nhan") {
		returnURL := s.config.VNPayConfig.ReturnURL
		c.Redirect(http.StatusFound, fmt.Sprintf("%s?status=failed&booking_id=%d&transaction_code=%s&error=invalid_booking_status", returnURL, bookingID, txnRef))
		return
	}

	// Kiểm tra response code
	if responseCode == "00" && transactionStatus == "00" {
		fmt.Printf("Payment SUCCESS - ResponseCode: %s, TransactionStatus: %s\n", responseCode, transactionStatus)
		// Thanh toán thành công
		// Cập nhật transaction
		transaction, err = s.z.GetTransactionByCode(c.Request.Context(), txnRef)
		if err == nil {
			trangThai := db.NullTrangThaiThanhToan{
				TrangThaiThanhToan: "thanh_cong",
				Valid:              true,
			}
			_, err = s.z.UpdateTransactionStatus(c.Request.Context(), db.UpdateTransactionStatusParams{
				ID:          transaction.ID,
				TrangThai:   trangThai,
				MaThamChieu: &transactionNo,
			})
			if err != nil {
				// Log error nhưng vẫn tiếp tục cập nhật booking
				fmt.Printf("ERROR: Failed to update transaction status: %v\n", err)
			} else {
				fmt.Printf("Transaction status updated successfully - ID: %d\n", transaction.ID)
			}
		}

		// Cập nhật booking status
		phuongThuc := "vnpay"
		updatedBooking, err := s.z.UpdateBookingPaymentStatus(c.Request.Context(), db.UpdateBookingPaymentStatusParams{
			ID:                  int32(bookingID),
			PhuongThucThanhToan: &phuongThuc,
		})
		if err != nil {
			// Log error
			fmt.Printf("ERROR: Failed to update booking payment status: %v\n", err)
			returnURL := s.config.VNPayConfig.ReturnURL
			c.Redirect(http.StatusFound, fmt.Sprintf("%s?status=failed&booking_id=%d&transaction_code=%s&error=update_failed", returnURL, bookingID, txnRef))
			return
		}
		fmt.Printf("Booking payment status updated successfully - Booking ID: %d\n", bookingID)

		// Kiểm tra xem booking đã được cập nhật thành công chưa
		if !updatedBooking.TrangThai.Valid || updatedBooking.TrangThai.TrangThaiDatCho != "da_thanh_toan" {
			fmt.Printf("WARNING: Booking status not updated correctly. Booking ID: %d, Status: %v\n", bookingID, updatedBooking.TrangThai)
		}

		// Redirect về frontend với success
		returnURL := s.config.VNPayConfig.ReturnURL
		fmt.Printf("Redirecting to frontend: %s?status=success&booking_id=%d&transaction_code=%s\n", returnURL, bookingID, txnRef)
		c.Redirect(http.StatusFound, fmt.Sprintf("%s?status=success&booking_id=%d&transaction_code=%s", returnURL, bookingID, txnRef))
		return
	}

	// Thanh toán thất bại
	fmt.Printf("Payment FAILED - ResponseCode: %s, TransactionStatus: %s\n", responseCode, transactionStatus)
	transaction, err = s.z.GetTransactionByCode(c.Request.Context(), txnRef)
	if err == nil {
		trangThai := db.NullTrangThaiThanhToan{
			TrangThaiThanhToan: "that_bai",
			Valid:              true,
		}
		_, err = s.z.UpdateTransactionStatus(c.Request.Context(), db.UpdateTransactionStatusParams{
			ID:        transaction.ID,
			TrangThai: trangThai,
		})
		if err != nil {
			fmt.Printf("ERROR: Failed to update transaction status to failed: %v\n", err)
		} else {
			fmt.Printf("Transaction status updated to failed - ID: %d\n", transaction.ID)
		}
	}

	returnURL := s.config.VNPayConfig.ReturnURL
	fmt.Printf("Redirecting to frontend: %s?status=failed&booking_id=%d&transaction_code=%s&error_code=%s\n", returnURL, bookingID, txnRef, responseCode)
	c.Redirect(http.StatusFound, fmt.Sprintf("%s?status=failed&booking_id=%d&transaction_code=%s&error_code=%s", returnURL, bookingID, txnRef, responseCode))
}

// VNPayIPN xử lý IPN (Instant Payment Notification) từ VNPay
// @Summary VNPay IPN Callback
// @Description Xử lý IPN callback từ VNPay server (tương đương payment_ipn trong Python)
// @Tags Payment
// @Accept json
// @Produce json
// @Param vnp_Amount formData string true "Amount"
// @Param vnp_BankCode formData string false "Bank Code"
// @Param vnp_BankTranNo formData string false "Bank Transaction No"
// @Param vnp_CardType formData string false "Card Type"
// @Param vnp_OrderInfo formData string true "Order Info"
// @Param vnp_PayDate formData string true "Pay Date"
// @Param vnp_ResponseCode formData string true "Response Code"
// @Param vnp_TmnCode formData string true "TMN Code"
// @Param vnp_TransactionNo formData string true "Transaction No"
// @Param vnp_TransactionStatus formData string true "Transaction Status"
// @Param vnp_TxnRef formData string true "Transaction Ref"
// @Param vnp_SecureHash formData string true "Secure Hash"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Router /payment/vnpay/ipn [post]
func (s *Server) VNPayIPN(c *gin.Context) {
	// Log để debug
	fmt.Printf("=== VNPay IPN Callback ===\n")
	fmt.Printf("Request URL: %s\n", c.Request.URL.String())
	fmt.Printf("Method: %s\n", c.Request.Method)
	fmt.Printf("Content-Type: %s\n", c.Request.Header.Get("Content-Type"))
	fmt.Printf("Client IP: %s\n", c.ClientIP())
	fmt.Printf("User Agent: %s\n", c.Request.UserAgent())

	// Lấy tất cả form data (VNPay gửi qua POST form-data)
	inputData := c.Request.PostForm
	fmt.Printf("Form Data: %v\n", inputData)
	fmt.Printf("========================\n")

	// Kiểm tra nếu không có data
	if len(inputData) == 0 {
		fmt.Printf("ERROR: No form data received in IPN\n")
		c.JSON(http.StatusOK, gin.H{
			"RspCode": "99",
			"Message": "Invalid request",
		})
		return
	}

	// Verify signature
	if !s.verifyVNPaySignature(inputData) {
		fmt.Printf("ERROR: Invalid signature in IPN callback\n")
		// Invalid Signature
		c.JSON(http.StatusOK, gin.H{
			"RspCode": "97",
			"Message": "Invalid Signature",
		})
		return
	}
	fmt.Printf("IPN Signature verified successfully\n")

	// Parse các tham số từ VNPay (tương đương Python: order_id, amount, order_desc, vnp_TransactionNo, etc.)
	txnRef := inputData.Get("vnp_TxnRef")               // order_id
	amountStr := inputData.Get("vnp_Amount")            // amount
	_ = inputData.Get("vnp_OrderInfo")                  // order_desc
	transactionNo := inputData.Get("vnp_TransactionNo") // vnp_TransactionNo
	responseCode := inputData.Get("vnp_ResponseCode")   // vnp_ResponseCode
	// Các tham số khác (có thể dùng để log hoặc debug)
	_ = inputData.Get("vnp_TmnCode")  // tmnCode
	_ = inputData.Get("vnp_PayDate")  // payDate
	_ = inputData.Get("vnp_BankCode") // bankCode
	_ = inputData.Get("vnp_CardType") // cardType

	// Tìm transaction trong database
	transaction, err := s.z.GetTransactionByCode(c.Request.Context(), txnRef)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"RspCode": "01",
			"Message": "Transaction not found",
		})
		return
	}

	// Kiểm tra nếu đã xử lý rồi (firstTimeUpdate = false)
	firstTimeUpdate := !(transaction.TrangThai.Valid && transaction.TrangThai.TrangThaiThanhToan == "thanh_cong")

	if !firstTimeUpdate {
		// Already Update
		c.JSON(http.StatusOK, gin.H{
			"RspCode": "02",
			"Message": "Order Already Update",
		})
		return
	}

	// Kiểm tra amount (totalAmount check)
	// Convert amount từ VNPay (amount * 100) về số tiền thực tế
	amountFromVNPay, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"RspCode": "03",
			"Message": "Invalid amount format",
		})
		return
	}
	amountFromVNPay = amountFromVNPay / 100 // VNPay gửi amount * 100

	// So sánh với số tiền trong transaction
	var transactionAmount float64
	if transaction.SoTien.Valid {
		floatVal, err := transaction.SoTien.Float64Value()
		if err == nil && floatVal.Valid {
			transactionAmount = floatVal.Float64
		} else {
			intVal, err := transaction.SoTien.Int64Value()
			if err == nil && intVal.Valid {
				transactionAmount = float64(intVal.Int64)
			}
		}
	}

	totalAmount := int64(transactionAmount) == amountFromVNPay
	if !totalAmount {
		// invalid amount
		c.JSON(http.StatusOK, gin.H{
			"RspCode": "04",
			"Message": "invalid amount",
		})
		return
	}

	// Xử lý thanh toán thành công
	if responseCode == "00" {
		fmt.Printf("IPN: Payment SUCCESS - ResponseCode: %s, TransactionNo: %s\n", responseCode, transactionNo)
		if transaction.DatChoID == nil {
			fmt.Printf("IPN ERROR: Transaction has no booking ID\n")
			c.JSON(http.StatusOK, gin.H{
				"RspCode": "05",
				"Message": "Invalid transaction",
			})
			return
		}
		bookingID := int(*transaction.DatChoID)
		fmt.Printf("IPN: Processing booking ID: %d\n", bookingID)

		// Kiểm tra booking tồn tại và ở trạng thái hợp lệ
		booking, err := s.z.GetBookingById(c.Request.Context(), int32(bookingID))
		if err != nil {
			fmt.Printf("IPN ERROR: Booking not found - ID: %d, Error: %v\n", bookingID, err)
			c.JSON(http.StatusOK, gin.H{
				"RspCode": "06",
				"Message": "Booking not found",
			})
			return
		}

		// Kiểm tra booking status hợp lệ
		if !booking.TrangThai.Valid || (booking.TrangThai.TrangThaiDatCho != "cho_xac_nhan" && booking.TrangThai.TrangThaiDatCho != "da_xac_nhan") {
			fmt.Printf("IPN ERROR: Invalid booking status - ID: %d, Status: %v\n", bookingID, booking.TrangThai)
			c.JSON(http.StatusOK, gin.H{
				"RspCode": "07",
				"Message": "Invalid booking status",
			})
			return
		}

		// Cập nhật transaction
		trangThai := db.NullTrangThaiThanhToan{
			TrangThaiThanhToan: "thanh_cong",
			Valid:              true,
		}
		_, err = s.z.UpdateTransactionStatus(c.Request.Context(), db.UpdateTransactionStatusParams{
			ID:          transaction.ID,
			TrangThai:   trangThai,
			MaThamChieu: &transactionNo,
		})
		if err != nil {
			fmt.Printf("IPN ERROR: Failed to update transaction status - ID: %d, Error: %v\n", transaction.ID, err)
			c.JSON(http.StatusOK, gin.H{
				"RspCode": "08",
				"Message": "Failed to update transaction",
			})
			return
		}
		fmt.Printf("IPN: Transaction status updated successfully - ID: %d\n", transaction.ID)

		// Cập nhật booking status
		phuongThuc := "vnpay"
		updatedBooking, err := s.z.UpdateBookingPaymentStatus(c.Request.Context(), db.UpdateBookingPaymentStatusParams{
			ID:                  int32(bookingID),
			PhuongThucThanhToan: &phuongThuc,
		})
		if err != nil {
			fmt.Printf("IPN ERROR: Failed to update booking payment status - ID: %d, Error: %v\n", bookingID, err)
			c.JSON(http.StatusOK, gin.H{
				"RspCode": "09",
				"Message": "Failed to update booking",
			})
			return
		}
		fmt.Printf("IPN: Booking payment status updated successfully - ID: %d\n", bookingID)

		// Kiểm tra xem booking đã được cập nhật thành công chưa
		if !updatedBooking.TrangThai.Valid || updatedBooking.TrangThai.TrangThaiDatCho != "da_thanh_toan" {
			fmt.Printf("IPN WARNING: Booking status not updated correctly - ID: %d, Status: %v\n", bookingID, updatedBooking.TrangThai)
		}

		// Payment Success
		fmt.Printf("IPN: Successfully processed payment - Transaction: %s, Booking: %d\n", txnRef, bookingID)
		c.JSON(http.StatusOK, gin.H{
			"RspCode": "00",
			"Message": "Confirm Success",
		})
		return
	}

	// Payment Error (responseCode != "00")
	fmt.Printf("IPN: Payment FAILED - ResponseCode: %s\n", responseCode)
	trangThai := db.NullTrangThaiThanhToan{
		TrangThaiThanhToan: "that_bai",
		Valid:              true,
	}
	_, err = s.z.UpdateTransactionStatus(c.Request.Context(), db.UpdateTransactionStatusParams{
		ID:        transaction.ID,
		TrangThai: trangThai,
	})
	if err != nil {
		fmt.Printf("IPN ERROR: Failed to update transaction status to failed - ID: %d, Error: %v\n", transaction.ID, err)
	} else {
		fmt.Printf("IPN: Transaction status updated to failed - ID: %d\n", transaction.ID)
	}

	c.JSON(http.StatusOK, gin.H{
		"RspCode": "10",
		"Message": "Payment Error",
	})
}

// Helper functions

type CreateVNPayPaymentRequest struct {
	BookingID int32  `json:"booking_id" binding:"required"`
	ReturnURL string `json:"return_url"`
}

func (s *Server) createVNPayPaymentURL(booking db.GetBookingByIdRow, transactionCode, returnURL, clientIP string) (string, error) {
	vnpayConfig := s.config.VNPayConfig

	// Validate VNPay config trước khi tạo URL
	if vnpayConfig.TMNCode == "" {
		return "", fmt.Errorf("VNPAY_TMN_CODE is not configured. Please set VNPAY_TMN_CODE in environment variables")
	}
	if vnpayConfig.HashSecret == "" {
		return "", fmt.Errorf("VNPAY_HASH_SECRET is not configured. Please set VNPAY_HASH_SECRET in environment variables")
	}
	if vnpayConfig.PaymentURL == "" {
		return "", fmt.Errorf("VNPAY_PAYMENT_URL is not configured. Please set VNPAY_PAYMENT_URL in environment variables (e.g., https://sandbox.vnpayment.vn/paymentv2/vpcpay.html)")
	}

	// 1. CỐ ĐỊNH MÚI GIỜ VIỆT NAM (Sửa lỗi "Quá thời gian")
	// Sử dụng FixedZone để đảm bảo không phụ thuộc vào OS/Docker timezone
	location := time.FixedZone("ICT", 7*3600)
	now := time.Now().In(location)

	// 2. Tính toán số tiền (VND * 100)
	var amount int64
	if booking.TongTien.Valid {
		floatVal, _ := booking.TongTien.Float64Value()
		amount = int64(floatVal.Float64)
	}

	// 3. Khởi tạo danh sách tham số
	params := url.Values{}
	params.Add("vnp_Version", "2.1.0")
	params.Add("vnp_Command", "pay")
	params.Add("vnp_TmnCode", vnpayConfig.TMNCode)
	params.Add("vnp_Amount", strconv.FormatInt(amount*100, 10))
	params.Add("vnp_CurrCode", "VND")
	params.Add("vnp_TxnRef", transactionCode)
	params.Add("vnp_OrderInfo", fmt.Sprintf("Thanh toan don dat cho #%d", booking.ID))
	params.Add("vnp_OrderType", "other")
	params.Add("vnp_Locale", "vn")
	// Sử dụng returnURL từ request hoặc config
	finalReturnURL := returnURL
	if finalReturnURL == "" || finalReturnURL == "string" {
		finalReturnURL = vnpayConfig.ReturnURL
	}

	// Validate Return URL
	if finalReturnURL == "" || finalReturnURL == "string" {
		return "", fmt.Errorf("Return URL is not configured. Please set VNPAY_RETURN_URL in .env or provide return_url in request")
	}

	// Log để debug
	fmt.Printf("=== Creating VNPay Payment URL ===\n")
	fmt.Printf("TMN Code: %s\n", vnpayConfig.TMNCode)
	fmt.Printf("Payment URL (VNPay Gateway): %s\n", vnpayConfig.PaymentURL)
	fmt.Printf("Return URL (from request): %s\n", returnURL)
	fmt.Printf("Return URL (final): %s\n", finalReturnURL)
	fmt.Printf("IPN URL (from config): %s\n", vnpayConfig.IPNURL)
	fmt.Printf("Transaction Code: %s\n", transactionCode)

	// Cảnh báo nếu TMN Code rỗng
	if vnpayConfig.TMNCode == "" {
		fmt.Printf("❌ ERROR: TMN Code is empty! Payment will fail.\n")
	}

	// Cảnh báo nếu Payment URL rỗng hoặc không phải VNPay gateway
	if vnpayConfig.PaymentURL == "" {
		fmt.Printf("❌ ERROR: Payment URL is empty! Cannot create payment URL.\n")
	} else if !strings.Contains(vnpayConfig.PaymentURL, "vnpayment.vn") {
		fmt.Printf("⚠️  WARNING: Payment URL does not look like VNPay gateway: %s\n", vnpayConfig.PaymentURL)
	}

	// Cảnh báo nếu IPN URL là localhost
	if strings.Contains(vnpayConfig.IPNURL, "localhost") || strings.Contains(vnpayConfig.IPNURL, "127.0.0.1") {
		fmt.Printf("⚠️  WARNING: IPN URL is localhost! VNPay cannot call localhost. Use ngrok or deploy to production.\n")
	}

	// Cảnh báo nếu Return URL là localhost
	if strings.Contains(finalReturnURL, "localhost") || strings.Contains(finalReturnURL, "127.0.0.1") {
		fmt.Printf("⚠️  WARNING: Return URL is localhost! VNPay may not redirect correctly. Use ngrok or deploy to production.\n")
	}

	fmt.Printf("================================\n")

	params.Add("vnp_ReturnUrl", finalReturnURL)
	params.Add("vnp_IpAddr", clientIP)
	params.Add("vnp_CreateDate", now.Format("20060102150405"))
	params.Add("vnp_ExpireDate", now.Add(15*time.Minute).Format("20060102150405"))

	// 4. SẮP XẾP KEY THEO ALPHABET (Bắt buộc để đúng chữ ký)
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 5. TẠO CHUỖI QUERY (Dùng dấu + đồng nhất)
	// params.Encode() của Go sẽ tự sắp xếp và dùng dấu + cho khoảng trắng
	// Nhưng ta build thủ công theo thứ tự keys đã sort ở trên cho chắc chắn
	var queryBuilder strings.Builder
	for i, k := range keys {
		if i > 0 {
			queryBuilder.WriteString("&")
		}
		queryBuilder.WriteString(k)
		queryBuilder.WriteString("=")
		queryBuilder.WriteString(url.QueryEscape(params.Get(k)))
	}
	queryString := queryBuilder.String()

	// 6. TÍNH CHỮ KÝ (HMAC-SHA512)
	mac := hmac.New(sha512.New, []byte(vnpayConfig.HashSecret))
	mac.Write([]byte(queryString))
	secureHash := hex.EncodeToString(mac.Sum(nil))

	// 7. URL CUỐI CÙNG
	finalURL := fmt.Sprintf("%s?%s&vnp_SecureHash=%s", vnpayConfig.PaymentURL, queryString, secureHash)

	// Validate final URL
	if !strings.HasPrefix(finalURL, "https://") && !strings.HasPrefix(finalURL, "http://") {
		return "", fmt.Errorf("Invalid payment URL format: %s", finalURL)
	}

	// Log final URL (chỉ log một phần để không expose sensitive data)
	fmt.Printf("Final Payment URL (first 100 chars): %s...\n", finalURL[:min(100, len(finalURL))])
	fmt.Printf("================================\n")

	return finalURL, nil
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// verifyVNPaySignature xác thực signature từ VNPay callback
func (s *Server) verifyVNPaySignature(params url.Values) bool {
	receivedHash := params.Get("vnp_SecureHash")
	if receivedHash == "" {
		return false
	}

	// Tạo bản sao của params và loại bỏ SecureHash
	verifyParams := make(url.Values)
	for k, v := range params {
		if k != "vnp_SecureHash" && k != "vnp_SecureHashType" {
			verifyParams[k] = v
		}
	}

	// Sắp xếp keys theo alphabet (giống như createVNPayPaymentURL)
	keys := make([]string, 0, len(verifyParams))
	for k := range verifyParams {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Tạo query string thủ công (giống như createVNPayPaymentURL)
	var queryBuilder strings.Builder
	for i, k := range keys {
		if i > 0 {
			queryBuilder.WriteString("&")
		}
		queryBuilder.WriteString(k)
		queryBuilder.WriteString("=")
		queryBuilder.WriteString(url.QueryEscape(verifyParams.Get(k)))
	}
	queryString := queryBuilder.String()

	// Tính signature bằng HMAC SHA512
	mac := hmac.New(sha512.New, []byte(s.config.VNPayConfig.HashSecret))
	mac.Write([]byte(queryString))
	calculatedHash := hex.EncodeToString(mac.Sum(nil))

	return receivedHash == calculatedHash
}
