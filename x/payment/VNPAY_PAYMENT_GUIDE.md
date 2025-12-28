# Hướng dẫn sử dụng VNPay Payment API

## Tổng quan

Hệ thống hỗ trợ thanh toán qua VNPay với 3 endpoint chính:
1. **CreateVNPayPaymentURL** - Tạo URL thanh toán
2. **VNPayCallback** - Xử lý Return URL (redirect từ VNPay)
3. **VNPayIPN** - Xử lý IPN callback (server-to-server)

---

## 1. Cấu hình môi trường

Thêm các biến môi trường sau vào file `.env`:

```env
# VNPay Configuration
VNPAY_TMN_CODE=your_tmn_code                    # Mã website từ VNPay
VNPAY_HASH_SECRET=your_hash_secret              # Secret key từ VNPay
VNPAY_PAYMENT_URL=https://sandbox.vnpayment.vn/paymentv2/vpcpay.html  # URL thanh toán (sandbox hoặc production)
VNPAY_RETURN_URL=http://localhost:5173/payment/vnpay/return  # URL redirect sau khi thanh toán
VNPAY_IPN_URL=http://localhost:3000/api/payment/vnpay/ipn   # URL nhận IPN callback
```

**Lưu ý:**
- Sandbox: `https://sandbox.vnpayment.vn/paymentv2/vpcpay.html`
- Production: `https://www.vnpayment.vn/paymentv2/vpcpay.html`
- `VNPAY_IPN_URL` phải là URL public (không thể là localhost trong production)

---

## 2. Endpoint 1: Tạo URL thanh toán

### `POST /api/payment/vnpay/create`

**Mục đích:** Tạo URL thanh toán VNPay cho một booking

**Authentication:** Required (JWT token)

**Request Body:**
```json
{
  "booking_id": 123,
  "return_url": "http://localhost:5173/payment/vnpay/return"  // Optional
}
```

**Response Success (200):**
```json
{
  "payment_url": "https://sandbox.vnpayment.vn/paymentv2/vpcpay.html?vnp_Amount=...",
  "transaction_code": "TRAVIA1231234567890",
  "booking_id": 123
}
```

**Response Error:**
```json
{
  "error": "Booking not found"
}
```

**Ví dụ sử dụng (JavaScript/Frontend):**
```javascript
// 1. Gọi API tạo payment URL
const response = await fetch('/api/payment/vnpay/create', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    // Cookie sẽ tự động gửi kèm (httpOnly cookie)
  },
  credentials: 'include',
  body: JSON.stringify({
    booking_id: 123,
    return_url: 'http://localhost:5173/payment/vnpay/return'
  })
});

const data = await response.json();

if (response.ok) {
  // 2. Redirect user đến VNPay
  window.location.href = data.payment_url;
} else {
  console.error('Error:', data.error);
}
```

**Ví dụ sử dụng (cURL):**
```bash
curl -X POST http://localhost:3000/api/payment/vnpay/create \
  -H "Content-Type: application/json" \
  -H "Cookie: access_token=your_token" \
  -d '{
    "booking_id": 123,
    "return_url": "http://localhost:5173/payment/vnpay/return"
  }'
```

**Flow:**
1. User chọn booking cần thanh toán
2. Frontend gọi API này với `booking_id`
3. Backend tạo transaction và payment URL
4. Frontend redirect user đến `payment_url`
5. User thanh toán trên VNPay

---

## 3. Endpoint 2: Return URL Callback

### `GET /api/payment/vnpay/return`

**Mục đích:** Xử lý callback khi user quay lại từ VNPay sau khi thanh toán

**Authentication:** Không cần (public endpoint)

**Query Parameters (từ VNPay):**
```
vnp_Amount=10000000
vnp_BankCode=NCB
vnp_BankTranNo=VNP12345678
vnp_CardType=ATM
vnp_OrderInfo=Thanh toan don dat cho #123
vnp_PayDate=20240101120000
vnp_ResponseCode=00
vnp_TmnCode=YOUR_TMN_CODE
vnp_TransactionNo=12345678
vnp_TransactionStatus=00
vnp_TxnRef=TRAVIA1231234567890
vnp_SecureHash=abc123...
```

