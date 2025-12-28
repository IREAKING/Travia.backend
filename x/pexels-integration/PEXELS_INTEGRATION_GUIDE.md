# Pexels API Integration cho Destination Images

Tài liệu này hướng dẫn cách sử dụng Pexels API để tự động cập nhật hình ảnh cho bảng `diem_den` trong hệ thống Travia.

## Tổng quan

Hệ thống tích hợp Pexels API để:
- Tự động tìm kiếm hình ảnh chất lượng cao cho các điểm đến
- Cập nhật cột `anh` trong bảng `diem_den` với URL hình ảnh từ Pexels
- Cung cấp API endpoints để quản lý hình ảnh destinations
- Sử dụng SQLC để quản lý database queries

## Cấu hình

### 1. Environment Variables

Thêm vào file `.env`:

```bash
# Pexels API Configuration
PEXELS_API_KEY=your_pexels_api_key_here
```

### 2. Lấy Pexels API Key

1. Đăng ký tài khoản tại [Pexels](https://www.pexels.com/api/)
2. Tạo API key từ dashboard
3. Free tier cho phép 200 requests/hour

## Cấu trúc Files

```
Travia.backend/
├── config/
│   └── config.go                    # Thêm PexelsConfig
├── api/
│   ├── services/
│   │   └── pexels_service.go       # Service xử lý Pexels API
│   └── handler/
│       └── images.go               # API handlers cho image management
├── db/
│   └── query/
│       └── destination.sql         # SQLC queries cho destination images
├── scripts/
│   └── update_destination_images_sqlc.go  # Script cập nhật hình ảnh
└── cmd/
    └── update_destination_images_sqlc/
        └── main.go                # Standalone script
```

## SQLC Queries

### Các query mới được thêm:

1. **GetDestinationsNeedingImageUpdate**: Lấy destinations cần cập nhật hình ảnh
2. **UpdateDestinationImage**: Cập nhật hình ảnh cho destination
3. **GetDestinationImageStatus**: Lấy trạng thái hình ảnh của tất cả destinations
4. **CountDestinationsByImageStatus**: Đếm số lượng destinations theo trạng thái hình ảnh

### Chạy SQLC để generate code:

```bash
cd Travia.backend
sqlc generate
```

## Sử dụng

### 1. Script Command Line

#### Chạy script đơn giản:
```bash
cd Travia.backend/scripts
go run update_destination_images_sqlc.go
```

#### Chạy script với full project structure:
```bash
cd Travia.backend/cmd/update_destination_images_sqlc
go run main.go
```

### 2. API Endpoints

#### Cập nhật hình ảnh cho một destination:
```http
POST /api/destinations/images/update
Content-Type: application/json

{
    "destination_id": 1
}
```

#### Lấy danh sách destinations cần cập nhật hình ảnh:
```http
GET /api/destinations/images/needing-update
```

#### Lấy trạng thái hình ảnh của tất cả destinations:
```http
GET /api/destinations/images/status
```

#### Cập nhật hàng loạt:
```http
POST /api/destinations/images/bulk-update?ids=1,2,3,4,5
```

### 3. Sử dụng trong Code

```go
// Initialize Pexels service
pexelsService := services.NewPexelsService(cfg.PexelsConfig.APIKey)

// Search for image
imageURL, err := pexelsService.SearchImage("Hà Nội", "Hà Nội")
if err != nil {
    log.Printf("Failed to find image: %v", err)
    return
}

// Update database using SQLC
_, err = queries.UpdateDestinationImage(ctx, sqlc.UpdateDestinationImageParams{
    Anh: imageURL,
    ID:  destinationID,
})
```

## Tính năng

### 1. Pexels Service Features

- **Multiple Search Strategies**: Thử nhiều cách tìm kiếm khác nhau
- **Rate Limiting**: Tự động delay giữa các requests
- **Error Handling**: Xử lý lỗi chi tiết
- **Image Quality Selection**: Ưu tiên hình ảnh chất lượng cao
- **API Key Validation**: Kiểm tra API key hợp lệ

### 2. Search Strategies

Service sẽ thử các chiến lược tìm kiếm sau:
1. `"{destination} {province} city"`
2. `"{destination} city"`
3. `"{destination} travel"`
4. `"{destination} tourism"`
5. `"{destination}"`

### 3. Image Status Types

- **missing**: Không có hình ảnh
- **local_file**: File hình ảnh local (.jpg, .png, .jpeg)
- **external_url**: URL hình ảnh từ external source (http/https)
- **unknown**: Trạng thái không xác định

## Rate Limits

- **Free Tier**: 200 requests/hour
- **Paid Tier**: 10,000 requests/hour
- Script tự động delay 2 giây giữa các requests

## Error Handling

### Common Errors:

1. **Invalid API Key**: Kiểm tra PEXELS_API_KEY trong .env
2. **Rate Limit Exceeded**: Giảm tần suất requests
3. **No Images Found**: Thử với destination name khác
4. **Database Connection**: Kiểm tra database configuration

### Debugging:

```bash
# Enable debug logging
export DEBUG=true
go run update_destination_images_sqlc.go
```

## Monitoring

### 1. Check Image Status:
```sql
SELECT 
    CASE 
        WHEN anh IS NULL OR anh = '' THEN 'missing'
        WHEN anh LIKE '%.jpg' OR anh LIKE '%.png' THEN 'local_file'
        WHEN anh LIKE 'http%' THEN 'external_url'
        ELSE 'unknown'
    END AS image_status,
    COUNT(*) AS total
FROM diem_den 
GROUP BY image_status;
```

### 2. Check Recent Updates:
```sql
SELECT ten, anh, ngay_cap_nhat 
FROM diem_den 
WHERE ngay_cap_nhat > NOW() - INTERVAL '1 hour'
ORDER BY ngay_cap_nhat DESC;
```

## Best Practices

1. **Chạy script trong giờ thấp điểm** để tránh rate limit
2. **Backup database** trước khi chạy bulk update
3. **Monitor API usage** để tránh vượt quá giới hạn
4. **Test với một vài destinations** trước khi chạy toàn bộ
5. **Sử dụng transaction** cho bulk operations

## Troubleshooting

### 1. Script không chạy được:
- Kiểm tra database connection
- Verify PEXELS_API_KEY
- Check Go modules: `go mod tidy`

### 2. Không tìm thấy hình ảnh:
- Thử với tên destination bằng tiếng Anh
- Kiểm tra spelling của destination name
- Sử dụng manual search trên Pexels website

### 3. Database errors:
- Kiểm tra SQLC generated code: `sqlc generate`
- Verify database schema matches
- Check foreign key constraints

## Future Enhancements

1. **Caching**: Cache hình ảnh để giảm API calls
2. **Image Optimization**: Resize/compress hình ảnh
3. **Fallback Sources**: Sử dụng multiple image sources
4. **Batch Processing**: Xử lý hàng loạt với progress tracking
5. **Admin Dashboard**: UI để quản lý hình ảnh destinations
