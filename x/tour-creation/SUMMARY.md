# Tóm tắt: Tạo Tour với nhiều bảng liên quan

## 🎯 Câu trả lời ngắn gọn

**KHÔNG THỂ** viết 1 câu lệnh SQL để INSERT vào nhiều bảng khác nhau.

**PHẢI** dùng nhiều câu INSERT riêng biệt, **BỌC TRONG 1 TRANSACTION**.

## 📊 So sánh các phương pháp

| Tiêu chí | Transaction trong Go | Stored Procedure | CTE với RETURNING | Nhiều API calls riêng |
|----------|---------------------|------------------|-------------------|----------------------|
| **Tính toàn vẹn dữ liệu** | ✅ Excellent | ✅ Excellent | ✅ Good | ❌ Poor |
| **Dễ maintain** | ✅ Excellent | ❌ Poor | ⚠️ Medium | ⚠️ Medium |
| **Performance** | ✅ Excellent | ✅ Excellent | ✅ Good | ❌ Poor (network) |
| **Dễ debug** | ✅ Excellent | ❌ Poor | ⚠️ Medium | ✅ Good |
| **Flexibility** | ✅ Excellent | ⚠️ Medium | ❌ Poor | ✅ Excellent |
| **Type Safety** | ✅ Yes (Go + sqlc) | ❌ No | ⚠️ Limited | ✅ Yes |
| **Rollback tự động** | ✅ Yes | ✅ Yes | ✅ Yes | ❌ No |
| **Khuyến nghị** | ✅ **RECOMMENDED** | ⚠️ Backup option | ❌ Not suitable | ❌ NOT RECOMMENDED |

## ✅ Phương pháp được KHUYẾN NGHỊ: Transaction trong Go

### Ưu điểm:
- ✅ **Tách biệt logic rõ ràng** - Mỗi bước INSERT có ý nghĩa riêng
- ✅ **Tận dụng sqlc** - Sử dụng queries đã generate
- ✅ **Type-safe** - Go compiler kiểm tra types
- ✅ **Dễ test** - Mock từng bước
- ✅ **Rollback tự động** - `defer tx.Rollback()` nếu có lỗi
- ✅ **Business logic linh hoạt** - Validate, transform dữ liệu
- ✅ **Dễ version control** - Code rõ ràng, dễ review

### Cấu trúc:

```go
func CreateTourWithDetails(ctx context.Context, params CreateTourWithDetailsParams) (*Result, error) {
    // 1. Begin Transaction
    tx, err := db.Begin(ctx)
    defer tx.Rollback(ctx)
    
    // 2. Create Tour
    tour := CreateTour(ctx, tourParams)
    
    // 3. Add Images
    for img := range params.Images {
        AddTourImage(ctx, img)
    }
    
    // 4. Add Destinations
    for dest := range params.Destinations {
        AddTourDestination(ctx, dest)
    }
    
    // 5. Create Itineraries
    for itin := range params.Itineraries {
        lichTrinh := CreateItinerary(ctx, itin)
        
        // 6. Add Activities for each Itinerary
        for act := range itin.Activities {
            CreateActivity(ctx, act)
        }
    }
    
    // 7. Commit
    tx.Commit(ctx)
    
    return result, nil
}
```

## 🔴 Phương pháp KHÔNG NÊN dùng

### ❌ Nhiều API calls riêng lẻ (WITHOUT Transaction)

```
POST /api/tour/create          → Tour created (ID: 123)
POST /api/tour/123/images      → Images added
POST /api/tour/123/destinations → ERROR! Network timeout
→ Kết quả: Tour có nhưng THIẾU destinations = INCONSISTENT DATA
```

**Vấn đề:**
- ❌ Không đảm bảo tính toàn vẹn
- ❌ Khó rollback
- ❌ Network overhead cao
- ❌ User experience kém (phải đợi nhiều requests)

## 📝 Các bước thực hiện

### Bước 1: Thêm SQL queries còn thiếu
```bash
# File đã tạo: db/query/itinerary.sql
# Chứa: CreateItinerary, CreateActivity, CreateGroupConfig, etc.
```

### Bước 2: Generate sqlc code
```bash
sqlc generate
```

### Bước 3: Tạo transaction handler
```bash
# File: db/sqlc/tour_tx.go
# Chứa: CreateTourWithDetails function
```

### Bước 4: Cập nhật interface Z
```go
// db/sqlc/travia.go
type Z interface {
    Querier
    CreateTourWithDetails(ctx, params) (*Result, error)
}
```

### Bước 5: Thêm method WithTx
```go
// db/sqlc/db.go
func (q *Queries) WithTx(tx pgx.Tx) *Queries {
    return &Queries{db: tx}
}
```

### Bước 6: Tạo handler API
```go
// api/handler/tour.go
func (s *Server) CreateTourFull(c *gin.Context) {
    // Parse request
    // Call s.z.CreateTourWithDetails()
    // Return response
}
```

### Bước 7: Thêm route
```go
// api/handler/router.go
tour.POST("/create-full", middleware.AuthMiddleware(), s.CreateTourFull)
```

### Bước 8: Test
```bash
# Sử dụng file: x/tour-creation/test_create_tour.http
```

## 🎯 Kết luận

**Dùng Transaction trong Go code** để tạo tour với nhiều bảng liên quan vì:

1. ✅ **An toàn** - Đảm bảo data consistency
2. ✅ **Hiệu quả** - Chỉ 1 API call
3. ✅ **Dễ maintain** - Code rõ ràng, dễ đọc
4. ✅ **Linh hoạt** - Dễ mở rộng, thêm logic
5. ✅ **Best practice** - Đúng chuẩn industry standard

## 📂 Files đã tạo

1. ✅ `x/tour-creation/TOUR_CREATION_GUIDE.md` - Hướng dẫn chi tiết
2. ✅ `db/query/itinerary.sql` - SQL queries cho itinerary & activities
3. ✅ `x/tour-creation/example_create_tour_request.json` - Ví dụ request đầy đủ
4. ✅ `x/tour-creation/test_create_tour.http` - HTTP requests để test

## 🚀 Next Steps

1. Chạy `sqlc generate` để generate Go code
2. Implement `db/sqlc/tour_tx.go`
3. Implement handler `api/handler/tour.go`
4. Test với request mẫu
5. Deploy và monitor

