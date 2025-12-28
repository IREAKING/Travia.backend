# Dữ Liệu Mẫu Cho Hệ Thống Travia

Thư mục này chứa dữ liệu mẫu cho hệ thống quản lý tour du lịch Travia.

## 📋 Danh Sách File Dữ Liệu

### 1. File Chính
- `00_insert_all_data.sql` - **Master file** chạy tất cả file dữ liệu theo thứ tự

### 2. File Dữ Liệu Chi Tiết

| STT | File | Mô Tả | Phụ Thuộc |
|-----|------|-------|-----------|
| 1 | `nguoidung.sql` | Người dùng (admin, nhà cung cấp, khách hàng) | - |
| 2 | `danhmuctour.sql` | Danh mục tour (18 loại) | - |
| 3 | `diemden.sql` | Điểm đến (25 địa điểm) | - |
| 4 | `nhacungcap.sql` | Nhà cung cấp tour (5 công ty) | `nguoidung` |
| 5 | `tour.sql` | Thông tin tour (21 tour) | `nguoidung`, `nhacungcap`, `danhmuctour` |
| 6 | `anhtour.sql` | Ảnh của các tour | `tour` |
| 7 | `tourdiemden.sql` | Liên kết tour với điểm đến | `tour`, `diemden` |
| 8 | `lichtrinhtour.sql` | Lịch trình chi tiết (4 tour mẫu) | `tour` |
| 9 | `hoatdonglichTrinh.sql` | Hoạt động trong lịch trình | `lichtrinhtour` |
| 10 | `cauhinhnhomtour.sql` | Cấu hình số lượng khách | `tour` |
| 11 | `khoihanhtour.sql` | Lịch khởi hành (50+ lịch) | `tour` |
| 12 | `giamgiatour.sql` | Chương trình giảm giá | `tour` |

## 🚀 Cách Sử Dụng

### Cách 1: Chạy Tất Cả (Khuyến Nghị)

```bash
# Di chuyển đến thư mục data
cd Travia.backend/db/data

# Chạy master file
psql -U postgres -d travia_db -f 00_insert_all_data.sql
```

### Cách 2: Chạy Từng File Riêng Lẻ

```bash
# 1. Người dùng
psql -U postgres -d travia_db -f nguoidung.sql

# 2. Danh mục tour
psql -U postgres -d travia_db -f danhmuctour.sql

# 3. Điểm đến
psql -U postgres -d travia_db -f diemden.sql

# 4. Nhà cung cấp
psql -U postgres -d travia_db -f nhacungcap.sql

# 5. Tour
psql -U postgres -d travia_db -f tour.sql

# 6-12. Các file còn lại...
```

### Cách 3: Trong PostgreSQL Shell

```sql
-- Kết nối database
\c travia_db

-- Chạy file
\i /path/to/00_insert_all_data.sql
```

## 📊 Thống Kê Dữ Liệu

### Người Dùng
- **Admin**: 1 tài khoản
- **Nhà cung cấp**: 5 tài khoản
- **Khách hàng**: 3 tài khoản
- **Mật khẩu mặc định**: `Password123!` (đã hash bcrypt)

### Tour
- **Tổng số**: 21 tour
- **Trạng thái công bố**: 19 tour
- **Trạng thái nháp**: 1 tour (Châu Âu)
- **Trạng thái lưu trữ**: 1 tour
- **Tour nổi bật**: 10 tour

### Phân Loại Tour
- **Tour nội địa**: 9 tour (Hạ Long, Đà Nẵng, Phú Quốc, Nha Trang, Đà Lạt, Cần Thơ, Ninh Bình, Côn Đảo, Quy Nhơn)
- **Tour quốc tế**: 6 tour (Thái Lan, Singapore, Bali, Hàn Quốc, Nhật Bản)
- **Tour đặc biệt**: 4 tour (Team Building, Trăng mật, Gia đình, Học sinh)

### Lịch Trình Chi Tiết
Có lịch trình chi tiết cho 4 tour mẫu:
1. **Tour 1**: Hạ Long - Sapa (5 ngày, 20+ hoạt động)
2. **Tour 2**: Đà Nẵng - Hội An - Huế (4 ngày, 15+ hoạt động)
3. **Tour 3**: Phú Quốc (4 ngày, 12+ hoạt động)
4. **Tour 10**: Bangkok - Pattaya (5 ngày, 18+ hoạt động)

### Lịch Khởi Hành
- **Tổng số**: 50+ lịch khởi hành
- **Đã xác nhận**: ~20 lịch
- **Lên lịch**: ~25 lịch
- **Hoàn thành**: 3 lịch
- **Đã hủy**: 2 lịch

### Giảm Giá
- **Black Friday**: 5 tour giảm 10-20%
- **Giáng Sinh**: 3 tour giảm 10-12%
- **Tết Dương lịch**: 2 tour giảm 8-10%
- **Khuyến mãi đặc biệt**: Tour trăng mật giảm 20%

## 🗂️ Cấu Trúc Dữ Liệu

### Người Dùng Mẫu