**Response:**
- **Success:** Redirect về frontend với query params
  ```
  http://localhost:5173/payment/vnpay/return?status=success&booking_id=123&transaction_code=TRAVIA1231234567890
  ```

- **Failed:** Redirect về frontend với error
  ```
  http://localhost:5173/payment/vnpay/return?status=failed&booking_id=123&transaction_code=TRAVIA1231234567890&error_code=07
  ```

**Ví dụ xử lý trên Frontend:**
```javascript
// Trong component PaymentReturnPage.tsx
import { useSearchParams, useNavigate } from 'react-router-dom';

function PaymentReturnPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  
  const status = searchParams.get('status');
  const bookingId = searchParams.get('booking_id');
  const transactionCode = searchParams.get('transaction_code');
  const errorCode = searchParams.get('error_code');

  useEffect(() => {
    if (status === 'success') {
      // Hiển thị thông báo thành công
      showSuccessToast('Thanh toán thành công!');
      // Redirect về trang booking
      setTimeout(() => {
        navigate(`/booking/${bookingId}`);
      }, 3000);
    } else if (status === 'failed') {
      // Hiển thị thông báo lỗi
      const errorMessages = {
        '07': 'Trừ tiền thành công nhưng giao dịch bị nghi ngờ',
        '09': 'Thẻ/Tài khoản chưa đăng ký dịch vụ InternetBanking',
        '10': 'Xác thực thông tin thẻ/tài khoản không đúng quá 3 lần',
        '11': 'Đã hết hạn chờ thanh toán',
        '12': 'Thẻ/Tài khoản bị khóa',
        '51': 'Tài khoản không đủ số dư để thực hiện giao dịch',
        '65': 'Tài khoản đã vượt quá hạn mức giao dịch trong ngày',
        '75': 'Ngân hàng thanh toán đang bảo trì',
        '99': 'Lỗi không xác định'
      };
      
      const errorMsg = errorMessages[errorCode] || 'Thanh toán thất bại';
      showErrorToast(errorMsg);
      
      // Redirect về trang booking để thử lại
      setTimeout(() => {
        navigate(`/booking/${bookingId}`);
      }, 3000);
    }
  }, [status, bookingId, errorCode, navigate]);

  return (
    <div>
      {status === 'success' ? (
        <div>
          <h1>Thanh toán thành công!</h1>
          <p>Mã giao dịch: {transactionCode}</p>
        </div>
      ) : (
        <div>
          <h1>Thanh toán thất bại</h1>
          <p>Mã lỗi: {errorCode}</p>
        </div>
      )}
    </div>
  );
}
```

**Flow:**
1. User thanh toán trên VNPay
2. VNPay redirect về `ReturnURL` với query params
3. Backend verify signature và cập nhật transaction
4. Backend redirect về frontend với status
5. Frontend hiển thị kết quả cho user

---

## 4. Endpoint 3: IPN Callback

### `POST /api/payment/vnpay/ipn`

**Mục đích:** Xử lý IPN (Instant Payment Notification) từ VNPay server

**Authentication:** Không cần (public endpoint, chỉ VNPay server gọi)

**Request Body (Form Data từ VNPay):**
```
vnp_Amount=10000000
vnp_BankCode=NCB
vnp_BankTranNo=VNP12345678
vnp_CardType=ATM
vnp_OrderInfo=Thanh toan don dat cho #123
vnp_PayDate=20240101120000
vnp_ResponseCode=00
vnp_TmnCode=YOUR_TMN_CODE
vnp_TransactionNo=12345678
vnp_TransactionStatus=00
vnp_TxnRef=TRAVIA1231234567890
vnp_SecureHash=abc123...
```

**Response:**
- **Success:** `200 OK` với message
  ```
  Transaction processed successfully
  ```

- **Already Processed:** `200 OK`
  ```
  Transaction already processed
  ```

- **Invalid Signature:** `400 Bad Request`
  ```
  Invalid signature
  ```

