# 🔐 Hướng Dẫn Hệ Thống Phân Quyền Đăng Nhập

## Tổng Quan

Hệ thống Travia đã được cập nhật với hệ thống phân quyền đăng nhập cho 3 loại người dùng:

1. **Khách hàng/Người dùng** (`khach_hang`)
2. **Quản trị viên/Admin** (`quan_tri`)
3. **Nhà cung cấp/Supplier** (`nha_cung_cap`)

## Các Endpoint Đăng Nhập

### 1. Đăng nhập cho Khách hàng
**Endpoint:** `POST /api/auth/login/user`

**Mô tả:** Dành cho người dùng thông thường/khách hàng

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "yourpassword"
}
```

**Response Success (200):**
```json
{
  "message": "Đăng nhập thành công",
  "user": {
    "id": "uuid-here",
    "email": "user@example.com",
    "name": "Nguyen Van A",
    "role": "khach_hang"
  },
  "tokens": {
    "accessToken": "jwt-token-here",
    "refreshToken": "refresh-token-here"
  }
}
```

**Response Error (403):**
```json
{
  "error": "Bạn không có quyền đăng nhập vào hệ thống này"
}
```

---

### 2. Đăng nhập cho Admin
**Endpoint:** `POST /api/auth/login/admin`

**Mô tả:** Dành cho quản trị viên hệ thống

**Request Body:**
```json
{
  "email": "admin@example.com",
  "password": "adminpassword"
}
```

**Response Success (200):**
```json
{
  "message": "Đăng nhập thành công",
  "user": {
    "id": "uuid-here",
    "email": "admin@example.com",
    "name": "Admin Name",
    "role": "quan_tri"
  },
  "tokens": {
    "accessToken": "jwt-token-here",
    "refreshToken": "refresh-token-here"
  }
}
```

**Response Error (403):**
```json
{
  "error": "Bạn không có quyền đăng nhập vào hệ thống này"
}
```

---

### 3. Đăng nhập cho Nhà Cung Cấp
**Endpoint:** `POST /api/auth/login/supplier`

**Mô tả:** Dành cho nhà cung cấp tour

**Request Body:**
```json
{
  "email": "supplier@example.com",
  "password": "supplierpassword"
}
```

**Response Success (200):**
```json
{
  "message": "Đăng nhập thành công",
  "user": {
    "id": "uuid-here",
    "email": "supplier@example.com",
    "name": "Supplier Name",
    "role": "nha_cung_cap"
  },
  "tokens": {
    "accessToken": "jwt-token-here",
    "refreshToken": "refresh-token-here"
  }
}
```

**Response Error (403):**
```json
{
  "error": "Bạn không có quyền đăng nhập vào hệ thống này"
}
```

---

### 4. Đăng nhập Chung (Deprecated)
**Endpoint:** `POST /api/auth/login`

⚠️ **Deprecated:** Endpoint này vẫn hoạt động để đảm bảo backward compatibility, nhưng khuyến khích sử dụng các endpoint cụ thể theo vai trò ở trên.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "password"
}
```

---

## Các HTTP Status Codes

| Status Code | Ý Nghĩa |
|------------|---------|
| 200 | Đăng nhập thành công |
| 400 | Dữ liệu đầu vào không hợp lệ |
| 401 | Email hoặc mật khẩu không chính xác / Tài khoản bị khóa |
| 403 | Không có quyền đăng nhập vào hệ thống này (sai vai trò) |
| 500 | Lỗi hệ thống |

---

## Cơ Chế Bảo Mật

### 1. Kiểm tra Vai Trò
Mỗi endpoint đăng nhập sẽ:
- Xác thực thông tin đăng nhập (email + password)
- Kiểm tra tài khoản có đang hoạt động không
- **Kiểm tra vai trò của user có khớp với endpoint không**
- Chỉ cho phép đăng nhập nếu vai trò khớp

### 2. JWT Tokens
Sau khi đăng nhập thành công:
- **Access Token**: Thời hạn 7 ngày
- **Refresh Token**: Thời hạn 30 ngày
- Tokens được lưu trong cookies với flags:
  - `HttpOnly`: true (bảo vệ khỏi XSS)
  - `Secure`: false (dev mode) / true (production)

### 3. Kiểm tra Trạng Thái Tài Khoản
- Tài khoản bị khóa (`DangHoatDong = false`) không thể đăng nhập
- Thông báo lỗi: "Tài khoản đã bị khóa"

---

## Ví Dụ Sử Dụng

### cURL

#### Đăng nhập Khách hàng
```bash
curl -X POST http://localhost:8080/api/auth/login/user \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

#### Đăng nhập Admin
```bash
curl -X POST http://localhost:8080/api/auth/login/admin \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "adminpass"
  }'
```

#### Đăng nhập Nhà Cung Cấp
```bash
curl -X POST http://localhost:8080/api/auth/login/supplier \
  -H "Content-Type: application/json" \
  -d '{
    "email": "supplier@example.com",
    "password": "supplierpass"
  }'
```

---

### JavaScript/Fetch API

```javascript
// Đăng nhập Khách hàng
async function loginUser(email, password) {
  const response = await fetch('http://localhost:8080/api/auth/login/user', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ email, password }),
    credentials: 'include' // Quan trọng để lưu cookies
  });
  
  const data = await response.json();
  
  if (response.ok) {
    console.log('Đăng nhập thành công:', data);
    // Lưu tokens nếu cần
    localStorage.setItem('accessToken', data.tokens.accessToken);
  } else {
    console.error('Đăng nhập thất bại:', data.error);
  }
  
  return data;
}

