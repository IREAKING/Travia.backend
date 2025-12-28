# Hướng dẫn Test VNPay trên Localhost với Ngrok

## Vấn đề
VNPay không thể gọi trực tiếp về `localhost` hoặc `127.0.0.1` vì đó là địa chỉ nội bộ. Để test VNPay trên localhost, bạn cần sử dụng **ngrok** để tạo tunnel từ localhost ra internet public.

---

## Giải pháp: Sử dụng Ngrok

### Bước 1: Cài đặt Ngrok

**macOS:**
```bash
brew install ngrok/ngrok/ngrok
```

**Hoặc download từ:**
- https://ngrok.com/download
- Hoặc: `brew install ngrok`

**Windows/Linux:**
- Download từ https://ngrok.com/download
- Giải nén và thêm vào PATH

### Bước 2: Đăng ký tài khoản Ngrok (Miễn phí)

1. Truy cập: https://dashboard.ngrok.com/signup
2. Đăng ký tài khoản miễn phí
3. Lấy **Authtoken** từ dashboard: https://dashboard.ngrok.com/get-started/your-authtoken

### Bước 3: Cấu hình Ngrok

```bash
ngrok config add-authtoken YOUR_AUTHTOKEN
```

### Bước 4: Khởi động Ngrok Tunnel

Bạn cần tạo **2 tunnels**:
1. **Frontend tunnel** (port 5173) - cho Return URL
2. **Backend tunnel** (port 3000) - cho IPN URL

#### Cách 1: Sử dụng 2 terminal riêng

**Terminal 1 - Frontend tunnel:**
```bash
ngrok http 5173
```

**Terminal 2 - Backend tunnel:**
```bash
ngrok http 3000
```

#### Cách 2: Sử dụng ngrok config file (Khuyến nghị)

Tạo file `ngrok.yml` trong thư mục home:

```yaml
version: "2"
authtoken: YOUR_AUTHTOKEN

tunnels:
  frontend:
    addr: 5173
    proto: http
  backend:
    addr: 3000
    proto: http
```

Sau đó chạy:
```bash
ngrok start --all
```

Hoặc chạy từng tunnel:
```bash
ngrok start frontend
ngrok start backend
```

### Bước 5: Lấy Public URLs từ Ngrok

Sau khi khởi động ngrok, bạn sẽ thấy URLs như:

```
Forwarding  https://abc123.ngrok-free.app -> http://localhost:5173
Forwarding  https://xyz789.ngrok-free.app -> http://localhost:3000
```

**Lưu ý:** 
- URLs này sẽ thay đổi mỗi lần khởi động ngrok (trừ khi dùng plan trả phí)
- Với plan miễn phí, bạn có thể đặt tên cố định bằng cách đăng ký domain tùy chỉnh

### Bước 6: Cấu hình Environment Variables

Cập nhật file `.env` của backend:

```env
# VNPay Configuration
VNPAY_TMN_CODE=your_tmn_code
VNPAY_HASH_SECRET=your_hash_secret
VNPAY_PAYMENT_URL=https://sandbox.vnpayment.vn/paymentv2/vpcpay.html

# Sử dụng ngrok URLs
VNPAY_RETURN_URL=https://abc123.ngrok-free.app/payment/vnpay/return
VNPAY_IPN_URL=https://xyz789.ngrok-free.app/api/payment/vnpay/ipn
```

**Lưu ý:** 
- Thay `abc123` và `xyz789` bằng URLs thực tế từ ngrok của bạn
- Mỗi lần khởi động lại ngrok, bạn cần cập nhật lại URLs trong `.env`

### Bước 7: Khởi động lại Backend

```bash
# Restart backend để load environment variables mới
# Ctrl+C để dừng, sau đó chạy lại
go run main.go
```

---

## Script Helper Tự Động (Tùy chọn)

Tạo script để tự động lấy ngrok URLs và cập nhật `.env`:

### Script cho macOS/Linux (`setup-ngrok.sh`):

