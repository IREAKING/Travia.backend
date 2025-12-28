# Hướng dẫn Implementation - Tạo Tour với Transaction

## 📋 Tổng quan

Implementation này cho phép tạo tour với **TẤT CẢ dữ liệu liên quan trong 1 transaction duy nhất**, đảm bảo:
- ✅ **Data consistency** (all or nothing)
- ✅ **Rollback tự động** nếu có lỗi
- ✅ **Performance tốt** (1 API call thay vì nhiều calls)
- ✅ **Dễ maintain** và debug

## 🎯 Các bước thực hiện

### Bước 1: Generate SQL code với sqlc

Trước tiên, bạn cần generate Go code từ SQL queries mới:

```bash
# Từ thư mục root của project
cd /Users/macbook-pro/Library/CloudStorage/OneDrive-Personal/VirtualBox/Travia/Travia.backend

# Generate sqlc code
sqlc generate
```

**Lưu ý:** Nếu gặp lỗi, kiểm tra:
- File `sqlc.yaml` có đúng config không
- Các file `.sql` trong `db/query/` có syntax đúng không

### Bước 2: Kiểm tra code đã generate

Sau khi chạy `sqlc generate`, kiểm tra các file mới được tạo:

```bash
# Kiểm tra file itinerary.sql.go đã được generate
ls -la db/sqlc/itinerary.sql.go

# Xem nội dung (nếu cần)
head -n 50 db/sqlc/itinerary.sql.go
```

Bạn sẽ thấy các functions:
- `CreateItinerary`
- `CreateActivity`
- `CreateGroupConfig`
- `GetItinerariesByTour`
- v.v.

### Bước 3: Compile code

```bash
# Build toàn bộ project
go build -o tmp/travia-backend main.go

# Hoặc chỉ test compile
go build ./...
```

**Nếu gặp lỗi compile:**

```bash
# Check linter errors
golangci-lint run

# Hoặc chỉ check syntax
go vet ./...
```

### Bước 4: Fix các lỗi còn thiếu

Có thể bạn cần thêm một số helper queries chưa có:

#### A. CountBookingsByTour (nếu chưa có)

Thêm vào `db/query/booking.sql`:

```sql
-- name: CountBookingsByTour :one
SELECT COUNT(*) 
FROM dat_cho_tour dct
JOIN khoi_hanh_tour kht ON dct.khoi_hanh_id = kht.id
WHERE kht.tour_id = $1;
```

#### B. Missing Params types

Nếu sqlc báo lỗi về params, bạn có thể cần adjust queries trong `db/query/itinerary.sql`:

Ví dụ, thay vì:
```sql
INSERT INTO lich_trinh_tour (tour_id, ngay_thu, ...) VALUES ($1, $2, ...)
```

Có thể cần chỉ định rõ types:
```sql
INSERT INTO lich_trinh_tour (
    tour_id,           -- INTEGER
    ngay_thu,          -- INTEGER
    tieu_de,           -- VARCHAR
    mo_ta,             -- TEXT
    gio_bat_dau,       -- TIME
    gio_ket_thuc,      -- TIME
    dia_diem,          -- TEXT
    thong_tin_luu_tru  -- TEXT
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;
```

### Bước 5: Test manual với database

Trước khi test qua API, test trực tiếp với database:

```bash
# Connect vào PostgreSQL
psql -U your_user -d travia_db

# Test tạo itinerary (sau khi đã có tour_id)
INSERT INTO lich_trinh_tour (tour_id, ngay_thu, tieu_de, mo_ta) 
VALUES (1, 1, 'Ngày 1', 'Mô tả ngày 1') 
RETURNING *;

# Test tạo activity
INSERT INTO hoat_dong_lich_trinh (lich_trinh_id, ten, mo_ta, thu_tu) 
VALUES (1, 'Hoạt động 1', 'Mô tả', 1) 
RETURNING *;
```

### Bước 6: Chạy server

```bash
# Chạy server (development mode)
go run main.go

# Hoặc dùng air để auto-reload
air

# Hoặc dùng binary đã build
./tmp/travia-backend
```

Server sẽ chạy trên port mặc định (thường là `:8080`)

### Bước 7: Test API với HTTP client

#### Sử dụng REST Client (VSCode Extension)

1. Cài đặt extension "REST Client" trong VSCode
2. Mở file `x/tour-creation/test_create_tour.http`
3. Cập nhật `@authToken` với JWT token hợp lệ:

```http
### Login để lấy token
POST http://localhost:8080/api/auth/login
Content-Type: application/json

{
  "email": "admin@example.com",
  "mat_khau": "your_password"
}
```

4. Copy access token từ response
5. Update biến `@authToken` trong file `.http`:

