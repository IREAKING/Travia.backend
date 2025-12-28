# 📝 Thay Đổi Hệ Thống Chi Trả NCC (Đơn Giản Hóa)

## ✅ Đã Đơn Giản Hóa

### 1. **Bảng `tai_khoan_ngan_hang_ncc`**

**Trước (quá phức tạp):**
- ❌ 15+ trường (swift_code, iban, routing_number, giới hạn...)
- ❌ Nhiều trường không dùng cho Việt Nam

**Sau (đơn giản):**
- ✅ Chỉ 9 trường cơ bản
- ✅ ten_tai_khoan, so_tai_khoan, ten_ngan_hang, chi_nhanh
- ✅ la_tai_khoan_mac_dinh, da_xac_minh
- ✅ ghi_chu

### 2. **Bảng `cau_hinh_hoa_hong`**

**Trước:**
- ❌ Phức tạp: theo NCC, theo danh mục, theo thời gian
- ❌ Nhiều logic kiểm tra

**Sau:**
- ✅ Chỉ 1 setting toàn hệ thống: `ti_le_hoa_hong_platform = 15%`
- ✅ Nếu cần tùy chỉnh → Dùng `nha_cung_cap.ti_le_hoa_hong_platform`

### 3. **Bảng `chi_tra_nha_cung_cap`**

**Trước:**
- ❌ 30+ trường
- ❌ Nhiều computed columns phức tạp
- ❌ Điều chỉnh thêm/trừ, tranh chấp...
- ❌ Ngày tour, ngày sẵn sàng, ngày dự kiến...

**Sau:**
- ✅ Chỉ 11 trường cần thiết
- ✅ Không có computed columns (tính ở application)
- ✅ Giản lược: chỉ tổng tiền, % hoa hồng, tiền cho NCC
- ✅ Trạng thái đơn giản

### 4. **Đã Xóa (Không Cần Thiết)**

- ❌ Bảng `lich_su_chi_tra_ncc` - Có thể dùng audit log chung
- ❌ Bảng `chu_ky_chi_tra` - Chi trả theo yêu cầu, không cần chu kỳ
- ❌ Bảng `chi_tra_chu_ky` - Không cần nữa

### 5. **Functions & Triggers**

**Trước:**
- Function `lay_ti_le_hoa_hong()` - 60 dòng, nhiều logic
- Function `luu_lich_su_chi_tra()` - 50 dòng
- Function `dam_bao_tai_khoan_mac_dinh_duy_nhat()` - Vẫn giữ

**Sau:**
- Function `lay_ti_le_hoa_hong()` - Chỉ 15 dòng, đơn giản
- Trigger `tao_chi_tra_nha_cung_cap()` - Chỉ 30 dòng
- Xóa function `luu_lich_su_chi_tra()`

### 6. **Views**

**Trước:**
- 6 views phức tạp

**Sau:**
- 2 views đơn giản:
  - `v_chi_tra_ncc` - Danh sách chi trả
  - `v_ncc_chua_co_tai_khoan_ngan_hang` - Cảnh báo

---

## 📊 So Sánh

| Mục | Trước | Sau |
|-----|-------|-----|
| **Số bảng** | 5 | 3 |
| **Tổng số trường** | ~80 | ~30 |
| **Functions** | 4 | 2 |
| **Triggers** | 7 | 3 |
| **Views** | 6 | 2 |
| **Độ phức tạp** | ⭐⭐⭐⭐⭐ | ⭐⭐ |

---

## 🚀 Lợi Ích

1. **Dễ hiểu**: Schema đơn giản, dễ đọc
2. **Dễ maintain**: Ít code, ít bug
3. **Performance tốt hơn**: Ít join, ít computed columns
4. **Linh hoạt**: Dễ mở rộng khi cần

---

## 📖 Tài Liệu

- **Chi tiết**: `SIMPLE_PAYOUT_GUIDE.md`
- **Schema**: `db/migration/schema.sql`

---

## ⚠️ Migration Notes

Nếu đã có dữ liệu cũ, cần:
1. Export dữ liệu từ bảng cũ
2. Drop các bảng/views không dùng
3. Import lại data vào bảng mới

---

**Ngày thay đổi**: {{ DATE }}
**Người thực hiện**: Backend Team
