```bash
#!/bin/bash

# Lấy frontend URL từ ngrok
FRONTEND_URL=$(curl -s http://localhost:4040/api/tunnels | jq -r '.tunnels[] | select(.config.addr == "localhost:5173") | .public_url')

# Lấy backend URL từ ngrok
BACKEND_URL=$(curl -s http://localhost:4040/api/tunnels | jq -r '.tunnels[] | select(.config.addr == "localhost:3000") | .public_url')

if [ -z "$FRONTEND_URL" ] || [ -z "$BACKEND_URL" ]; then
    echo "❌ Không tìm thấy ngrok tunnels. Hãy đảm bảo ngrok đang chạy."
    exit 1
fi

echo "✅ Frontend URL: $FRONTEND_URL"
echo "✅ Backend URL: $BACKEND_URL"

# Cập nhật .env
sed -i.bak "s|VNPAY_RETURN_URL=.*|VNPAY_RETURN_URL=$FRONTEND_URL/payment/vnpay/return|" .env
sed -i.bak "s|VNPAY_IPN_URL=.*|VNPAY_IPN_URL=$BACKEND_URL/api/payment/vnpay/ipn|" .env

echo "✅ Đã cập nhật .env file"
```

**Sử dụng:**
```bash
chmod +x setup-ngrok.sh
./setup-ngrok.sh
```

---

## Workflow Test Hoàn Chỉnh

1. **Khởi động ngrok:**
   ```bash
   ngrok start --all
   # Hoặc chạy riêng:
   # Terminal 1: ngrok http 5173
   # Terminal 2: ngrok http 3000
   ```

2. **Cập nhật `.env` với ngrok URLs**

3. **Khởi động Frontend:**
   ```bash
   cd Travia.frontend
   npm run dev
   ```

4. **Khởi động Backend:**
   ```bash
   cd Travia.backend
   go run main.go
   ```

5. **Test thanh toán:**
   - Tạo booking
   - Click thanh toán VNPay
   - VNPay sẽ redirect về ngrok URL
   - Ngrok sẽ forward về localhost của bạn

---

## Troubleshooting

### Lỗi: "ngrok: command not found"
- Đảm bảo đã cài đặt ngrok và thêm vào PATH
- macOS: `brew install ngrok`
- Kiểm tra: `ngrok version`

### Lỗi: "tunnel session failed"
- Kiểm tra authtoken có đúng không: `ngrok config check`
- Kiểm tra port có đang được sử dụng không

### Lỗi: "VNPay không redirect về"
- Kiểm tra ngrok có đang chạy không: http://localhost:4040
- Kiểm tra URLs trong `.env` có đúng không
- Kiểm tra frontend có đang chạy trên port 5173 không

### Lỗi: "IPN không được gọi"
- Kiểm tra backend tunnel có đang chạy không
- Kiểm tra `VNPAY_IPN_URL` trong `.env`
- Kiểm tra VNPay dashboard có cấu hình IPN URL không (nếu cần)

### URLs thay đổi mỗi lần khởi động
- Đây là hành vi bình thường với plan miễn phí
- Giải pháp: Sử dụng ngrok config file và script tự động cập nhật `.env`
- Hoặc nâng cấp lên plan trả phí để có fixed domain

---

## Lưu Ý Quan Trọng

1. **Security:** Ngrok URLs là public, ai cũng có thể truy cập. Chỉ dùng cho development/test.

2. **Rate Limits:** Plan miễn phí có giới hạn số request. Nếu vượt quá, ngrok sẽ hiển thị trang "Visit Site" trước khi redirect.

3. **HTTPS:** Ngrok tự động cung cấp HTTPS, phù hợp với VNPay yêu cầu.

4. **Production:** Không dùng ngrok trong production. Deploy lên server thật với domain riêng.

---

## Tài Liệu Tham Khảo

- Ngrok Documentation: https://ngrok.com/docs
- Ngrok Dashboard: https://dashboard.ngrok.com
- VNPay Sandbox: https://sandbox.vnpayment.vn/

