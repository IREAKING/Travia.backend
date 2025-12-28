# 🎨 Hướng Dẫn Triển Khai Frontend cho Hệ Thống Phân Quyền

## Tổng Quan

Với 3 endpoint đăng nhập riêng biệt, bạn có **2 phương án** để thiết kế giao diện frontend:

### ✅ Phương Án 1: 3 Giao Diện Login Riêng Biệt (KHUYẾN NGHỊ)
### ⚠️ Phương Án 2: 1 Giao Diện với Dropdown Chọn Vai Trò

---

## 📱 Phương Án 1: 3 Giao Diện Riêng Biệt (Best Practice)

### Tại sao nên chọn phương án này?

✅ **Ưu điểm:**
1. **Bảo mật cao hơn**: Mỗi loại user có URL riêng, khó bị nhầm lẫn
2. **UX rõ ràng**: User biết chính xác mình đang ở đâu
3. **Branding tốt hơn**: Mỗi portal có theme/logo riêng
4. **Dễ quản lý**: Tách biệt logic, dễ maintain
5. **SEO tốt hơn**: Mỗi trang có meta tags riêng
6. **Analytics rõ ràng**: Theo dõi traffic từng loại user

❌ **Nhược điểm:**
- Cần tạo 3 trang/components
- Code có thể duplicate một chút (nhưng có thể reuse component)

### Cấu trúc Routes

```javascript
// React Router example
const routes = [
  // Landing page
  { path: '/', component: HomePage },
  
  // Customer/User login
  { path: '/login', component: UserLoginPage },
  { path: '/login/user', component: UserLoginPage }, // alias
  
  // Admin login
  { path: '/admin/login', component: AdminLoginPage },
  
  // Supplier login
  { path: '/supplier/login', component: SupplierLoginPage },
  
  // After login redirects
  { path: '/dashboard', component: UserDashboard },
  { path: '/admin/dashboard', component: AdminDashboard },
  { path: '/supplier/dashboard', component: SupplierDashboard },
]
```

### Cấu trúc Thư Mục

```
src/
├── pages/
│   ├── auth/
│   │   ├── UserLogin.jsx          // Khách hàng
│   │   ├── AdminLogin.jsx         // Admin
│   │   └── SupplierLogin.jsx      // Nhà cung cấp
│   ├── user/
│   │   └── Dashboard.jsx
│   ├── admin/
│   │   └── Dashboard.jsx
│   └── supplier/
│       └── Dashboard.jsx
├── components/
│   ├── auth/
│   │   ├── LoginForm.jsx          // Shared form component
│   │   └── LoginLayout.jsx        // Shared layout
│   └── ...
├── services/
│   └── authService.js             // API calls
└── hooks/
    └── useAuth.js                 // Custom hook
```

### Ví Dụ Implementation

#### 1. Shared Login Form Component

```jsx
// components/auth/LoginForm.jsx
import React, { useState } from 'react';

export const LoginForm = ({ onSubmit, title, subtitle, isLoading }) => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');

  const handleSubmit = (e) => {
    e.preventDefault();
    onSubmit({ email, password });
  };

  return (
    <div className="login-form">
      <h1>{title}</h1>
      <p>{subtitle}</p>
      
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label>Email</label>
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="name@example.com"
            required
          />
        </div>

        <div className="form-group">
          <label>Mật khẩu</label>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="••••••••"
            required
          />
        </div>

        <button type="submit" disabled={isLoading}>
          {isLoading ? 'Đang đăng nhập...' : 'Đăng nhập'}
        </button>
      </form>
    </div>
  );
};
```

#### 2. Auth Service