**Lưu ý quan trọng:**
- Endpoint này được VNPay server gọi tự động (server-to-server)
- Phải verify signature để đảm bảo request đến từ VNPay
- Phải xử lý idempotent (không xử lý trùng transaction)
- Response phải trả về nhanh (< 5 giây)

**Flow:**
1. User thanh toán trên VNPay
2. VNPay xử lý thanh toán
3. VNPay gọi IPN URL (server-to-server)
4. Backend verify signature và cập nhật transaction
5. Backend trả về status code cho VNPay

---

## 5. Response Codes từ VNPay

### Response Code (`vnp_ResponseCode`):
- `00`: Giao dịch thành công
- `07`: Trừ tiền thành công nhưng giao dịch bị nghi ngờ
- `09`: Thẻ/Tài khoản chưa đăng ký dịch vụ InternetBanking
- `10`: Xác thực thông tin thẻ/tài khoản không đúng quá 3 lần
- `11`: Đã hết hạn chờ thanh toán
- `12`: Thẻ/Tài khoản bị khóa
- `51`: Tài khoản không đủ số dư để thực hiện giao dịch
- `65`: Tài khoản đã vượt quá hạn mức giao dịch trong ngày
- `75`: Ngân hàng thanh toán đang bảo trì
- `99`: Lỗi không xác định

### Transaction Status (`vnp_TransactionStatus`):
- `00`: Giao dịch thành công
- Khác `00`: Giao dịch thất bại

---

## 6. Luồng thanh toán hoàn chỉnh

```
┌─────────┐         ┌──────────┐         ┌─────────┐         ┌──────────┐
│ Frontend│         │ Backend  │         │  VNPay  │         │  VNPay   │
│         │         │          │         │  (User) │         │  Server  │
└────┬────┘         └────┬─────┘         └────┬────┘         └────┬─────┘
     │                   │                    │                    │
     │ 1. POST /create   │                    │                    │
     │──────────────────>│                    │                    │
     │                   │                    │                    │
     │ 2. payment_url    │                    │                    │
     │<──────────────────│                    │                    │
     │                   │                    │                    │
     │ 3. Redirect       │                    │                    │
     │───────────────────────────────────────>│                    │
     │                   │                    │                    │
     │                   │                    │ 4. User thanh toán │
     │                   │                    │<───────────────────│
     │                   │                    │                    │
     │                   │                    │ 5. Return URL      │
     │                   │<───────────────────│                    │
     │                   │                    │                    │
     │                   │ 6. IPN Callback    │                    │
     │                   │<───────────────────────────────────────│
     │                   │                    │                    │
     │                   │ 7. Update DB       │                    │
     │                   │                    │                    │
     │ 8. Redirect       │                    │                    │
     │<──────────────────│                    │                    │
     │                   │                    │                    │
```

---

## 7. Testing với Swagger UI

### 7.0. Flow gọi API - Endpoint nào gọi đầu tiên?

**Thứ tự các bước:**

1. **Bước 1: Đăng nhập** (Bắt buộc)
   - Endpoint: `POST /api/auth/login/user`
   - Mục đích: Lấy authentication token/cookie
   - **Đây là endpoint đầu tiên bạn cần gọi**

2. **Bước 2: Tạo Payment URL** (Sau khi đã login)
   - Endpoint: `POST /api/payment/vnpay/create`
   - Mục đích: Tạo URL thanh toán VNPay
   - **Đây là endpoint thứ 2 cần gọi**

3. **Bước 3: Redirect đến VNPay** (Tự động)
   - Mở `payment_url` từ response bước 2
   - User thanh toán trên VNPay

4. **Bước 4: Callback tự động** (VNPay gọi)
   - `GET /api/payment/vnpay/return` - VNPay redirect về
   - `POST /api/payment/vnpay/ipn` - VNPay server gọi

**Tóm lại: Endpoint đầu tiên = `POST /api/auth/login/user`**

### 7.1. Truy cập Swagger UI

1. **Khởi động server:**
   ```bash
   cd Travia.backend
   go run main.go
   ```

