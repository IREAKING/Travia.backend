# Travia Backend

Backend API cho hệ thống đặt tour du lịch Travia, xây dựng bằng Go + Gin, sử dụng PostgreSQL và Redis.

## Tech stack
- Go 1.25
- Gin Web Framework
- PostgreSQL (pgx)
- Redis (cache + rate limit)
- sqlc (type-safe SQL)
- Swagger/OpenAPI

## Yêu cầu
- Go 1.25+
- PostgreSQL 14+
- Redis 7+
- sqlc CLI

## Cấu trúc chính
```
Travia.backend/
├── api/               # Handlers, middleware, models, utils
├── db/
│   ├── migration/     # schema.sql
│   ├── query/         # SQL dùng cho sqlc
│   └── sqlc/          # Code Go sinh ra
├── docs/              # swagger.json/yaml
├── config/            # cấu hình
└── main.go
```

## Cấu hình môi trường
Tạo file `env/.env`:
```
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=travia
DB_SSLMODE=disable

# Server
PORT_NUMBER=8080
HOST=0.0.0.0
ENVIRONMENT=development
SECRET_KEY=your-secret-key
API_SECRET=your-api-secret

# Redis
REDIS_ADDRESS=localhost:6379
REDIS_DB=0
REDIS_PASSWORD=

# Email (OTP)
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

## Cài đặt & chạy
```
cd Travia.backend
go mod download
sqlc generate
psql "$DATABASE_URL" -f db/migration/schema.sql
go run main.go
```

## Swagger
Sau khi chạy server:
```
http://localhost:8080/swagger/index.html
```

## Quy ước xác thực
Gửi JWT trong header:
```
Authorization: Bearer <token>
```

## Các lệnh hữu ích
```
# Sinh lại sqlc sau khi sửa query
sqlc generate

# Tạo swagger (nếu có cập nhật comment)
swag init

# Chạy test
go test ./...
```

## Ghi chú
- Dữ liệu tour chi tiết lấy từ `GET /tour/:id`.
- Cache có thể bật/tắt trong `api/handler/router.go`.