```javascript
// services/authService.js
const API_BASE = 'http://localhost:8080/api';

export const authService = {
  // Đăng nhập khách hàng
  loginUser: async (email, password) => {
    const response = await fetch(`${API_BASE}/auth/login/user`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ email, password })
    });
    
    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Đăng nhập thất bại');
    }
    
    return await response.json();
  },

  // Đăng nhập admin
  loginAdmin: async (email, password) => {
    const response = await fetch(`${API_BASE}/auth/login/admin`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ email, password })
    });
    
    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Đăng nhập thất bại');
    }
    
    return await response.json();
  },

  // Đăng nhập nhà cung cấp
  loginSupplier: async (email, password) => {
    const response = await fetch(`${API_BASE}/auth/login/supplier`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ email, password })
    });
    
    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Đăng nhập thất bại');
    }
    
    return await response.json();
  },

  // Đăng xuất
  logout: async () => {
    const response = await fetch(`${API_BASE}/auth/logout`, {
      method: 'POST',
      credentials: 'include',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('accessToken')}`
      }
    });
    
    localStorage.removeItem('accessToken');
    localStorage.removeItem('refreshToken');
    localStorage.removeItem('user');
    
    return await response.json();
  }
};
```

#### 3. User Login Page

```jsx
// pages/auth/UserLogin.jsx
import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { LoginForm } from '../../components/auth/LoginForm';
import { authService } from '../../services/authService';
import './UserLogin.css';