2. **Mở Swagger UI:**
   - Truy cập: `http://localhost:3000/swagger/index.html`
   - (Thay `3000` bằng port của bạn nếu khác)

### 7.2. Test Endpoint 1: Tạo Payment URL

#### Bước 1: Đăng nhập để lấy token

1. Trong Swagger UI, tìm endpoint `POST /api/auth/login/user`
2. Click "Try it out"
3. Nhập thông tin:
   ```json
   {
     "email": "customer@example.com",
     "password": "your_password"
   }
   ```
4. Click "Execute"
5. Copy `access_token` từ response (hoặc cookie sẽ tự động được set)

#### Bước 2: Authorize trong Swagger

1. Click nút **"Authorize"** ở góc trên bên phải
2. Nếu dùng cookie authentication:
   - Cookie sẽ tự động được gửi kèm khi đã login
   - Không cần nhập gì thêm
3. Nếu dùng Bearer token:
   - Nhập `Bearer {your_access_token}`
4. Click **"Authorize"** và **"Close"**

#### Bước 3: Test Create Payment URL

1. Tìm endpoint `POST /api/payment/vnpay/create` trong section **Payment**
2. Click **"Try it out"**
3. Nhập Request Body:
   ```json
   {
     "booking_id": 1,
     "return_url": "http://localhost:5173/payment/vnpay/return"
   }
   ```

   **Giải thích về `return_url`:**
   
   - **`return_url` là gì?**
     - Đây là URL mà VNPay sẽ redirect user về sau khi họ hoàn tất thanh toán (thành công hoặc thất bại)
     - URL này phải là trang frontend của bạn để hiển thị kết quả thanh toán cho user
   
   - **Nên nhập gì vào `return_url`?**
     
     **Option 1: Để trống (Recommended cho test)**
     ```json
     {
       "booking_id": 1
     }
     ```
     - Nếu không nhập, hệ thống sẽ tự động dùng `VNPAY_RETURN_URL` từ file `.env`
     - Default: `http://localhost:5173/payment/vnpay/return`
     
     **Option 2: Nhập URL frontend của bạn**
     ```json
     {
       "booking_id": 1,
       "return_url": "http://localhost:5173/payment/vnpay/return"
     }
     ```
     - **Local development:** `http://localhost:5173/payment/vnpay/return`
     - **Production:** `https://yourdomain.com/payment/vnpay/return`
     - URL này phải là route hợp lệ trong frontend của bạn
   
   - **Lưu ý quan trọng:**
     - ✅ `return_url` là **optional** (có thể để trống)
     - ✅ URL phải là **URL public** (không thể là localhost trong production)
     - ✅ URL phải có route tương ứng trong frontend để xử lý callback
     - ✅ URL sẽ nhận query parameters từ VNPay: `?status=success&booking_id=1&transaction_code=...`
   
   - **Ví dụ cụ thể:**
     
     **Test local:**
     ```json
     {
       "booking_id": 1,
       "return_url": "http://localhost:5173/payment/vnpay/return"
     }
     ```
     
     **Test với frontend khác port:**
     ```json
     {
       "booking_id": 1,
       "return_url": "http://localhost:3000/payment/vnpay/return"
     }
     ```
     
     **Production:**
     ```json
     {
       "booking_id": 1,
       "return_url": "https://travia.com/payment/vnpay/return"
     }
     ```
   
   **Lưu ý khác:** 
   - `booking_id` phải là ID của booking hợp lệ và thuộc về user đã đăng nhập

4. Click **"Execute"**
5. Kiểm tra Response:
   - **200 OK:** Sẽ trả về `payment_url`, `transaction_code`, `booking_id`
   - **400/401/404:** Kiểm tra error message

6. **Copy `payment_url`** từ response

#### Bước 4: Test Payment trên VNPay

1. Mở `payment_url` trong browser mới (hoặc tab mới)
2. Bạn sẽ thấy trang thanh toán VNPay
3. Sử dụng test card:
   - **Card number:** `9704198526191432198`
   - **Card holder:** Bất kỳ tên nào
   - **Expire date:** Tương lai (ví dụ: 12/25)
   - **CVV:** Bất kỳ 3 số nào
   - **OTP:** `123456`

