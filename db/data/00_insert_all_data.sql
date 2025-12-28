-- ===========================================
-- MASTER FILE: Insert tất cả dữ liệu mẫu
-- ===========================================
-- File này sẽ insert dữ liệu theo đúng thứ tự phụ thuộc
-- Chạy sau khi đã tạo schema (schema.sql)
-- 
-- Cách sử dụng:
-- psql -U postgres -d travia_db -f 00_insert_all_data.sql
-- 
-- Hoặc từng file riêng lẻ:
-- psql -U postgres -d travia_db -f nguoidung.sql
-- ===========================================

\echo '=========================================='
\echo 'Bắt đầu insert dữ liệu mẫu...'
\echo '=========================================='

-- 0. CỔNG THANH TOÁN (không phụ thuộc bảng nào, cần thiết cho lich_su_giao_dich)
\echo 'Đang insert cong_thanh_toan...'
\i congthanhtoan.sql

-- 1. NGƯỜI DÙNG (bắt buộc cho các bảng khác)
\echo 'Đang insert nguoi_dung...'
\i nguoidung.sql

-- 2. DANH MỤC TOUR
\echo 'Đang insert danh_muc_tour...'
\i danhmuctour.sql

-- 3. ĐIỂM ĐẾN
\echo 'Đang insert diem_den...'
\i diemden.sql

-- 4. NHÀ CUNG CẤP (phụ thuộc nguoi_dung)
\echo 'Đang insert nha_cung_cap...'
\i nhacungcap.sql

-- 5. TOUR (phụ thuộc nguoi_dung, nha_cung_cap, danh_muc_tour)
\echo 'Đang insert tour...'
\i tour.sql

-- 6. ẢNH TOUR (phụ thuộc tour)
\echo 'Đang insert anh_tour...'
\i anhtour.sql

-- 7. TOUR - ĐIỂM ĐẾN (phụ thuộc tour, diem_den)
\echo 'Đang insert tour_diem_den...'
\i tourdiemden.sql

-- 8. LỊCH TRÌNH TOUR (phụ thuộc tour)
\echo 'Đang insert lich_trinh...'
\i lichtrinhtour.sql

-- 9. HOẠT ĐỘNG LỊCH TRÌNH (phụ thuộc lich_trinh)
\echo 'Đang insert hoat_dong_trong_ngay...'
\i hoatdonglichTrinh.sql

-- 10. CẤU HÌNH NHÓM TOUR (phụ thuộc tour)
\echo 'Đang insert cau_hinh_nhom_tour...'
\i cauhinhnhomtour.sql

-- 11. KHỞI HÀNH TOUR (phụ thuộc tour)
\echo 'Đang insert khoi_hanh_tour...'
\i khoihanhtour.sql

-- 12. GIẢM GIÁ TOUR (phụ thuộc tour - optional)
\echo 'Đang insert giam_gia_tour...'
\i giamgiatour.sql

\echo '=========================================='
\echo 'Hoàn tất insert dữ liệu mẫu!'
\echo '=========================================='

-- Kiểm tra số lượng dữ liệu đã insert
\echo ''
\echo 'Thống kê dữ liệu:'
\echo '----------------------------------------'

SELECT 'cong_thanh_toan' as bang, COUNT(*) as so_luong FROM cong_thanh_toan
UNION ALL
SELECT 'nguoi_dung', COUNT(*) FROM nguoi_dung
UNION ALL
SELECT 'danh_muc_tour', COUNT(*) FROM danh_muc_tour
UNION ALL
SELECT 'diem_den', COUNT(*) FROM diem_den
UNION ALL
SELECT 'nha_cung_cap', COUNT(*) FROM nha_cung_cap
UNION ALL
SELECT 'tour', COUNT(*) FROM tour
UNION ALL
SELECT 'anh_tour', COUNT(*) FROM anh_tour
UNION ALL
SELECT 'tour_diem_den', COUNT(*) FROM tour_diem_den
UNION ALL
SELECT 'lich_trinh', COUNT(*) FROM lich_trinh
UNION ALL
SELECT 'hoat_dong_trong_ngay', COUNT(*) FROM hoat_dong_trong_ngay
UNION ALL
SELECT 'cau_hinh_nhom_tour', COUNT(*) FROM cau_hinh_nhom_tour
UNION ALL
SELECT 'khoi_hanh_tour', COUNT(*) FROM khoi_hanh_tour
UNION ALL
SELECT 'giam_gia_tour', COUNT(*) FROM giam_gia_tour;

\echo ''
\echo 'Chi tiết tour theo trạng thái:'
SELECT trang_thai, COUNT(*) as so_luong
FROM tour
GROUP BY trang_thai
ORDER BY trang_thai;

\echo ''
\echo 'Chi tiết tour nổi bật:'
SELECT COUNT(*) as so_tour_noi_bat
FROM tour
WHERE noi_bat = TRUE AND dang_hoat_dong = TRUE;

