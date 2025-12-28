# 💰 Hệ Thống Chi Trả Nhà Cung Cấp - Phiên Bản Đơn Giản

## 📋 Tổng Quan

Khi khách hàng thanh toán → Tiền về Admin → Admin chi trả cho Nhà cung cấp (trừ hoa hồng)

```
Khách trả 10,000,000 VND
    ↓
Platform giữ 1,500,000 VND (15%)
    ↓  
NCC nhận 8,500,000 VND (85%)
```

---

## 🗄️ Cấu Trúc Database (Đã Đơn Giản Hóa)

### 1. **Bảng `tai_khoan_ngan_hang_ncc`** - Tài Khoản Ngân Hàng NCC

```sql
CREATE TABLE tai_khoan_ngan_hang_ncc (
    id UUID PRIMARY KEY,
    nha_cung_cap_id UUID,                   -- ID nhà cung cấp
    ten_tai_khoan VARCHAR(255),              -- Tên chủ TK
    so_tai_khoan VARCHAR(50),                -- Số TK
    ten_ngan_hang VARCHAR(255),              -- VD: Vietcombank
    chi_nhanh VARCHAR(255),                  -- Chi nhánh
    la_tai_khoan_mac_dinh BOOLEAN,           -- TK mặc định
    da_xac_minh BOOLEAN,                     -- Admin đã xác minh
    ghi_chu TEXT
);
```

**Ví dụ:**
```sql
INSERT INTO tai_khoan_ngan_hang_ncc (
    nha_cung_cap_id, ten_tai_khoan, so_tai_khoan, 
    ten_ngan_hang, la_tai_khoan_mac_dinh
) VALUES (
    'uuid-ncc', 'CÔNG TY ABC TRAVEL', '0123456789',
    'Vietcombank', TRUE
);
```

### 2. **Bảng `cau_hinh_hoa_hong`** - Cấu Hình % Hoa Hồng

```sql
CREATE TABLE cau_hinh_hoa_hong (
    id SERIAL PRIMARY KEY,
    ti_le_hoa_hong_platform DECIMAL(5,2) DEFAULT 15.00  -- 15%
);
```

**Mặc định:** Platform nhận 15%, NCC nhận 85%

**Tùy chỉnh cho từng NCC** (optional):
```sql
-- Trong bảng nha_cung_cap
UPDATE nha_cung_cap 
SET ti_le_hoa_hong_platform = 10.00  -- NCC VIP chỉ mất 10%
WHERE id = 'uuid-ncc-vip';
```

### 3. **Bảng `chi_tra_nha_cung_cap`** - Chi Trả

```sql
CREATE TABLE chi_tra_nha_cung_cap (
    id UUID PRIMARY KEY,
    nha_cung_cap_id UUID,
    thanh_toan_id UUID,                      -- Link đến thanh toán của khách
    
    tong_tien_khach_tra DECIMAL(12,2),       -- 100% tiền khách trả
    ti_le_hoa_hong_platform DECIMAL(5,2),    -- % hoa hồng (15%)
    tien_cho_nha_cung_cap DECIMAL(12,2),     -- Tiền NCC nhận (85%)
    
    trang_thai ENUM (                        -- Trạng thái
        'cho_chi_tra',                       --   Chờ chi trả
        'san_sang',                          --   Sẵn sàng chi trả
        'dang_xu_ly',                        --   Đang xử lý
        'da_chi_tra',                        --   Đã chi trả
        'that_bai',                          --   Thất bại
        'da_giu',                            --   Bị giữ (tranh chấp)
        'da_huy'                             --   Đã hủy
    ),
    
    tai_khoan_ngan_hang_id UUID,            -- TK ngân hàng nhận tiền
    ma_giao_dich VARCHAR(100),               -- Mã GD ngân hàng
    ngay_chi_tra TIMESTAMP,                  -- Ngày chi trả
    nguoi_duyet_id UUID,                     -- Admin duyệt
    ghi_chu TEXT
);
```

---

## 🔄 Luồng Hoạt Động

### **Bước 1: Khách thanh toán**

```sql
-- Khách thanh toán 10,000,000 VND qua Stripe
INSERT INTO thanh_toan (...) VALUES (...);
UPDATE thanh_toan SET trang_thai = 'thanh_cong' WHERE id = '...';
```

### **Bước 2: Hệ thống TỰ ĐỘNG tạo chi trả** ⚡

Trigger tự động chạy:
```sql
-- Tự động insert vào chi_tra_nha_cung_cap
{
    tong_tien_khach_tra: 10,000,000 VND
    ti_le_hoa_hong_platform: 15%
    tien_cho_nha_cung_cap: 8,500,000 VND
    trang_thai: 'cho_chi_tra'
}
```

### **Bước 3: Admin chi trả**

```sql
-- Xem danh sách cần chi trả
SELECT * FROM v_chi_tra_ncc 
WHERE trang_thai = 'cho_chi_tra';

-- Cập nhật trạng thái sẵn sàng
UPDATE chi_tra_nha_cung_cap
SET trang_thai = 'san_sang'
WHERE id = '...';

-- Chuyển khoản thủ công → Cập nhật đã chi trả
UPDATE chi_tra_nha_cung_cap
SET 
    trang_thai = 'da_chi_tra',
    ma_giao_dich = 'BANK_REF_123',
    ngay_chi_tra = NOW(),
    nguoi_duyet_id = 'uuid-admin'
WHERE id = '...';
```

