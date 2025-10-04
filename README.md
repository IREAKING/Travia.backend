# Travia Backend API

Backend API cho hệ thống quản lý tour du lịch Travia, được xây dựng bằng Go (Gin Framework) với PostgreSQL và Redis.

## 🚀 Tech Stack

- **Language:** Go 1.25+
- **Framework:** Gin Web Framework
- **Database:** PostgreSQL (Supabase)
- **Cache:** Redis
- **ORM:** sqlc (Type-safe SQL)
- **Authentication:** JWT + OAuth2 (Google)
- **Documentation:** Swagger/OpenAPI

## 📁 Cấu trúc dự án

```
.
├── api/
│   ├── handler/          # HTTP handlers
│   │   ├── auth.go      # Authentication & user management
│   │   ├── tour.go      # Tour management
│   │   ├── admin.go     # Admin statistics & analytics
│   │   ├── router.go    # Route definitions
│   │   └── server.go    # Server setup
│   ├── middleware/      # Middleware (JWT, CORS, timeout, etc.)
│   ├── models/          # Request/response models
│   ├── helpers/         # Helper functions
│   └── utils/           # Utilities (JWT, hash, etc.)
├── config/              # Configuration management
├── db/
│   ├── migration/       # Database schema
│   ├── query/           # SQL queries (sqlc)
│   └── sqlc/            # Generated Go code from SQL
├── docs/                # Swagger documentation
└── main.go
```

## ⚙️ Cài đặt

### Prerequisites

- Go 1.25+
- PostgreSQL 14+
- Redis 7+
- sqlc CLI

### Environment Variables

Tạo file `env/.env`:

```env
# Database
DB_HOST=your-db-host
DB_PORT=5432
DB_USER=your-db-user
DB_PASSWORD=your-db-password
DB_NAME=your-db-name
DB_SSLMODE=require

# Server
PORT_NUMBER=8080
HOST=0.0.0.0
ENVIRONMENT=development
SECRET_KEY=your-secret-key
API_SECRET=your-api-secret

# Redis
REDIS_ADDRESS=your-redis-address:6379
REDIS_DB=0
REDIS_PASSWORD=your-redis-password

# Email (for OTP)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
FROM_EMAIL=your-email@gmail.com
FROM_NAME=Travia

# Google OAuth
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_REDIRECT_URIS=http://localhost:8080/api/auth/oauth/google/callback

# Supabase Storage
SUPABASE_URL=your-supabase-url
SUPABASE_KEY_ROLE=your-supabase-key
SUPABASE_BUCKET=your-bucket-name
SUPABASE_SERVICE_KEY=your-service-key
```

### Installation

```bash
# Clone repository
git clone <repository-url>
cd Travia.backend

# Install dependencies
go mod download

# Generate sqlc code
sqlc generate

# Run migrations
psql $DATABASE_URL -f db/migration/schema.sql

# Run server
go run main.go
```

## 🔐 Authentication & Authorization

### Vai trò (Roles)

- **`khach_hang`** - Khách hàng thông thường
- **`nha_cung_cap`** - Nhà cung cấp tour
- **`quan_tri`** - Quản trị viên

### Authentication Flow

1. **Đăng ký:** `POST /api/auth/createUserForm` → nhận OTP qua email
2. **Xác thực OTP:** `POST /api/auth/createUser`
3. **Đăng nhập:** `POST /api/auth/login` → nhận JWT token
4. **Google OAuth:** `GET /api/auth/oauth/google`

### Protected Routes

Tất cả các route yêu cầu xác thực phải gửi JWT token trong header:

```
Authorization: Bearer <your_jwt_token>
```

## 📡 API Endpoints

### 🔑 Authentication (`/api/auth`)

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/createUserForm` | Đăng ký (gửi OTP) | ❌ |
| POST | `/createUser` | Xác thực OTP và tạo tài khoản | ❌ |
| POST | `/login` | Đăng nhập | ❌ |
| POST | `/logout` | Đăng xuất | ✅ |
| GET | `/getUserById/:id` | Lấy thông tin user | ✅ |
| PUT | `/updateUserById/:id` | Cập nhật user (self or admin) | ✅ |
| GET | `/oauth/:provider` | OAuth login | ❌ |
| GET | `/oauth/:provider/callback` | OAuth callback | ❌ |

### 🏖️ Tours (`/api/tour`)

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/getAllTourCategory` | Lấy danh sách danh mục tour | ❌ |
| GET | `/getAllTour` | Lấy danh sách tour | ❌ |
| GET | `/getTourDetailByID/:id` | Chi tiết tour | ❌ |

### 📊 Admin Analytics (`/api/admin`)

**⚠️ Tất cả endpoints yêu cầu role `quan_tri`**

#### Dashboard & Tổng quan

| Endpoint | Description | Query Params |
|----------|-------------|--------------|
| `GET /getAdminSummary` | Tổng quan hệ thống (users, tours, bookings, revenue, ratings) | - |

#### Phân tích Doanh thu