4. Click "Thanh toán"
5. Sau khi thanh toán, VNPay sẽ redirect về `return_url`

### 7.3. Test Endpoint 2: Return URL Callback

**Lưu ý:** Endpoint này thường được gọi tự động bởi VNPay sau khi thanh toán. Tuy nhiên, bạn có thể test thủ công trong Swagger:

1. Tìm endpoint `GET /api/payment/vnpay/return`
2. Click **"Try it out"**
3. Nhập các query parameters (mô phỏng response từ VNPay):
   ```
   vnp_Amount: 10000000
   vnp_BankCode: NCB
   vnp_BankTranNo: VNP12345678
   vnp_CardType: ATM
   vnp_OrderInfo: Thanh toan don dat cho #1
   vnp_PayDate: 20240101120000
   vnp_ResponseCode: 00
   vnp_TmnCode: YOUR_TMN_CODE
   vnp_TransactionNo: 12345678
   vnp_TransactionStatus: 00
   vnp_TxnRef: TRAVIA11234567890
   vnp_SecureHash: [tính toán signature]
   ```

   **⚠️ Quan trọng:** 
   - `vnp_TxnRef` phải là transaction_code đã được tạo từ endpoint 1
   - `vnp_SecureHash` phải được tính toán đúng (sử dụng `createVNPaySignature`)
   - Nếu signature không đúng, sẽ trả về `400 Bad Request`

4. Click **"Execute"**
5. Response sẽ là redirect (302) về frontend URL

**Cách tính signature (để test):**
```go
// Sử dụng helper function createVNPaySignature trong code
// Hoặc sử dụng công cụ online để tính HMAC SHA512
```

### 7.4. Test Endpoint 3: IPN Callback

**Lưu ý:** Endpoint này được VNPay server gọi tự động. Để test trong Swagger:

1. Tìm endpoint `POST /api/payment/vnpay/ipn`
2. Click **"Try it out"**
3. Nhập form data (mô phỏng request từ VNPay):
   ```
   vnp_Amount: 10000000
   vnp_BankCode: NCB
   vnp_BankTranNo: VNP12345678
   vnp_CardType: ATM
   vnp_OrderInfo: Thanh toan don dat cho #1
   vnp_PayDate: 20240101120000
   vnp_ResponseCode: 00
   vnp_TmnCode: YOUR_TMN_CODE
   vnp_TransactionNo: 12345678
   vnp_TransactionStatus: 00
   vnp_TxnRef: TRAVIA11234567890
   vnp_SecureHash: [tính toán signature]
   ```

4. Click **"Execute"**
5. Response:
   - **200 OK:** "Transaction processed successfully" hoặc "Transaction already processed"
   - **400 Bad Request:** "Invalid signature"
   - **404 Not Found:** "Transaction not found"

### 7.5. Test Flow hoàn chỉnh trong Swagger

**Scenario 1: Thanh toán thành công**

1. ✅ Login và authorize trong Swagger
2. ✅ Tạo payment URL với `booking_id` hợp lệ
3. ✅ Copy `payment_url` và mở trong browser
4. ✅ Thanh toán với test card thành công
5. ✅ Kiểm tra Return URL callback (tự động)
6. ✅ Kiểm tra IPN callback (tự động hoặc test thủ công)
7. ✅ Verify booking status đã được cập nhật thành `da_thanh_toan`

**Scenario 2: Thanh toán thất bại**

1. ✅ Tạo payment URL
2. ✅ Thanh toán với test card thất bại (`9704198526191432199`)
3. ✅ Kiểm tra Return URL với `vnp_ResponseCode != "00"`
4. ✅ Verify booking status vẫn là `cho_xac_nhan` hoặc `da_xac_nhan`
5. ✅ Verify transaction status là `that_bai`

### 7.6. Debugging trong Swagger

**Kiểm tra logs:**
```bash
# Xem logs của server khi test
# Logs sẽ hiển thị:
# - Request details
# - Signature verification
# - Database updates
# - Errors (nếu có)
```