| Email | Vai Trò | Mật Khẩu |
|-------|---------|----------|
| admin@travia.vn | Quản trị | Password123! |
| minh@saigontourist.net | Nhà cung cấp | Password123! |
| thu@vietravel.com | Nhà cung cấp | Password123! |
| lan.vo@gmail.com | Khách hàng | Password123! |

### Danh Mục Tour (18 loại)

1. Tour nội địa
2. Tour quốc tế
3. Tour inbound
4. Du lịch nghỉ dưỡng
5. Du lịch khám phá - mạo hiểm
6. Du lịch sinh thái
7. Du lịch văn hóa - lịch sử
8. Du lịch tâm linh
9. Du lịch ẩm thực
10. Du lịch MICE
11. Du lịch học tập - trải nghiệm
12. Du lịch chăm sóc sức khỏe
13. Du lịch cộng đồng
14. Tour gia đình
15. Tour trăng mật
16. Tour đoàn thể - team building
17. Tour học sinh - sinh viên
18. Tour cao cấp

### Điểm Đến (25 địa điểm)

**Việt Nam** (15 điểm):
- Miền Bắc: Hà Nội, Hạ Long, Sapa, Ninh Bình
- Miền Trung: Đà Nẵng, Hội An, Huế, Phong Nha, Quy Nhơn
- Miền Nam: TP.HCM, Đà Lạt, Nha Trang, Phú Quốc, Cần Thơ, Mũi Né

**Quốc Tế** (10 điểm):
- Đông Nam Á: Bangkok, Phuket, Singapore, Bali
- Đông Á: Tokyo, Kyoto, Seoul, Jeju
- Châu Âu: Paris, Rome

## 📝 Lưu Ý Quan Trọng

### Trước Khi Chạy
1. ✅ Đảm bảo đã chạy `schema.sql` để tạo các bảng
2. ✅ Database đã được tạo: `travia_db`
3. ✅ Có quyền INSERT vào các bảng

### Thứ Tự Phụ Thuộc
⚠️ **Phải tuân thủ thứ tự sau:**
```
nguoidung → nha_cung_cap
nguoidung, danhmuctour, nha_cung_cap → tour
tour → anh_tour, tourdiemden, lichtrinhtour, cauhinhnhomtour, khoihanhtour, giamgiatour
lichtrinhtour → hoatdonglichTrinh
```

### Xóa Dữ Liệu Cũ
Nếu muốn xóa dữ liệu và insert lại:

```sql
-- ⚠️ CẢNH BÁO: Lệnh này sẽ XÓA TẤT CẢ dữ liệu!
TRUNCATE TABLE 
    hoat_dong_trong_ngay,
    lich_trinh,
    giam_gia_tour,
    khoi_hanh_tour,
    cau_hinh_nhom_tour,
    tour_diem_den,
    anh_tour,
    tour,
    nha_cung_cap,
    diem_den,
    danh_muc_tour,
    nguoi_dung
CASCADE;
```

## 🔍 Kiểm Tra Dữ Liệu

### Kiểm Tra Số Lượng
```sql
-- Xem tổng số tour
SELECT COUNT(*) FROM tour;

-- Xem tour theo trạng thái
SELECT trang_thai, COUNT(*) 
FROM tour 
GROUP BY trang_thai;

-- Xem tour nổi bật
SELECT tieu_de, gia_nguoi_lon, trang_thai
FROM tour 
WHERE noi_bat = TRUE AND dang_hoat_dong = TRUE;
```

### Kiểm Tra Lịch Khởi Hành
```sql
-- Xem lịch khởi hành sắp tới
SELECT t.tieu_de, k.ngay_khoi_hanh, k.trang_thai, k.suc_chua
FROM khoi_hanh_tour k
JOIN tour t ON k.tour_id = t.id
WHERE k.ngay_khoi_hanh >= CURRENT_DATE
ORDER BY k.ngay_khoi_hanh;
```

### Kiểm Tra Giảm Giá
```sql
-- Xem các tour đang giảm giá
SELECT t.tieu_de, g.phan_tram, g.ngay_bat_dau, g.ngay_ket_thuc
FROM giam_gia_tour g
JOIN tour t ON g.tour_id = t.id
WHERE CURRENT_DATE BETWEEN g.ngay_bat_dau AND g.ngay_ket_thuc
ORDER BY g.phan_tram DESC;
```

## 🛠️ Troubleshooting

### Lỗi: "relation does not exist"
➡️ Chưa chạy `schema.sql`. Chạy lệnh:
```bash
psql -U postgres -d travia_db -f ../migration/schema.sql
```

### Lỗi: "duplicate key value"
➡️ Dữ liệu đã tồn tại. Xóa và chạy lại hoặc sửa UUID/ID.

### Lỗi: "foreign key constraint"
➡️ Sai thứ tự chạy file. Phải chạy theo thứ tự phụ thuộc.

### Lỗi: "permission denied"
➡️ Không có quyền. Chạy với user có quyền:
```bash
psql -U postgres -d travia_db
```

## 📞 Hỗ Trợ

Nếu gặp vấn đề, kiểm tra:
1. Log file PostgreSQL
2. Connection string
3. User permissions
4. Database existence

---

**Phiên bản**: 1.0  
**Ngày cập nhật**: 2025-11-15  
**Tác giả**: Travia Development Team

