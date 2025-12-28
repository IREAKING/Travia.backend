# 🚀 Quick Start - Hệ Thống Phân Quyền Đăng Nhập

## Tóm Tắt Nhanh

Hệ thống đã được **phân quyền đăng nhập** cho 3 loại người dùng với **3 endpoint riêng biệt**.

---

## 📍 3 Endpoints Mới

```bash
# 1. Khách hàng
POST /api/auth/login/user

# 2. Admin  
POST /api/auth/login/admin

# 3. Nhà cung cấp
POST /api/auth/login/supplier
```

---

## 🎯 Câu Trả Lời Câu Hỏi Của Bạn

> **"Tách ra 3 endpoint, khi viết frontend thì viết 3 giao diện login và gọi 3 endpoint à?"**

### ✅ Đáp án: **CÓ - Nên tạo 3 trang login riêng**

#### Ví dụ thực tế:

**1. Trang dành cho Khách hàng** 
- URL: `https://travia.com/login`
- Gọi API: `POST /api/auth/login/user`
- Theme: Sáng sủa, thân thiện, ảnh du lịch đẹp
- Redirect: `/dashboard` (trang của khách)

**2. Trang dành cho Admin**
- URL: `https://travia.com/admin/login` hoặc `https://admin.travia.com`  
- Gọi API: `POST /api/auth/login/admin`
- Theme: Tối, chuyên nghiệp, có cảnh báo "Chỉ dành cho admin"
- Redirect: `/admin/dashboard`

**3. Trang dành cho Nhà Cung Cấp**
- URL: `https://travia.com/supplier/login` hoặc `https://partner.travia.com`
- Gọi API: `POST /api/auth/login/supplier`  
- Theme: Business-style, thông tin về partner benefits
- Redirect: `/supplier/dashboard`

---

## 💡 Tại Sao 3 Trang Riêng?

✅ **Bảo mật**: Mỗi loại user có URL riêng, khó nhầm lẫn  
✅ **UX tốt hơn**: User biết rõ mình đang ở đâu  
✅ **Professional**: Giống các platform lớn (Shopify, WordPress, AWS...)  
✅ **Branding**: Mỗi portal có theme/logo riêng  
✅ **Dễ quản lý**: Code tách biệt, dễ maintain  

---

## 🔄 Hoặc Dùng 1 Trang (Alternative)

Nếu muốn đơn giản hơn, có thể dùng **1 trang với dropdown chọn loại user**:

```jsx
<select>
  <option value="user">Khách hàng</option>
  <option value="admin">Admin</option>
  <option value="supplier">Nhà cung cấp</option>
</select>
```

Nhưng cách này **KHÔNG khuyến nghị** vì:
- ❌ Kém chuyên nghiệp
- ❌ Dễ nhầm lẫn
- ❌ Kém bảo mật

---

## 📝 Code Example Nhanh

### Frontend React

```jsx
// pages/UserLogin.jsx
import { authService } from '../services/authService';

function UserLogin() {
  const handleLogin = async (email, password) => {
    const data = await authService.loginUser(email, password);
    // Lưu token và redirect
    localStorage.setItem('accessToken', data.tokens.accessToken);
    window.location.href = '/dashboard';
  };
  
  return <LoginForm onSubmit={handleLogin} />;
}
```

### Auth Service

```javascript
// services/authService.js
export const authService = {
  loginUser: (email, password) => 
    fetch('http://localhost:8080/api/auth/login/user', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password })
    }),
    
  loginAdmin: (email, password) =>
    fetch('http://localhost:8080/api/auth/login/admin', {...}),
    
  loginSupplier: (email, password) =>
    fetch('http://localhost:8080/api/auth/login/supplier', {...})
};
```

---

## 📚 Tài Liệu Chi Tiết

| File | Nội dung |
|------|----------|
| `AUTHENTICATION_GUIDE.md` | API documentation chi tiết |
| `FRONTEND_AUTHENTICATION_GUIDE.md` | Hướng dẫn code frontend đầy đủ |
| `test_authentication.http` | Test requests mẫu |
| `IMPLEMENTATION_SUMMARY.md` | Tổng quan implementation |

---

## ✅ Backend Đã Xong!

- ✅ 3 endpoints đã implement
- ✅ Logic phân quyền hoàn chỉnh  
- ✅ Code đã compile thành công
- ✅ Documentation đầy đủ

### 🎯 Bước Tiếp Theo:

1. **Tạo test users trong database** (xem `db/data/test_users_authentication.sql`)
2. **Test API** với `test_authentication.http`
3. **Implement frontend** theo `FRONTEND_AUTHENTICATION_GUIDE.md`

---

## 🧪 Test Nhanh

```bash
# Khách hàng
curl -X POST http://localhost:8080/api/auth/login/user \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password"}'

# Admin
curl -X POST http://localhost:8080/api/auth/login/admin \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password"}'
```

---

## 🎓 Tóm Lại

### Backend (✅ Đã xong)
- 3 endpoints riêng biệt
- Kiểm tra vai trò khi đăng nhập
- Return 403 nếu sai vai trò

### Frontend (⏭️ Cần làm)
- **Khuyến nghị**: 3 trang login riêng
- Mỗi trang gọi endpoint tương ứng
- Redirect đến dashboard riêng sau login

---

**Cần giúp gì thêm về frontend implementation?** 🚀