// Đăng nhập Admin
async function loginAdmin(email, password) {
  const response = await fetch('http://localhost:8080/api/auth/login/admin', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ email, password }),
    credentials: 'include'
  });
  
  return await response.json();
}

// Đăng nhập Nhà Cung Cấp
async function loginSupplier(email, password) {
  const response = await fetch('http://localhost:8080/api/auth/login/supplier', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ email, password }),
    credentials: 'include'
  });
  
  return await response.json();
}
```

---

## Xử Lý Lỗi

### Lỗi 403 - Sai Vai Trò
Khi user cố gắng đăng nhập vào endpoint không phù hợp với vai trò:

```json
{
  "error": "Bạn không có quyền đăng nhập vào hệ thống này"
}
```

**Giải pháp:**
- Khách hàng phải dùng `/api/auth/login/user`
- Admin phải dùng `/api/auth/login/admin`
- Nhà cung cấp phải dùng `/api/auth/login/supplier`

### Lỗi 401 - Sai Thông Tin
```json
{
  "error": "Email hoặc mật khẩu không chính xác"
}
```

### Lỗi 401 - Tài Khoản Bị Khóa
```json
{
  "error": "Tài khoản đã bị khóa"
}
```

---

## Migration từ Endpoint Cũ

Nếu bạn đang sử dụng endpoint cũ `/api/auth/login`, hãy cập nhật sang endpoint mới:

### Trước (Deprecated)
```javascript
POST /api/auth/login
```

### Sau (Khuyến nghị)
```javascript
// Tùy theo loại user
POST /api/auth/login/user      // Khách hàng
POST /api/auth/login/admin     // Admin
POST /api/auth/login/supplier  // Nhà cung cấp
```

---

## Testing

### Test với Postman/Thunder Client

1. **Tạo 3 users với vai trò khác nhau trong database:**
   - User với vai trò `khach_hang`
   - User với vai trò `quan_tri`
   - User với vai trò `nha_cung_cap`

2. **Test Case 1: Đăng nhập đúng vai trò**
   - ✅ User `khach_hang` login vào `/login/user` → Success
   - ✅ User `quan_tri` login vào `/login/admin` → Success
   - ✅ User `nha_cung_cap` login vào `/login/supplier` → Success

3. **Test Case 2: Đăng nhập sai vai trò**
   - ❌ User `khach_hang` login vào `/login/admin` → Error 403
   - ❌ User `quan_tri` login vào `/login/supplier` → Error 403
   - ❌ User `nha_cung_cap` login vào `/login/user` → Error 403

4. **Test Case 3: Sai thông tin đăng nhập**
   - ❌ Email sai → Error 401
   - ❌ Password sai → Error 401

5. **Test Case 4: Tài khoản bị khóa**
   - ❌ User với `DangHoatDong = false` → Error 401

---

## Câu Hỏi Thường Gặp (FAQ)

### Q1: Tại sao cần tách endpoint theo vai trò?
**A:** Để tăng cường bảo mật và rõ ràng hơn trong việc phân quyền. Mỗi loại người dùng sẽ có giao diện đăng nhập riêng và endpoint riêng, tránh nhầm lẫn và tăng khả năng kiểm soát.

### Q2: Endpoint cũ `/api/auth/login` còn hoạt động không?
**A:** Có, endpoint này vẫn hoạt động để đảm bảo backward compatibility, nhưng nên chuyển sang sử dụng endpoint mới.

### Q3: Làm sao để tạo user với vai trò khác nhau?
**A:** Khi tạo user qua API, vai trò mặc định là `khach_hang`. Admin có thể cập nhật vai trò trong database hoặc sẽ có API riêng để tạo admin/supplier (cần implement thêm).

### Q4: Token có khác nhau giữa các vai trò không?
**A:** Không, cấu trúc token giống nhau, nhưng trong token sẽ chứa thông tin vai trò để backend có thể phân quyền các API khác.

### Q5: Làm sao để kiểm tra vai trò từ token?
**A:** Token JWT đã chứa thông tin vai trò trong claims. Middleware `AuthMiddleware` và `RequireRoles` đã xử lý việc này.

---

## Thông Tin Kỹ Thuật

### Implementation Details

**File:** `api/handler/auth.go`

**Helper Function:**
```go
func (s *Server) loginWithRole(c *gin.Context, requiredRole db.VaiTroNguoiDung)
```

**Login Functions:**
- `LoginUser()` → Calls `loginWithRole(c, db.VaiTroNguoiDungKhachHang)`
- `LoginAdmin()` → Calls `loginWithRole(c, db.VaiTroNguoiDungQuanTri)`
- `LoginSupplier()` → Calls `loginWithRole(c, db.VaiTroNguoiDungNhaCungCap)`

**Role Constants:**
- `db.VaiTroNguoiDungKhachHang` = `"khach_hang"`
- `db.VaiTroNguoiDungQuanTri` = `"quan_tri"`
- `db.VaiTroNguoiDungNhaCungCap` = `"nha_cung_cap"`

---

## Support

Nếu có vấn đề hoặc câu hỏi, vui lòng liên hệ team phát triển hoặc tạo issue trong repository.

---

**Last Updated:** October 12, 2025
**Version:** 1.0.0