**Common Issues:**

1. **"Unauthorized" (401):**
   - ✅ Đảm bảo đã login và authorize trong Swagger
   - ✅ Kiểm tra cookie/token còn valid không

2. **"Booking not found" (404):**
   - ✅ Kiểm tra `booking_id` có tồn tại không
   - ✅ Kiểm tra booking có thuộc về user đã đăng nhập không

3. **"Invalid signature" (400):**
   - ✅ Kiểm tra `VNPAY_HASH_SECRET` trong `.env`
   - ✅ Kiểm tra query params có bị thay đổi không
   - ✅ Verify signature calculation

4. **"Transaction not found" (404):**
   - ✅ Kiểm tra `vnp_TxnRef` có đúng format không
   - ✅ Kiểm tra transaction có được tạo trong DB không

### 7.7. Test với Sandbox VNPay

1. **Đăng ký tài khoản sandbox:**
   - Truy cập: https://sandbox.vnpayment.vn/
   - Đăng ký tài khoản test
   - Lấy `TMN Code` và `Hash Secret`

2. **Test card numbers:**
   - Thẻ thành công: `9704198526191432198`
   - Thẻ thất bại: `9704198526191432199`
   - OTP: `123456`

3. **Cấu hình trong `.env`:**
   ```env
   VNPAY_TMN_CODE=your_sandbox_tmn_code
   VNPAY_HASH_SECRET=your_sandbox_hash_secret
   VNPAY_PAYMENT_URL=https://sandbox.vnpayment.vn/paymentv2/vpcpay.html
   ```

---

## 8. Troubleshooting

### Lỗi "Invalid signature":
- Kiểm tra `VNPAY_HASH_SECRET` có đúng không
- Kiểm tra query params có bị thay đổi không
- Kiểm tra encoding (phải là UTF-8)

### Lỗi "Transaction not found":
- Kiểm tra `vnp_TxnRef` có đúng format không
- Kiểm tra transaction có được tạo trong DB không

### IPN không được gọi:
- Kiểm tra `VNPAY_IPN_URL` phải là URL public
- Kiểm tra firewall/security group
- Kiểm tra VNPay dashboard có cấu hình IPN URL không

### Return URL không redirect:
- Kiểm tra `VNPAY_RETURN_URL` có đúng không
- Kiểm tra frontend route có tồn tại không

---

## 9. Security Best Practices

1. **Luôn verify signature** trước khi xử lý callback
2. **Sử dụng HTTPS** trong production
3. **Validate tất cả input** từ VNPay
4. **Xử lý idempotent** để tránh duplicate processing
5. **Log tất cả transactions** để audit
6. **Giới hạn số lần retry** cho IPN callback

---

## 10. Ví dụ code hoàn chỉnh

### Frontend (React):
```typescript
// services/paymentService.ts
export const paymentService = {
  createVNPayPayment: async (bookingId: number, returnUrl?: string) => {
    const response = await api.post('/payment/vnpay/create', {
      booking_id: bookingId,
      return_url: returnUrl || window.location.origin + '/payment/vnpay/return'
    });
    return response.data;
  }
};

// components/PaymentButton.tsx
function PaymentButton({ bookingId }: { bookingId: number }) {
  const handlePayment = async () => {
    try {
      const { payment_url } = await paymentService.createVNPayPayment(bookingId);
      window.location.href = payment_url;
    } catch (error) {
      console.error('Payment error:', error);
    }
  };

  return <button onClick={handlePayment}>Thanh toán VNPay</button>;
}
```

### Backend (đã implement):
- Xem file `payment.go` để tham khảo implementation chi tiết

---

## 11. Tài liệu tham khảo

- [VNPay Integration Guide](https://sandbox.vnpayment.vn/apis/)
- [VNPay API Documentation](https://sandbox.vnpayment.vn/apis/docs/thanh-toan-pay.html)

---

**Lưu ý:** Đây là tài liệu hướng dẫn cơ bản. Vui lòng tham khảo tài liệu chính thức của VNPay để biết thêm chi tiết.