| Endpoint | Description | Query Params |
|----------|-------------|--------------|
| `GET /getRevenueByMonth` | Doanh thu theo từng tháng trong năm | `year` (required) |
| `GET /getRevenueByYear` | Doanh thu theo năm | `year` (required) |
| `GET /getRevenueByDateRange` | Doanh thu trong khoảng thời gian | `from`, `to` (YYYY-MM-DD) |
| `GET /getRevenueBySupplier` | Doanh thu theo nhà cung cấp | `limit` (default: 10) |

#### Đặt chỗ & Tour

| Endpoint | Description | Query Params |
|----------|-------------|--------------|
| `GET /getBookingsByStatus` | Thống kê đặt chỗ theo trạng thái | - |
| `GET /getBookingsByMonth` | Số lượng đặt chỗ theo tháng | `year` (required) |
| `GET /getTopToursByBookings` | Top tour có nhiều đặt chỗ nhất | `limit` (default: 10) |
| `GET /getToursByCategory` | Thống kê tour theo danh mục | - |
| `GET /getUpcomingDepartures` | Lịch khởi hành sắp tới (còn chỗ) | `limit` (default: 20) |

#### Người dùng & Khách hàng

| Endpoint | Description | Query Params |
|----------|-------------|--------------|
| `GET /getNewUsersByMonth` | Số người dùng mới theo tháng | `year` (required) |
| `GET /getUserGrowth` | Tăng trưởng người dùng theo ngày | `from`, `to` (YYYY-MM-DD) |
| `GET /getTopCustomers` | Top khách hàng chi tiêu nhiều nhất | `limit` (default: 10) |

#### Nhà cung cấp & Đánh giá

| Endpoint | Description | Query Params |
|----------|-------------|--------------|
| `GET /getTopSuppliers` | Top nhà cung cấp theo tour & bookings | `limit` (default: 10) |
| `GET /getReviewStatsByTour` | Chi tiết phân bố rating (1-5 sao) | `limit` (default: 10) |

## 📝 Example Usage

### Đăng nhập Admin

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "your-password"
  }'
```

Response:
```json
{
  "message": "Đăng nhập thành công",
  "user": {
    "id": "uuid",
    "email": "admin@example.com",
    "name": "Admin User",
    "role": "quan_tri"
  },
  "tokens": {
    "accessToken": "eyJhbGc...",
    "refreshToken": "eyJhbGc..."
  }
}
```

### Lấy tổng quan hệ thống (Admin)

```bash
curl http://localhost:8080/api/admin/getAdminSummary \
  -H "Authorization: Bearer <your_token>"
```

Response:
```json
{
  "message": "Lấy tổng quan admin thành công",
  "data": {
    "total_users": 1500,
    "active_users": 1200,
    "total_tours": 250,
    "active_tours": 180,
    "total_bookings": 3500,
    "total_revenue": 125000000,
    "avg_rating": 4.5
  }
}
```

### Doanh thu theo tháng

```bash
curl "http://localhost:8080/api/admin/getRevenueByMonth?year=2024" \
  -H "Authorization: Bearer <your_token>"
```

Response:
```json
{
  "message": "Thành công",
  "data": [
    {"month": 1, "revenue": 15000000},
    {"month": 2, "revenue": 12000000},
    {"month": 3, "revenue": 18000000},
    ...
  ]
}
```

### Top khách hàng VIP

```bash
curl "http://localhost:8080/api/admin/getTopCustomers?limit=5" \
  -H "Authorization: Bearer <your_token>"
```

## 🔒 Middleware

### AuthMiddleware
Xác thực JWT token và đưa claims vào context.

### RequireRoles(roles...)
Kiểm tra vai trò người dùng, chỉ cho phép các role được chỉ định.

### SelfOrRoles(roles...)
Cho phép nếu là chính user đó hoặc có role trong danh sách.

Example:
```go
authAuth.PUT("/updateUserById/:id", 
    middleware.SelfOrRoles("quan_tri"), 
    s.UpdateUserById)
```

## 📚 Swagger Documentation

Truy cập tài liệu API đầy đủ tại:
```
http://localhost:8080/swagger/index.html
```

## 🏗️ Database Schema

Xem chi tiết schema trong `db/migration/schema.sql`

Các bảng chính:
- `nguoi_dung` - Người dùng
- `tour` - Tour du lịch
- `khoi_hanh_tour` - Lịch khởi hành
- `dat_cho_tour` - Đặt chỗ
- `thanh_toan` - Thanh toán
- `danh_gia` - Đánh giá
- `nha_cung_cap` - Nhà cung cấp
- `diem_den` - Điểm đến

## 🧪 Testing

```bash
# Run tests
go test ./...

# Run with coverage
go test -cover ./...
```

## 📦 Deployment

### Build

```bash
go build -o travia-backend main.go
```

### Run

```bash
./travia-backend
```

### Docker (Optional)

```bash
docker build -t travia-backend .
docker run -p 8080:8080 travia-backend
```

## 🛠️ Development

### Generate SQL code (sau khi thay đổi queries)

```bash
sqlc generate
```

### Update Swagger docs

```bash
swag init
```

## 📄 License

[Your License Here]

## 👥 Contributors

[Your Team]

## 📞 Support

Email: support@travia.com