---

## 📊 VIEWs Hữu Ích

### 1. Danh sách chi trả

```sql
SELECT * FROM v_chi_tra_ncc 
ORDER BY ngay_tao DESC;
```

| Tên NCC      | Tổng tiền     | % HH | Tiền cho NCC | Trạng thái    |
|-------------|--------------|------|-------------|--------------|
| ABC Travel  | 10,000,000   | 15%  | 8,500,000   | cho_chi_tra  |
| XYZ Tours   | 5,000,000    | 15%  | 4,250,000   | da_chi_tra   |

### 2. NCC chưa có tài khoản

```sql
SELECT * FROM v_ncc_chua_co_tai_khoan_ngan_hang;
```

---

## 💼 Các Thao Tác Thường Dùng

### 1. Thêm tài khoản ngân hàng cho NCC

```sql
INSERT INTO tai_khoan_ngan_hang_ncc (
    nha_cung_cap_id,
    ten_tai_khoan,
    so_tai_khoan,
    ten_ngan_hang,
    chi_nhanh,
    la_tai_khoan_mac_dinh,
    da_xac_minh
) VALUES (
    'uuid-ncc',
    'CÔNG TY ABC TRAVEL',
    '0123456789',
    'Vietcombank',
    'Chi nhánh Hà Nội',
    TRUE,
    FALSE  -- Admin sẽ xác minh sau
);
```

### 2. Admin xác minh tài khoản

```sql
UPDATE tai_khoan_ngan_hang_ncc
SET da_xac_minh = TRUE
WHERE id = 'uuid-tai-khoan';
```

### 3. Thay đổi % hoa hồng cho NCC cụ thể

```sql
-- NCC VIP: Giảm hoa hồng xuống 10%
UPDATE nha_cung_cap
SET ti_le_hoa_hong_platform = 10.00
WHERE id = 'uuid-ncc-vip';
```

### 4. Xem báo cáo chi trả

```sql
-- Tổng tiền cần chi trả
SELECT 
    SUM(tien_cho_nha_cung_cap) as tong_tien,
    COUNT(*) as so_luong
FROM chi_tra_nha_cung_cap
WHERE trang_thai IN ('cho_chi_tra', 'san_sang');

-- Chi tiết theo NCC
SELECT 
    ncc.ten,
    COUNT(ct.id) as so_giao_dich,
    SUM(ct.tien_cho_nha_cung_cap) as tong_tien
FROM chi_tra_nha_cung_cap ct
JOIN nha_cung_cap ncc ON ct.nha_cung_cap_id = ncc.id
WHERE ct.trang_thai = 'cho_chi_tra'
GROUP BY ncc.ten
ORDER BY tong_tien DESC;
```

### 5. Chi trả hàng loạt

```sql
-- Đánh dấu tất cả sẵn sàng chi trả
UPDATE chi_tra_nha_cung_cap
SET trang_thai = 'san_sang'
WHERE trang_thai = 'cho_chi_tra'
    AND ngay_tao < NOW() - INTERVAL '7 days';

-- Sau khi chuyển khoản, cập nhật hàng loạt
UPDATE chi_tra_nha_cung_cap
SET 
    trang_thai = 'da_chi_tra',
    ngay_chi_tra = NOW(),
    nguoi_duyet_id = 'uuid-admin'
WHERE trang_thai = 'san_sang'
    AND nha_cung_cap_id IN ('uuid-1', 'uuid-2', 'uuid-3');
```

---

## 🔒 Lưu Ý Quan Trọng

1. **Xác minh tài khoản**: Phải xác minh trước khi chi trả lần đầu
2. **Kiểm tra trùng lặp**: Tránh chi trả 2 lần cho cùng 1 booking
3. **Lưu mã giao dịch**: Luôn lưu `ma_giao_dich` từ ngân hàng
4. **Backup**: Backup dữ liệu trước khi chi trả hàng loạt

---

## 📱 API Endpoints (Cần Implement)

### Admin
- `GET /admin/payouts` - Danh sách chi trả
- `POST /admin/payouts/{id}/approve` - Duyệt chi trả
- `POST /admin/payouts/{id}/complete` - Hoàn thành chi trả
- `GET /admin/suppliers/no-bank` - NCC chưa có TK

### Supplier  
- `GET /supplier/payouts` - Xem chi trả của mình
- `POST /supplier/bank-accounts` - Thêm TK ngân hàng
- `GET /supplier/bank-accounts` - Danh sách TK

---

## 🎯 Tóm Tắt

✅ **3 Bảng chính:**
1. `tai_khoan_ngan_hang_ncc` - TK ngân hàng NCC
2. `cau_hinh_hoa_hong` - % hoa hồng (15%)
3. `chi_tra_nha_cung_cap` - Quản lý chi trả

✅ **Tự động:** Khi thanh toán thành công → Tự động tạo chi trả

✅ **Đơn giản:** Admin chỉ cần duyệt và cập nhật trạng thái

---

**Liên hệ:** backend@travia.com
























