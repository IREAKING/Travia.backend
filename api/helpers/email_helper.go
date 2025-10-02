package helpers

import (
	"fmt"
	"log"
	"math/rand"
	"net/smtp"
	"strings"

	"travia.backend/config"
)

// GenerateVerificationCode generates a random 6-digit verification code
func GenerateVerificationCode() string {
	// Generate 6 random digits
	code := make([]byte, 6)
	for i := range code {
		code[i] = byte('0' + rand.Intn(10))
	}
	return string(code)
}

// SendVerificationEmail sends verification code to user's email
func SendVerificationEmail(toEmail, verificationCode string, e *config.EmailConfig) error {
	subject := "Travia - Xác thực Email"

	// HTML email template
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Xác thực Email - Travia</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .verification-code { background: #fff; border: 2px dashed #667eea; padding: 20px; text-align: center; margin: 20px 0; border-radius: 8px; }
        .code { font-size: 32px; font-weight: bold; color: #667eea; letter-spacing: 5px; }
        .footer { text-align: center; margin-top: 30px; color: #666; font-size: 14px; }
        .warning { background: #fff3cd; border: 1px solid #ffeaa7; padding: 15px; border-radius: 5px; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎉 Chào mừng đến với Travia!</h1>
            <p>Xác thực email của bạn để hoàn tất đăng ký</p>
        </div>
        
        <div class="content">
            <h2>Xin chào!</h2>
            <p>Cảm ơn bạn đã đăng ký tài khoản tại <strong>Travia</strong>. Để hoàn tất quá trình đăng ký, vui lòng nhập mã xác thực 6 số dưới đây:</p>
            
            <div class="verification-code">
                <div class="code">%s</div>
                <p><strong>Mã xác thực của bạn</strong></p>
            </div>
            
            <div class="warning">
                <strong>⚠️ Lưu ý quan trọng:</strong>
                <ul>
                    <li>Mã này có hiệu lực trong 10 phút</li>
                    <li>Không chia sẻ mã này với bất kỳ ai</li>
                    <li>Nếu bạn không thực hiện yêu cầu này, vui lòng bỏ qua email này</li>
                </ul>
            </div>
            
            <p>Nếu bạn gặp vấn đề, vui lòng liên hệ với chúng tôi tại <a href="mailto:support@travia.com">support@travia.com</a></p>
        </div>
        
        <div class="footer">
            <p>© 2024 Travia. Tất cả quyền được bảo lưu.</p>
            <p>Email này được gửi tự động, vui lòng không trả lời.</p>
        </div>
    </div>
</body>
</html>`, verificationCode)

	// Plain text version
	textBody := fmt.Sprintf(`
Chào mừng đến với Travia!

Cảm ơn bạn đã đăng ký tài khoản. Để hoàn tất quá trình đăng ký, vui lòng nhập mã xác thực sau:

Mã xác thực: %s

Mã này có hiệu lực trong 10 phút.

Lưu ý:
- Không chia sẻ mã này với bất kỳ ai
- Nếu bạn không thực hiện yêu cầu này, vui lòng bỏ qua email này

Nếu bạn gặp vấn đề, vui lòng liên hệ: support@travia.com

© 2024 Travia. Tất cả quyền được bảo lưu.
`, verificationCode)

	return sendEmail(toEmail, subject, textBody, htmlBody, e)
}

// SendWelcomeEmail sends welcome email after successful verification
func SendWelcomeEmail(toEmail, firstName string, e *config.EmailConfig) error {
	subject := "Travia - Chào mừng bạn!"

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Chào mừng - Travia</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .cta-button { display: inline-block; background: #667eea; color: white; padding: 15px 30px; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { text-align: center; margin-top: 30px; color: #666; font-size: 14px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎉 Chào mừng %s!</h1>
            <p>Tài khoản của bạn đã được xác thực thành công</p>
        </div>
        
        <div class="content">
            <h2>Xin chào %s!</h2>
            <p>Chúc mừng! Tài khoản Travia của bạn đã được xác thực thành công. Bây giờ bạn có thể:</p>
            
            <ul>
                <li>🔍 Khám phá các điểm đến tuyệt vời</li>
                <li>📅 Đặt tour du lịch</li>
                <li>⭐ Đánh giá và chia sẻ trải nghiệm</li>
                <li>💳 Quản lý thông tin cá nhân</li>
            </ul>
            
            <p style="text-align: center;">
                <a href="https://travia.com" class="cta-button">Bắt đầu khám phá</a>
            </p>
            
            <p>Nếu bạn có câu hỏi, đừng ngần ngại liên hệ với chúng tôi tại <a href="mailto:support@travia.com">support@travia.com</a></p>
        </div>
        
        <div class="footer">
            <p>© 2024 Travia. Tất cả quyền được bảo lưu.</p>
        </div>
    </div>
</body>
</html>`, firstName, firstName)

	textBody := fmt.Sprintf(`
Chào mừng %s!

Chúc mừng! Tài khoản Travia của bạn đã được xác thực thành công.

Bây giờ bạn có thể:
- Khám phá các điểm đến tuyệt vời
- Đặt tour du lịch
- Đánh giá và chia sẻ trải nghiệm
- Quản lý thông tin cá nhân

Truy cập: https://travia.com

Nếu bạn có câu hỏi, vui lòng liên hệ: support@travia.com

© 2024 Travia. Tất cả quyền được bảo lưu.
`, firstName)

	return sendEmail(toEmail, subject, textBody, htmlBody, e)
}

// sendEmail sends an email using SMTP
func sendEmail(toEmail, subject, textBody, htmlBody string, e *config.EmailConfig) error {
	// SMTP configuration
	auth := smtp.PlainAuth("", e.SMTPUsername, e.SMTPPassword, e.SMTPHost)

	// Email headers
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", e.FromName, e.FromEmail)
	headers["To"] = toEmail
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	// Build email message
	var message strings.Builder
	for key, value := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}
	message.WriteString("\r\n")
	message.WriteString(htmlBody)

	// Send email
	err := smtp.SendMail(
		e.SMTPHost+":"+e.SMTPPort,
		auth,
		e.FromEmail,
		[]string{toEmail},
		[]byte(message.String()),
	)

	if err != nil {
		log.Printf("Failed to send email to %s: %v", toEmail, err)
		return err
	}

	log.Printf("Email sent successfully to %s", toEmail)
	return nil
}

// MockEmailService for development/testing
type MockEmailService struct{}

func NewMockEmailService() *MockEmailService {
	return &MockEmailService{}
}

func (m *MockEmailService) SendVerificationEmail(toEmail, verificationCode string) error {
	log.Printf("MOCK: Verification email sent to %s with code: %s", toEmail, verificationCode)
	return nil
}

func (m *MockEmailService) SendWelcomeEmail(toEmail, firstName string) error {
	log.Printf("MOCK: Welcome email sent to %s for %s", toEmail, firstName)
	return nil
}