export const UserLoginPage = () => {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  const navigate = useNavigate();

  const handleLogin = async ({ email, password }) => {
    setIsLoading(true);
    setError('');

    try {
      const data = await authService.loginUser(email, password);
      
      // Lưu tokens và user info
      localStorage.setItem('accessToken', data.tokens.accessToken);
      localStorage.setItem('refreshToken', data.tokens.refreshToken);
      localStorage.setItem('user', JSON.stringify(data.user));
      
      // Redirect to user dashboard
      navigate('/dashboard');
    } catch (err) {
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="user-login-page">
      <div className="login-container">
        <div className="login-left">
          <img src="/images/user-login-bg.jpg" alt="Travel" />
        </div>
        
        <div className="login-right">
          <LoginForm
            title="Đăng Nhập"
            subtitle="Khám phá những chuyến đi tuyệt vời"
            onSubmit={handleLogin}
            isLoading={isLoading}
          />
          
          {error && <div className="error-message">{error}</div>}
          
          <div className="login-footer">
            <a href="/forgot-password">Quên mật khẩu?</a>
            <p>
              Chưa có tài khoản? <a href="/register">Đăng ký ngay</a>
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};
```

#### 4. Admin Login Page

```jsx
// pages/auth/AdminLogin.jsx
import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { LoginForm } from '../../components/auth/LoginForm';
import { authService } from '../../services/authService';
import './AdminLogin.css';

export const AdminLoginPage = () => {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  const navigate = useNavigate();

  const handleLogin = async ({ email, password }) => {
    setIsLoading(true);
    setError('');

    try {
      const data = await authService.loginAdmin(email, password);
      
      localStorage.setItem('accessToken', data.tokens.accessToken);
      localStorage.setItem('refreshToken', data.tokens.refreshToken);
      localStorage.setItem('user', JSON.stringify(data.user));
      
      // Redirect to admin dashboard
      navigate('/admin/dashboard');
    } catch (err) {
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="admin-login-page">
      <div className="login-container admin-theme">
        <div className="login-box">
          <div className="admin-logo">
            <img src="/images/admin-logo.svg" alt="Admin" />
          </div>
          
          <LoginForm
            title="Quản Trị Viên"
            subtitle="Đăng nhập vào hệ thống quản lý"
            onSubmit={handleLogin}
            isLoading={isLoading}
          />
          
          {error && (
            <div className="error-message admin-error">{error}</div>
          )}
          
          <div className="admin-notice">
            <p>⚠️ Chỉ dành cho quản trị viên</p>
          </div>
        </div>
      </div>
    </div>
  );
};
```

#### 5. Supplier Login Page

```jsx
// pages/auth/SupplierLogin.jsx
import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { LoginForm } from '../../components/auth/LoginForm';
import { authService } from '../../services/authService';
import './SupplierLogin.css';

export const SupplierLoginPage = () => {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  const navigate = useNavigate();

  const handleLogin = async ({ email, password }) => {
    setIsLoading(true);
    setError('');

    try {
      const data = await authService.loginSupplier(email, password);
      
      localStorage.setItem('accessToken', data.tokens.accessToken);
      localStorage.setItem('refreshToken', data.tokens.refreshToken);
      localStorage.setItem('user', JSON.stringify(data.user));
      
      // Redirect to supplier dashboard
      navigate('/supplier/dashboard');
    } catch (err) {
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="supplier-login-page">
      <div className="login-container supplier-theme">
        <div className="login-split">
          <div className="login-info">
            <h2>Partner Portal</h2>
            <p>Quản lý tours và đặt chỗ của bạn</p>
            <ul>
              <li>✓ Quản lý tour của bạn</li>
              <li>✓ Theo dõi đặt chỗ</li>
              <li>✓ Báo cáo doanh thu</li>
            </ul>
          </div>
          
          <div className="login-form-area">
            <LoginForm
              title="Nhà Cung Cấp"
              subtitle="Đăng nhập vào portal đối tác"
              onSubmit={handleLogin}
              isLoading={isLoading}
            />
            
            {error && <div className="error-message">{error}</div>}
            
            <div className="supplier-help">
              <p>Cần hỗ trợ? <a href="/supplier/contact">Liên hệ chúng tôi</a></p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
```

#### 6. Custom Hook cho Authentication

```javascript
// hooks/useAuth.js
import { useState, useEffect, createContext, useContext } from 'react';
import { authService } from '../services/authService';

const AuthContext = createContext();

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Check if user is logged in on mount
    const storedUser = localStorage.getItem('user');
    if (storedUser) {
      setUser(JSON.parse(storedUser));
    }
    setLoading(false);
  }, []);

  const logout = async () => {
    await authService.logout();
    setUser(null);
  };

  return (
    <AuthContext.Provider value={{ user, setUser, logout, loading }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
};
```

---

## 🔄 Phương Án 2: 1 Giao Diện với Dropdown (Alternative)

### Khi nào dùng phương án này?

- Ứng dụng nhỏ, ít user
- Muốn giảm số lượng trang
- Dễ test (1 page thay vì 3)

### Implementation

```jsx
// pages/auth/UnifiedLogin.jsx
import React, { useState } from 'react';
import { authService } from '../../services/authService';

const USER_TYPES = {
  user: {
    label: 'Khách hàng',
    endpoint: authService.loginUser,
    redirect: '/dashboard'
  },
  admin: {
    label: 'Quản trị viên',
    endpoint: authService.loginAdmin,
    redirect: '/admin/dashboard'
  },
  supplier: {
    label: 'Nhà cung cấp',
    endpoint: authService.loginSupplier,
    redirect: '/supplier/dashboard'
  }
};

export const UnifiedLoginPage = () => {
  const [userType, setUserType] = useState('user');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e) => {
    e.preventDefault();
    setIsLoading(true);
    setError('');

    try {
      const loginFn = USER_TYPES[userType].endpoint;
      const data = await loginFn(email, password);
      
      localStorage.setItem('accessToken', data.tokens.accessToken);
      localStorage.setItem('refreshToken', data.tokens.refreshToken);
      localStorage.setItem('user', JSON.stringify(data.user));
      
      window.location.href = USER_TYPES[userType].redirect;
    } catch (err) {
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="unified-login">
      <form onSubmit={handleSubmit}>
        <h1>Đăng Nhập</h1>
        
        {/* User Type Selector */}
        <div className="form-group">
          <label>Loại tài khoản</label>
          <select 
            value={userType} 
            onChange={(e) => setUserType(e.target.value)}
            className="user-type-select"
          >
            {Object.entries(USER_TYPES).map(([key, { label }]) => (
              <option key={key} value={key}>{label}</option>
            ))}
          </select>
        </div>

        <div className="form-group">
          <label>Email</label>
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
        </div>

        <div className="form-group">
          <label>Mật khẩu</label>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
        </div>

        {error && <div className="error">{error}</div>}

        <button type="submit" disabled={isLoading}>
          {isLoading ? 'Đang đăng nhập...' : 'Đăng nhập'}
        </button>
      </form>
    </div>
  );
};
```

---

## 🎯 So Sánh 2 Phương Án

| Tiêu Chí | 3 Giao Diện Riêng | 1 Giao Diện + Dropdown |
|----------|-------------------|------------------------|
| **Bảo mật** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |
| **UX/UI** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |
| **Dễ develop** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Maintain** | ⭐⭐⭐⭐ | ⭐⭐⭐ |
| **Branding** | ⭐⭐⭐⭐⭐ | ⭐⭐ |
| **SEO** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |
| **Professional** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |

---

## 🎨 Thiết Kế UI Khuyến Nghị

### 1. User/Customer Login
- **Theme**: Sáng, thân thiện, màu xanh dương/xanh lá
- **Images**: Ảnh du lịch, phong cảnh đẹp
- **Style**: Modern, minimal, inviting
- **CTA**: "Khám phá ngay", "Đặt tour"

### 2. Admin Login
- **Theme**: Tối, chuyên nghiệp, màu xám/đen
- **Images**: Minimal hoặc không có
- **Style**: Professional, secure-looking
- **Warning**: "⚠️ Chỉ dành cho quản trị viên"

### 3. Supplier/Partner Login
- **Theme**: Business-oriented, màu cam/tím
- **Images**: Office, partnership imagery
- **Style**: Professional, business-like
- **Info**: Highlight về partner benefits

---

## 🔒 Best Practices

### 1. Security
```javascript
// Luôn validate user role sau khi login
const validateUserRole = (user, expectedRole) => {
  if (user.role !== expectedRole) {
    // Log out và redirect
    authService.logout();
    window.location.href = '/';
    throw new Error('Unauthorized access');
  }
};
```

### 2. Protected Routes
```jsx
// components/ProtectedRoute.jsx
import { Navigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';

export const ProtectedRoute = ({ children, allowedRoles }) => {
  const { user, loading } = useAuth();

  if (loading) return <div>Loading...</div>;
  
  if (!user) {
    return <Navigate to="/login" />;
  }

  if (allowedRoles && !allowedRoles.includes(user.role)) {
    return <Navigate to="/unauthorized" />;
  }

  return children;
};

// Usage
<Route path="/admin/dashboard" element={
  <ProtectedRoute allowedRoles={['quan_tri']}>
    <AdminDashboard />
  </ProtectedRoute>
} />
```

### 3. Error Handling
```javascript
// Xử lý error 403 - Wrong role
if (error.includes('không có quyền')) {
  setError('Bạn đang sử dụng sai trang đăng nhập. Vui lòng chọn đúng loại tài khoản.');
}

// Hiển thị link đến các trang login khác
<div className="login-alternatives">
  <p>Có phải bạn muốn:</p>
  <a href="/admin/login">Đăng nhập Admin?</a>
  <a href="/supplier/login">Đăng nhập Nhà cung cấp?</a>
</div>
```

---

## 📱 Responsive Design

### Mobile Considerations
```css
/* Trên mobile, các trang login nên:
   - Full screen
   - Stack vertically
   - Larger touch targets
   - Easy to type
*/

@media (max-width: 768px) {
  .login-container {
    flex-direction: column;
    padding: 20px;
  }
  
  .login-left {
    display: none; /* Ẩn ảnh background trên mobile */
  }
  
  .login-form input {
    font-size: 16px; /* Tránh zoom trên iOS */
    padding: 12px;
  }
}
```

---

## 🚀 Deployment URLs

### Production URLs Khuyến Nghị
```
https://travia.com              → Homepage
https://travia.com/login        → User login
https://admin.travia.com        → Admin portal (subdomain)
https://partner.travia.com      → Supplier portal (subdomain)
```

### Alternative (Single Domain)
```
https://travia.com              → Homepage
https://travia.com/login        → User login
https://travia.com/admin        → Admin portal
https://travia.com/supplier     → Supplier portal
```

---

## ✅ Checklist Implementation

- [ ] Tạo 3 login pages (hoặc 1 unified)
- [ ] Implement authService với 3 methods
- [ ] Setup routing cho các pages
- [ ] Tạo protected routes
- [ ] Implement error handling
- [ ] Thiết kế UI cho từng role
- [ ] Test cross-browser
- [ ] Test responsive
- [ ] Add loading states
- [ ] Add success notifications
- [ ] Implement remember me (optional)
- [ ] Add forgot password links
- [ ] Test accessibility (a11y)
- [ ] Add analytics tracking

---

## 🎓 Kết Luận

**Khuyến nghị: Sử dụng Phương án 1 - 3 Giao diện riêng biệt**

Lý do:
1. Professional hơn cho sản phẩm thương mại
2. Bảo mật tốt hơn
3. UX tốt hơn cho từng loại user
4. Dễ scale và customize sau này
5. Phù hợp với best practices của các platform lớn

---

**Cần hỗ trợ thêm về implementation?** Hỏi ngay! 🚀