```http
@authToken = Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

6. Click "Send Request" trên endpoint `/api/tour/create-full`

#### Sử dụng curl

```bash
# Login để lấy token
TOKEN=$(curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","mat_khau":"password"}' \
  | jq -r '.access_token')

# Tạo tour với full details
curl -X POST http://localhost:8080/api/tour/create-full \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d @x/tour-creation/example_create_tour_request.json \
  | jq '.'
```

#### Sử dụng Postman

1. Import collection:
   - Method: POST
   - URL: `http://localhost:8080/api/tour/create-full`
   - Headers: 
     - `Content-Type: application/json`
     - `Authorization: Bearer YOUR_TOKEN`
   - Body: Copy từ `example_create_tour_request.json`

### Bước 8: Verify kết quả

Sau khi tạo tour thành công, verify trong database:

```sql
-- Check tour
SELECT * FROM tour ORDER BY id DESC LIMIT 1;

-- Check images
SELECT * FROM anh_tour WHERE tour_id = (SELECT MAX(id) FROM tour);

-- Check destinations
SELECT * FROM tour_diem_den WHERE tour_id = (SELECT MAX(id) FROM tour);

-- Check itineraries
SELECT * FROM lich_trinh_tour WHERE tour_id = (SELECT MAX(id) FROM tour);

-- Check activities
SELECT hd.* 
FROM hoat_dong_lich_trinh hd
JOIN lich_trinh_tour lt ON hd.lich_trinh_id = lt.id
WHERE lt.tour_id = (SELECT MAX(id) FROM tour);

-- Check group config
SELECT * FROM cau_hinh_nhom_tour WHERE tour_id = (SELECT MAX(id) FROM tour);
```

## 🐛 Troubleshooting

### Lỗi: "cannot find function CreateItinerary"

**Nguyên nhân:** sqlc chưa generate code

**Giải pháp:**
```bash
sqlc generate
go build ./...
```

### Lỗi: "WithTx is not defined"

**Nguyên nhân:** File `db.go` chưa có method `WithTx`

**Giải pháp:** Method này đã có sẵn trong code generated. Nếu không, thêm vào `db/sqlc/db.go`:

```go
func (q *Queries) WithTx(tx pgx.Tx) *Queries {
    return &Queries{db: tx}
}
```

### Lỗi: "transaction is already closed"

**Nguyên nhân:** Transaction bị commit hoặc rollback nhiều lần

**Giải pháp:** Kiểm tra logic trong `tour_tx.go`, đảm bảo:
- `defer tx.Rollback()` chỉ chạy khi có lỗi
- `tx.Commit()` chỉ chạy 1 lần ở cuối

### Lỗi: "failed to convert time"

**Nguyên nhân:** Format thời gian không đúng

**Giải pháp:** Đảm bảo format là `HH:MM:SS`:
```json
{
  "gio_bat_dau": "08:00:00",
  "gio_ket_thuc": "18:00:00"
}
```

### Lỗi: "foreign key violation"

**Nguyên nhân:** 
- `diem_den_id` không tồn tại
- `danh_muc_id` không tồn tại
- `nha_cung_cap_id` không tồn tại

**Giải pháp:** Kiểm tra và tạo data cần thiết:

```sql
-- Check destinations exist
SELECT id, ten FROM diem_den;

-- Check categories exist
SELECT id, ten FROM danh_muc_tour;

-- Check suppliers exist
SELECT id, ten FROM nha_cung_cap;
```

## ✅ Checklist trước khi deploy

- [ ] Đã chạy `sqlc generate` thành công
- [ ] Code compile không có lỗi
- [ ] Test API thành công với dữ liệu mẫu
- [ ] Verify dữ liệu trong database đúng
- [ ] Test rollback khi có lỗi (ví dụ: invalid destination_id)
- [ ] Test với nhiều scenarios:
  - [ ] Tour với 1 ảnh
  - [ ] Tour với nhiều ảnh
  - [ ] Tour với nhiều ngày lịch trình
  - [ ] Tour với nhiều hoạt động mỗi ngày
  - [ ] Tour không có group config
  - [ ] Tour có group config
- [ ] Update Swagger docs (nếu dùng)
- [ ] Thêm logging phù hợp
- [ ] Thêm monitoring/metrics

## 📊 Performance tips

### 1. Index optimization

Đảm bảo có indexes trên các foreign keys (đã có trong schema):
```sql
CREATE INDEX idx_lich_trinh_tour_tour_id ON lich_trinh_tour(tour_id);
CREATE INDEX idx_hoat_dong_lich_trinh_lich_trinh_id ON hoat_dong_lich_trinh(lich_trinh_id);
```

### 2. Batch insert optimization

Nếu cần tạo nhiều tours, có thể implement batch:

```go
func (t *Travia) CreateMultipleTours(ctx context.Context, toursList []CreateTourWithDetailsParams) error {
    for _, params := range toursList {
        _, err := t.CreateTourWithDetails(ctx, params)
        if err != nil {
            return err
        }
    }
    return nil
}
```

### 3. Connection pooling

Đảm bảo pgxpool được config đúng trong `config.go`:

```go
config, _ := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
config.MaxConns = 25
config.MinConns = 5
config.MaxConnLifetime = time.Hour
config.MaxConnIdleTime = 30 * time.Minute
```

## 🚀 Next Steps

Sau khi implementation này hoàn thành:

1. **Implement Update Tour with Transaction**
   - Update tour basic info
   - Smart diff cho images/destinations/itineraries
   - Soft delete cho các items không còn

2. **Implement Soft Delete**
   - Đánh dấu `dang_hoat_dong = false` thay vì DELETE
   - Có thể restore sau này

3. **Add Validation Layer**
   - Validate business rules trước khi insert
   - Check conflicts (duplicate itinerary days, etc.)

4. **Add Audit Logging**
   - Log mọi thay đổi vào bảng audit
   - Track ai tạo/sửa/xóa tour

5. **Implement Caching Strategy**
   - Cache tour details sau khi tạo
   - Invalidate cache khi update

## 📚 Tài liệu tham khảo

- [PostgreSQL Transactions](https://www.postgresql.org/docs/current/tutorial-transactions.html)
- [pgx Documentation](https://pkg.go.dev/github.com/jackc/pgx/v5)
- [sqlc Documentation](https://docs.sqlc.dev/)
- [Gin Web Framework](https://gin-gonic.com/docs/)

