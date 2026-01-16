-- ===========================================
-- ADMIN STATISTICS QUERIES
-- ===========================================

-- =====================
-- 1. DASHBOARD OVERVIEW
-- =====================
-- name: SupplierOptions :many
select nha_cung_cap.id, nha_cung_cap.ten from nha_cung_cap
join nguoi_dung on nguoi_dung.id = nha_cung_cap.id
where dang_hoat_dong = true
and nguoi_dung.vai_tro = 'nha_cung_cap'
order by nguoi_dung.ngay_tao desc;

-- name: GetDashboardOverview :one
-- Tổng quan dashboard: tổng số người dùng, tour, booking, doanh thu
SELECT
(SELECT COUNT(*) FROM nguoi_dung WHERE dang_hoat_dong = TRUE) AS tong_nguoi_dung,
(SELECT COUNT(*) FROM nguoi_dung WHERE dang_hoat_dong = TRUE AND vai_tro = 'khach_hang') AS tong_khach_hang,
(SELECT COUNT(*) FROM nha_cung_cap) AS tong_nha_cung_cap,
(SELECT COUNT(*) FROM tour WHERE dang_hoat_dong = TRUE) AS tong_tour,
(SELECT COUNT(*) FROM tour WHERE dang_hoat_dong = TRUE AND trang_thai = 'cong_bo') AS tour_dang_hoat_dong,
(SELECT COUNT(*) FROM dat_cho) AS tong_dat_cho,
(SELECT COUNT(*) FROM dat_cho WHERE trang_thai = 'da_thanh_toan') AS dat_cho_da_thanh_toan,
(SELECT COALESCE(SUM(tong_tien), 0) FROM dat_cho WHERE trang_thai IN ('da_thanh_toan', 'hoan_thanh')) AS tong_doanh_thu,
(SELECT COALESCE(SUM(tong_tien), 0) FROM dat_cho
WHERE trang_thai IN ('da_thanh_toan', 'hoan_thanh')
AND DATE_TRUNC('month', ngay_dat) = DATE_TRUNC('month', CURRENT_DATE)) AS doanh_thu_thang_nay,
(SELECT COUNT(*) FROM danh_gia WHERE dang_hoat_dong = TRUE) AS tong_danh_gia,
(SELECT COALESCE(AVG(diem_danh_gia), 0) FROM danh_gia WHERE dang_hoat_dong = TRUE) AS diem_danh_gia_trung_binh;

-- name: GetDashboardOverviewByMonthAndYear :one
-- Tổng quan dashboard: tổng số người dùng, tour, booking, doanh thu
SELECT 
    -- 1. ĐẶT CHỖ
    (SELECT COUNT(*) FROM dat_cho 
     WHERE (sqlc.arg('thang')::int = 0 OR EXTRACT(MONTH FROM ngay_dat)::int = sqlc.arg('thang')::int)
       AND (sqlc.arg('nam')::int = 0 OR EXTRACT(YEAR FROM ngay_dat)::int = sqlc.arg('nam')::int)
    ) AS tong_dat_cho,

    (SELECT COUNT(*) FROM dat_cho 
     WHERE trang_thai = 'da_huy'
       AND (sqlc.arg('thang')::int = 0 OR EXTRACT(MONTH FROM ngay_dat)::int = sqlc.arg('thang')::int)
       AND (sqlc.arg('nam')::int = 0 OR EXTRACT(YEAR FROM ngay_dat)::int = sqlc.arg('nam')::int)
    ) AS so_don_da_huy,

    -- 2. DOANH THU (Chỉ tính đơn thành công)
    (SELECT COALESCE(SUM(tong_tien), 0) FROM dat_cho 
     WHERE trang_thai IN ('da_thanh_toan', 'hoan_thanh') 
       AND (sqlc.arg('thang')::int = 0 OR EXTRACT(MONTH FROM ngay_dat)::int = sqlc.arg('thang')::int)
       AND (sqlc.arg('nam')::int = 0 OR EXTRACT(YEAR FROM ngay_dat)::int = sqlc.arg('nam')::int)
    ) AS doanh_thu,

    -- 3. VẬN HÀNH
    (SELECT COUNT(*) FROM khoi_hanh_tour 
     WHERE (sqlc.arg('thang')::int = 0 OR EXTRACT(MONTH FROM ngay_khoi_hanh)::int = sqlc.arg('thang')::int)
       AND (sqlc.arg('nam')::int = 0 OR EXTRACT(YEAR FROM ngay_khoi_hanh)::int = sqlc.arg('nam')::int)
    ) AS so_chuyen_khoi_hanh,

    -- 4. KHÁCH HÀNG
    (SELECT COALESCE(SUM(so_nguoi_lon + so_tre_em), 0) FROM dat_cho 
     WHERE trang_thai IN ('da_thanh_toan', 'hoan_thanh')
       AND (sqlc.arg('thang')::int = 0 OR EXTRACT(MONTH FROM ngay_dat)::int = sqlc.arg('thang')::int)
       AND (sqlc.arg('nam')::int = 0 OR EXTRACT(YEAR FROM ngay_dat)::int = sqlc.arg('nam')::int)
    ) AS tong_luong_khach,

    -- 5. ĐÁNH GIÁ
    (SELECT COUNT(*) FROM danh_gia 
     WHERE (sqlc.arg('thang')::int = 0 OR EXTRACT(MONTH FROM ngay_tao)::int = sqlc.arg('thang')::int)
       AND (sqlc.arg('nam')::int = 0 OR EXTRACT(YEAR FROM ngay_tao)::int = sqlc.arg('nam')::int)
    ) AS so_danh_gia_moi,

    (SELECT COALESCE(AVG(diem_danh_gia), 0) FROM danh_gia 
     WHERE (sqlc.arg('thang')::int = 0 OR EXTRACT(MONTH FROM ngay_tao)::int = sqlc.arg('thang')::int)
       AND (sqlc.arg('nam')::int = 0 OR EXTRACT(YEAR FROM ngay_tao)::int = sqlc.arg('nam')::int)
    ) AS diem_trung_binh;

-- name: GetRevenueByDay :many
-- Doanh thu theo năm và tháng
SELECT 
    DATE(dc.ngay_dat) AS ngay,
    COUNT(dc.id) AS so_booking,
    COALESCE(SUM(dc.tong_tien), 0) AS doanh_thu
FROM dat_cho dc
JOIN khoi_hanh_tour kh ON dc.khoi_hanh_id = kh.id
JOIN tour t ON kh.tour_id = t.id
WHERE 
    -- 1. Lọc Năm: Nếu tham số 'nam' truyền vào là 0, bỏ qua bộ lọc này
    (sqlc.arg('nam')::int = 0 OR EXTRACT(YEAR FROM dc.ngay_dat)::INT = sqlc.arg('nam')::int)
    
    -- 2. Lọc Tháng: Nếu tham số 'thang' truyền vào là 0, bỏ qua bộ lọc này
    AND (sqlc.arg('thang')::int = 0 OR EXTRACT(MONTH FROM dc.ngay_dat)::INT = sqlc.arg('thang')::int)
    
    -- 3. Lọc Nhà cung cấp: Nếu truyền vào NULL (narg), lấy tất cả
    AND (sqlc.narg('nha_cung_cap_id')::uuid IS NULL OR t.nha_cung_cap_id = sqlc.narg('nha_cung_cap_id'))
    
    -- 4. Trạng thái bắt buộc
    AND dc.trang_thai IN ('da_thanh_toan', 'hoan_thanh')
GROUP BY DATE(dc.ngay_dat)
ORDER BY ngay ASC;

-- =====================
-- 2. USER STATISTICS
-- =====================

-- name: GetUserStatsByRole :many
-- Thống kê người dùng theo vai trò
SELECT 
    vai_tro,
    COUNT(*) AS so_luong,
    COUNT(*) FILTER (WHERE dang_hoat_dong = TRUE) AS dang_hoat_dong,
    COUNT(*) FILTER (WHERE xac_thuc = TRUE) AS da_xac_thuc
FROM nguoi_dung
GROUP BY vai_tro
ORDER BY so_luong DESC;

-- name: GetTopBookedTours :many
-- Top tours được đặt nhiều nhất
SELECT 
    t.id,
    t.tieu_de,
    t.gia_nguoi_lon,
    ncc.ten AS ten_nha_cung_cap,
    dmt.ten AS ten_danh_muc,
    COUNT(dc.id) AS so_booking,
    COALESCE(SUM(dc.tong_tien), 0) AS tong_doanh_thu,
    COALESCE(AVG(dg.diem_danh_gia), 0) AS diem_trung_binh,
    (SELECT duong_dan FROM anh_tour WHERE tour_id = t.id AND la_anh_chinh = TRUE LIMIT 1) AS anh_chinh
FROM tour t
LEFT JOIN dat_cho dc ON dc.khoi_hanh_id IN (SELECT id FROM khoi_hanh_tour WHERE tour_id = t.id)
LEFT JOIN nha_cung_cap ncc ON t.nha_cung_cap_id = ncc.id
LEFT JOIN danh_muc_tour dmt ON t.danh_muc_id = dmt.id
LEFT JOIN danh_gia dg ON t.id = dg.tour_id AND dg.dang_hoat_dong = TRUE
WHERE t.dang_hoat_dong = TRUE AND t.trang_thai = 'cong_bo'
GROUP BY t.id, t.tieu_de, t.gia_nguoi_lon, ncc.ten, dmt.ten
ORDER BY so_booking DESC
LIMIT $1;

-- name: GetTourPriceDistribution :many
-- Phân bố giá tour
SELECT 
    CASE 
        WHEN gia_nguoi_lon < 1000000 THEN 'Dưới 1 triệu'
        WHEN gia_nguoi_lon >= 1000000 AND gia_nguoi_lon < 3000000 THEN '1-3 triệu'
        WHEN gia_nguoi_lon >= 3000000 AND gia_nguoi_lon < 5000000 THEN '3-5 triệu'
        WHEN gia_nguoi_lon >= 5000000 AND gia_nguoi_lon < 10000000 THEN '5-10 triệu'
        ELSE 'Trên 10 triệu'
    END AS khoang_gia,
    COUNT(*) AS so_luong
FROM tour
WHERE dang_hoat_dong = TRUE AND trang_thai = 'cong_bo'
GROUP BY khoang_gia
ORDER BY 
    CASE khoang_gia
        WHEN 'Dưới 1 triệu' THEN 1
        WHEN '1-3 triệu' THEN 2
        WHEN '3-5 triệu' THEN 3
        WHEN '5-10 triệu' THEN 4
        ELSE 5
    END;

-- name: GetBookingsByDayOfWeek :many
-- Thống kê booking theo ngày trong tuần
SELECT 
    EXTRACT(DOW FROM ngay_dat) AS ngay_trong_tuan,
    CASE EXTRACT(DOW FROM ngay_dat)
        WHEN 0 THEN 'Chủ nhật'
        WHEN 1 THEN 'Thứ hai'
        WHEN 2 THEN 'Thứ ba'
        WHEN 3 THEN 'Thứ tư'
        WHEN 4 THEN 'Thứ năm'
        WHEN 5 THEN 'Thứ sáu'
        WHEN 6 THEN 'Thứ bảy'
    END AS ten_ngay,
    COUNT(*) AS so_booking,
    COALESCE(SUM(tong_tien), 0) AS doanh_thu
FROM dat_cho
WHERE trang_thai IN ('da_thanh_toan', 'hoan_thanh')
GROUP BY EXTRACT(DOW FROM ngay_dat)
ORDER BY ngay_trong_tuan;

-- name: GetAllBookingsForAdmin :many
-- Lấy tất cả booking cho admin với filter và pagination
SELECT 
    dc.id,
    dc.nguoi_dung_id,
    dc.khoi_hanh_id,
    dc.so_nguoi_lon,
    dc.so_tre_em,
    dc.tong_tien,
    dc.don_vi_tien_te,
    dc.trang_thai,
    dc.phuong_thuc_thanh_toan,
    dc.ngay_dat,
    dc.ngay_cap_nhat,
    
    -- Thông tin khách hàng
    nd.id AS customer_id,
    nd.ho_ten AS customer_name,
    nd.email AS customer_email,
    nd.so_dien_thoai AS customer_phone,
    
    -- Thông tin tour
    t.id AS tour_id,
    t.tieu_de AS tour_title,
    ncc.id AS supplier_id,
    ncc.ten AS supplier_name,
    
    -- Thông tin khởi hành
    kh.id AS departure_id,
    kh.ngay_khoi_hanh,
    kh.ngay_ket_thuc,
    kh.trang_thai AS departure_status
FROM dat_cho dc
JOIN nguoi_dung nd ON nd.id = dc.nguoi_dung_id
JOIN khoi_hanh_tour kh ON kh.id = dc.khoi_hanh_id
JOIN tour t ON t.id = kh.tour_id
JOIN nha_cung_cap ncc ON ncc.id = t.nha_cung_cap_id
WHERE 
    -- Filter theo thời gian đặt chỗ
    (sqlc.narg('start_date')::timestamp IS NULL OR dc.ngay_dat >= sqlc.narg('start_date')::timestamp)
    AND (sqlc.narg('end_date')::timestamp IS NULL OR dc.ngay_dat <= sqlc.narg('end_date')::timestamp)
    -- Filter theo nhà cung cấp
    AND (sqlc.narg('supplier_id')::uuid IS NULL OR t.nha_cung_cap_id = sqlc.narg('supplier_id')::uuid)
    -- Filter theo trạng thái
    AND (sqlc.narg('trang_thai')::text IS NULL OR dc.trang_thai::text = sqlc.narg('trang_thai')::text)
    -- Filter theo tìm kiếm
    AND (
        sqlc.narg('search')::text IS NULL 
        OR sqlc.narg('search')::text = ''
        OR nd.ho_ten ILIKE '%' || sqlc.narg('search')::text || '%'
        OR nd.email ILIKE '%' || sqlc.narg('search')::text || '%'
        OR t.tieu_de ILIKE '%' || sqlc.narg('search')::text || '%'
        OR dc.id::text = sqlc.narg('search')::text
    )
ORDER BY dc.ngay_dat DESC
LIMIT sqlc.arg('limit')::int OFFSET sqlc.arg('offset')::int;

-- name: CountAllBookingsForAdmin :one
-- Đếm tổng số booking cho admin với filter
SELECT COUNT(*)::int AS total
FROM dat_cho dc
JOIN nguoi_dung nd ON nd.id = dc.nguoi_dung_id
JOIN khoi_hanh_tour kh ON kh.id = dc.khoi_hanh_id
JOIN tour t ON t.id = kh.tour_id
JOIN nha_cung_cap ncc ON ncc.id = t.nha_cung_cap_id
WHERE 
    -- Filter theo thời gian đặt chỗ
    (sqlc.narg('start_date')::timestamp IS NULL OR dc.ngay_dat >= sqlc.narg('start_date')::timestamp)
    AND (sqlc.narg('end_date')::timestamp IS NULL OR dc.ngay_dat <= sqlc.narg('end_date')::timestamp)
    -- Filter theo nhà cung cấp
    AND (sqlc.narg('supplier_id')::uuid IS NULL OR t.nha_cung_cap_id = sqlc.narg('supplier_id')::uuid)
    -- Filter theo trạng thái
    AND (sqlc.narg('trang_thai')::text IS NULL OR dc.trang_thai::text = sqlc.narg('trang_thai')::text)
    -- Filter theo tìm kiếm
    AND (
        sqlc.narg('search')::text IS NULL 
        OR sqlc.narg('search')::text = ''
        OR nd.ho_ten ILIKE '%' || sqlc.narg('search')::text || '%'
        OR nd.email ILIKE '%' || sqlc.narg('search')::text || '%'
        OR t.tieu_de ILIKE '%' || sqlc.narg('search')::text || '%'
        OR dc.id::text = sqlc.narg('search')::text
    );

-- name: GetRecentBookings :many
-- Booking gần đây
SELECT 
    dc.id,
    dc.ngay_dat,
    dc.tong_tien,
    dc.trang_thai,
    dc.so_nguoi_lon,
    dc.so_tre_em,
    nd.ho_ten AS ten_khach_hang,
    nd.email AS email_khach_hang,
    t.tieu_de AS ten_tour,
    kh.ngay_khoi_hanh
FROM dat_cho dc
JOIN nguoi_dung nd ON dc.nguoi_dung_id = nd.id
JOIN khoi_hanh_tour kh ON dc.khoi_hanh_id = kh.id
JOIN tour t ON kh.tour_id = t.id
ORDER BY dc.ngay_dat DESC
LIMIT $1;

-- name: GetSupplierStats :one
-- Tổng quan nhà cung cấp
SELECT 
    COUNT(*) AS tong_nha_cung_cap,
    (SELECT COUNT(*) FROM tour WHERE dang_hoat_dong = TRUE) AS tong_tour,
    (SELECT COUNT(*) FROM tour WHERE dang_hoat_dong = TRUE AND trang_thai = 'cong_bo') AS tour_dang_cong_bo;

-- name: GetTopSuppliersByRevenue :many
-- Top nhà cung cấp theo doanh thu
SELECT 
    ncc.id,
    ncc.ten,
    nd.email,
    ncc.logo,
    COUNT(DISTINCT t.id) AS so_tour,
    COUNT(dc.id) AS so_booking,
    COALESCE(SUM(dc.tong_tien), 0) AS tong_doanh_thu,
    COALESCE(AVG(dg.diem_danh_gia), 0) AS diem_trung_binh
FROM nha_cung_cap ncc
JOIN nguoi_dung nd ON ncc.id = nd.id
LEFT JOIN tour t ON ncc.id = t.nha_cung_cap_id AND t.dang_hoat_dong = TRUE
LEFT JOIN khoi_hanh_tour kh ON t.id = kh.tour_id
LEFT JOIN dat_cho dc ON kh.id = dc.khoi_hanh_id AND dc.trang_thai IN ('da_thanh_toan', 'hoan_thanh')
LEFT JOIN danh_gia dg ON t.id = dg.tour_id AND dg.dang_hoat_dong = TRUE
GROUP BY ncc.id, ncc.ten, nd.email, ncc.logo
ORDER BY tong_doanh_thu DESC
LIMIT $1;

-- name: GetTopSuppliersByTours :many
-- Top nhà cung cấp theo số tour
SELECT 
    ncc.id,
    ncc.ten,
    nd.email,
    ncc.logo,
    COUNT(t.id) AS so_tour,
    COUNT(t.id) FILTER (WHERE t.trang_thai = 'cong_bo') AS tour_cong_bo,
    COUNT(t.id) FILTER (WHERE t.noi_bat = TRUE) AS tour_noi_bat
FROM nha_cung_cap ncc
JOIN nguoi_dung nd ON ncc.id = nd.id
LEFT JOIN tour t ON ncc.id = t.nha_cung_cap_id AND t.dang_hoat_dong = TRUE
GROUP BY ncc.id, ncc.ten, nd.email, ncc.logo
ORDER BY so_tour DESC
LIMIT $1;

-- name: GetSupplierGrowthByMonth :many
-- Tăng trưởng nhà cung cấp theo tháng
SELECT 
    DATE_TRUNC('month', nd.ngay_tao) AS thang,
    COUNT(*) AS so_nha_cung_cap_moi
FROM nha_cung_cap ncc
JOIN nguoi_dung nd ON ncc.id = nd.id
WHERE nd.ngay_tao >= NOW() - INTERVAL '12 months'
GROUP BY DATE_TRUNC('month', nd.ngay_tao)
ORDER BY thang ASC;

-- name: GetTopDestinations :many
-- Top điểm đến phổ biến
SELECT 
    dd.id,
    dd.ten,
    dd.tinh,
    dd.quoc_gia,
    dd.anh,
    COUNT(DISTINCT tdd.tour_id) AS so_tour,
    COALESCE(SUM(dc.tong_tien), 0) AS tong_doanh_thu
FROM diem_den dd
LEFT JOIN tour_diem_den tdd ON dd.id = tdd.diem_den_id
LEFT JOIN tour t ON tdd.tour_id = t.id AND t.dang_hoat_dong = TRUE
LEFT JOIN khoi_hanh_tour kh ON t.id = kh.tour_id
LEFT JOIN dat_cho dc ON kh.id = dc.khoi_hanh_id AND dc.trang_thai IN ('da_thanh_toan', 'hoan_thanh')
GROUP BY dd.id, dd.ten, dd.tinh, dd.quoc_gia, dd.anh
ORDER BY so_tour DESC
LIMIT $1;


-- =====================
-- 9. DEPARTURE STATISTICS
-- =====================

-- name: GetUpcomingDepartures :many
-- Các khởi hành sắp tới
SELECT 
    kh.id,
    kh.ngay_khoi_hanh,
    kh.ngay_ket_thuc,
    kh.suc_chua,
    kh.so_cho_da_dat,
    kh.trang_thai,
    t.tieu_de AS ten_tour,
    t.gia_nguoi_lon,
    ncc.ten AS ten_nha_cung_cap,
    (kh.suc_chua - kh.so_cho_da_dat) AS cho_con_trong
FROM khoi_hanh_tour kh
JOIN tour t ON kh.tour_id = t.id
LEFT JOIN nha_cung_cap ncc ON t.nha_cung_cap_id = ncc.id
WHERE kh.ngay_khoi_hanh >= CURRENT_DATE
  AND kh.trang_thai IN ('len_lich', 'xac_nhan', 'con_cho')
ORDER BY kh.ngay_khoi_hanh ASC
LIMIT $1;

-- name: GetMostFavoritedTours :many
-- Tours được yêu thích nhiều nhất
SELECT 
    t.id,
    t.tieu_de,
    t.gia_nguoi_lon,
    ncc.ten AS ten_nha_cung_cap,
    COUNT(tyt.id) AS so_yeu_thich,
    (SELECT duong_dan FROM anh_tour WHERE tour_id = t.id AND la_anh_chinh = TRUE LIMIT 1) AS anh_chinh
FROM tour t
LEFT JOIN tour_yeu_thich tyt ON t.id = tyt.tour_id
LEFT JOIN nha_cung_cap ncc ON t.nha_cung_cap_id = ncc.id
WHERE t.dang_hoat_dong = TRUE AND t.trang_thai = 'cong_bo'
GROUP BY t.id, t.tieu_de, t.gia_nguoi_lon, ncc.ten
ORDER BY so_yeu_thich DESC
LIMIT $1;

-- =====================
-- 12. COMPREHENSIVE REPORTS
-- =====================

-- name: GetMonthlyReport :one
-- Báo cáo tháng
SELECT 
    -- Doanh thu
    COALESCE(SUM(dc.tong_tien) FILTER (WHERE dc.trang_thai IN ('da_thanh_toan', 'hoan_thanh')), 0) AS doanh_thu,
    COUNT(dc.id) AS tong_booking,
    COUNT(dc.id) FILTER (WHERE dc.trang_thai IN ('da_thanh_toan', 'hoan_thanh')) AS booking_thanh_cong,
    COUNT(dc.id) FILTER (WHERE dc.trang_thai = 'da_huy') AS booking_huy,
    
    -- Người dùng mới
    (SELECT COUNT(*) FROM nguoi_dung 
     WHERE DATE_TRUNC('month', ngay_tao) = DATE_TRUNC('month', $1::DATE)) AS nguoi_dung_moi,
    
    -- Tour mới
    (SELECT COUNT(*) FROM tour 
     WHERE DATE_TRUNC('month', ngay_tao) = DATE_TRUNC('month', $1::DATE)) AS tour_moi,
    
    -- Đánh giá
    (SELECT COUNT(*) FROM danh_gia 
     WHERE DATE_TRUNC('month', ngay_tao) = DATE_TRUNC('month', $1::DATE) AND dang_hoat_dong = TRUE) AS danh_gia_moi,
    (SELECT COALESCE(AVG(diem_danh_gia), 0) FROM danh_gia 
     WHERE DATE_TRUNC('month', ngay_tao) = DATE_TRUNC('month', $1::DATE) AND dang_hoat_dong = TRUE) AS diem_danh_gia_trung_binh

FROM dat_cho dc
WHERE DATE_TRUNC('month', dc.ngay_dat) = DATE_TRUNC('month', $1::DATE);

-- name: GetYearlyComparisonReport :many
-- So sánh theo năm
SELECT 
    EXTRACT(YEAR FROM ngay_dat) AS nam,
    COUNT(*) AS tong_booking,
    COALESCE(SUM(tong_tien) FILTER (WHERE trang_thai IN ('da_thanh_toan', 'hoan_thanh')), 0) AS doanh_thu,
    COALESCE(AVG(tong_tien) FILTER (WHERE trang_thai IN ('da_thanh_toan', 'hoan_thanh')), 0) AS gia_tri_trung_binh
FROM dat_cho
GROUP BY EXTRACT(YEAR FROM ngay_dat)
ORDER BY nam ASC;

-- name: GetQuarterlyReport :many
-- Báo cáo theo quý
SELECT 
    EXTRACT(YEAR FROM ngay_dat) AS nam,
    EXTRACT(QUARTER FROM ngay_dat) AS quy,
    COUNT(*) AS tong_booking,
    COALESCE(SUM(tong_tien) FILTER (WHERE trang_thai IN ('da_thanh_toan', 'hoan_thanh')), 0) AS doanh_thu
FROM dat_cho
WHERE ngay_dat >= NOW() - INTERVAL '2 years'
GROUP BY EXTRACT(YEAR FROM ngay_dat), EXTRACT(QUARTER FROM ngay_dat)
ORDER BY nam ASC, quy ASC;




-- name: AdminChartRevenueTrend :many
-- Xu hướng Doanh thu & Đặt chỗ
SELECT 
    DATE(ngay_dat) AS ngay,
    COUNT(id) AS tong_so_don,
    COALESCE(SUM(tong_tien), 0) AS doanh_thu_ngay,
    COALESCE(SUM(so_nguoi_lon + so_tre_em), 0) AS tong_khach_ngay
FROM dat_cho
WHERE 
    (sqlc.arg('nam')::int = 0 OR EXTRACT(YEAR FROM ngay_dat)::int = sqlc.arg('nam')::int)
    AND (sqlc.arg('thang')::int = 0 OR EXTRACT(MONTH FROM ngay_dat)::int = sqlc.arg('thang')::int)
    AND trang_thai IN ('da_thanh_toan', 'hoan_thanh')
GROUP BY DATE(ngay_dat)
ORDER BY ngay ASC;
-- name: AdminChartCategoryDistribution :many
--Cơ cấu Doanh thu theo Danh mục
SELECT 
    dm.ten AS ten_danh_muc,
    COUNT(dc.id) AS so_luong_dat,
    COALESCE(SUM(dc.tong_tien), 0)::numeric AS tong_doanh_thu
FROM dat_cho dc
JOIN khoi_hanh_tour kh ON dc.khoi_hanh_id = kh.id
JOIN tour t ON kh.tour_id = t.id
JOIN danh_muc_tour dm ON t.danh_muc_id = dm.id
WHERE 
    (sqlc.arg('nam')::int = 0 OR EXTRACT(YEAR FROM dc.ngay_dat)::int = sqlc.arg('nam')::int)
    AND (sqlc.arg('thang')::int = 0 OR EXTRACT(MONTH FROM dc.ngay_dat)::int = sqlc.arg('thang')::int)
    AND dc.trang_thai IN ('da_thanh_toan', 'hoan_thanh')
GROUP BY dm.id, dm.ten
ORDER BY tong_doanh_thu DESC;
-- name: AdminChartTopSuppliers :many
--Top 5 Nhà cung cấp xuất sắc
SELECT 
    ncc.ten AS ten_nha_cung_cap,
    COUNT(dc.id) AS so_don_hang,
    COALESCE(SUM(dc.tong_tien), 0)::numeric AS doanh_thu_dat_duoc
FROM dat_cho dc
JOIN khoi_hanh_tour kh ON dc.khoi_hanh_id = kh.id
JOIN tour t ON kh.tour_id = t.id
JOIN nha_cung_cap ncc ON t.nha_cung_cap_id = ncc.id
WHERE 
    (sqlc.arg('nam')::int = 0 OR EXTRACT(YEAR FROM dc.ngay_dat)::int = sqlc.arg('nam')::int)
    AND (sqlc.arg('thang')::int = 0 OR EXTRACT(MONTH FROM dc.ngay_dat)::int = sqlc.arg('thang')::int)
    AND dc.trang_thai IN ('da_thanh_toan', 'hoan_thanh')
GROUP BY ncc.id, ncc.ten
ORDER BY doanh_thu_dat_duoc DESC
LIMIT 5;
-- name: AdminChartBookingStatusStats :many
--Trạng thái Đặt chỗ
SELECT 
    trang_thai,
    COUNT(*) AS so_luong,
    SUM(tong_tien)::numeric AS gia_tri_uoc_tinh
FROM dat_cho
WHERE 
    (sqlc.arg('nam')::int = 0 OR EXTRACT(YEAR FROM ngay_dat)::int = sqlc.arg('nam')::int)
    AND (sqlc.arg('thang')::int = 0 OR EXTRACT(MONTH FROM ngay_dat)::int = sqlc.arg('thang')::int)
GROUP BY trang_thai;

-- name: GetAdminBookingStatistics :one
-- Thống kê đặt chỗ chi tiết cho admin với nhiều filter
SELECT 
    -- Tổng số booking
    COUNT(*)::int AS tong_so_booking,
    
    -- Thống kê theo trạng thái
    COUNT(*) FILTER (WHERE dc.trang_thai = 'cho_xac_nhan')::int AS cho_xac_nhan,
    COUNT(*) FILTER (WHERE dc.trang_thai = 'da_xac_nhan')::int AS da_xac_nhan,
    COUNT(*) FILTER (WHERE dc.trang_thai = 'da_thanh_toan')::int AS da_thanh_toan,
    COUNT(*) FILTER (WHERE dc.trang_thai = 'hoan_thanh')::int AS hoan_thanh,
    COUNT(*) FILTER (WHERE dc.trang_thai = 'da_huy')::int AS da_huy,
    
    -- Tổng tiền
    COALESCE(SUM(dc.tong_tien), 0)::numeric AS tong_tien,
    COALESCE(SUM(dc.tong_tien) FILTER (WHERE dc.trang_thai IN ('da_thanh_toan', 'hoan_thanh')), 0)::numeric AS tong_doanh_thu,
    COALESCE(SUM(dc.tong_tien) FILTER (WHERE dc.trang_thai = 'da_huy'), 0)::numeric AS tong_tien_da_huy,
    
    -- Tổng số khách hàng
    COUNT(DISTINCT dc.nguoi_dung_id)::int AS tong_so_khach_hang,
    
    -- Tổng số tour
    COUNT(DISTINCT t.id)::int AS tong_so_tour,
    
    -- Tổng số nhà cung cấp
    COUNT(DISTINCT t.nha_cung_cap_id)::int AS tong_so_nha_cung_cap,
    
    -- Tổng số khách (người lớn + trẻ em)
    COALESCE(SUM(COALESCE(dc.so_nguoi_lon, 0) + COALESCE(dc.so_tre_em, 0)), 0)::int AS tong_so_khach,
    COALESCE(SUM(COALESCE(dc.so_nguoi_lon, 0) + COALESCE(dc.so_tre_em, 0)) FILTER (WHERE dc.trang_thai IN ('da_thanh_toan', 'hoan_thanh')), 0)::int AS tong_so_khach_thanh_cong,
    
    -- Trung bình giá trị booking
    COALESCE(AVG(dc.tong_tien), 0)::numeric AS gia_tri_trung_binh,
    COALESCE(AVG(dc.tong_tien) FILTER (WHERE dc.trang_thai IN ('da_thanh_toan', 'hoan_thanh')), 0)::numeric AS gia_tri_trung_binh_thanh_cong
FROM dat_cho dc
JOIN khoi_hanh_tour kh ON kh.id = dc.khoi_hanh_id
JOIN tour t ON t.id = kh.tour_id
WHERE 
    -- Filter theo thời gian đặt chỗ
    (sqlc.narg('start_date')::timestamp IS NULL OR dc.ngay_dat >= sqlc.narg('start_date')::timestamp)
    AND (sqlc.narg('end_date')::timestamp IS NULL OR dc.ngay_dat <= sqlc.narg('end_date')::timestamp)
    -- Filter theo nhà cung cấp
    AND (sqlc.narg('supplier_id')::uuid IS NULL OR t.nha_cung_cap_id = sqlc.narg('supplier_id')::uuid)
    -- Filter theo trạng thái (có thể filter nhiều trạng thái)
    AND (sqlc.narg('trang_thai')::text IS NULL OR dc.trang_thai::text = sqlc.narg('trang_thai')::text);

--=====================================Nhà cung cấp=====================================
-- name: GetAllSuppliers :many
SELECT ncc.*, nd.ho_ten, nd.email, nd.so_dien_thoai, nd.ngay_tao, nd.xac_thuc, nd.dang_hoat_dong
FROM nha_cung_cap ncc
JOIN nguoi_dung nd ON nd.id = ncc.id
WHERE 
  (sqlc.narg('xac_thuc')::boolean IS NULL OR nd.xac_thuc = sqlc.narg('xac_thuc'))
  AND (sqlc.narg('dang_hoat_dong')::boolean IS NULL OR nd.dang_hoat_dong = sqlc.narg('dang_hoat_dong'))
ORDER BY nd.ngay_tao DESC;
-- name: SoftDeleteSupplier :exec
UPDATE nguoi_dung
SET
    dang_hoat_dong = FALSE,
    ngay_cap_nhat = CURRENT_TIMESTAMP
WHERE id = $1 AND vai_tro = 'nha_cung_cap';
-- name: RestoreSupplier :one
UPDATE nguoi_dung
SET
    dang_hoat_dong = TRUE,
    ngay_cap_nhat = CURRENT_TIMESTAMP
WHERE id = $1 AND vai_tro = 'nha_cung_cap'
RETURNING *;

-- name: DeleteSupplier :exec
DELETE FROM nha_cung_cap
WHERE id = $1 AND nha_cung_cap.dang_hoat_dong = TRUE;
-- name: RejectSupplier :one
-- từ chối nhà cung cấp
UPDATE nguoi_dung
SET 
    dang_hoat_dong = FALSE,
    xac_thuc = FALSE,
    ngay_cap_nhat = CURRENT_TIMESTAMP
WHERE id = $1 
    AND vai_tro = 'nha_cung_cap'
RETURNING *;
-- name: ApproveSupplier :one
-- phê duyệt nhà cung cấp
UPDATE nguoi_dung
SET 
    dang_hoat_dong = TRUE,
    xac_thuc = TRUE,
    ngay_cap_nhat = CURRENT_TIMESTAMP
WHERE id = $1 
    AND vai_tro = 'nha_cung_cap'
    AND dang_hoat_dong = TRUE
RETURNING *;

-- name: GetAdminSupplierByID :one
-- Lấy nhà cung cấp theo ID (admin)
SELECT ncc.*, nd.ho_ten, nd.email, nd.so_dien_thoai, nd.ngay_tao, nd.ngay_cap_nhat, nd.dang_hoat_dong, nd.xac_thuc
FROM nha_cung_cap ncc
JOIN nguoi_dung nd ON nd.id = ncc.id
WHERE ncc.id = $1;

--=====================================Khách hàng=====================================
-- name: GetTopActiveUsers :many
-- Top người dùng hoạt động nhiều nhất (theo số booking)
SELECT 
    nd.id,
    nd.ho_ten,
    nd.email,
    COUNT(dc.id) AS so_booking,
    COALESCE(SUM(dc.tong_tien), 0) AS tong_chi_tieu
FROM nguoi_dung nd
LEFT JOIN dat_cho dc ON nd.id = dc.nguoi_dung_id
WHERE nd.vai_tro = 'khach_hang' AND nd.dang_hoat_dong = TRUE
GROUP BY nd.id, nd.ho_ten, nd.email
ORDER BY so_booking DESC
LIMIT $1;
-- name: AdminCustomerGrowthMonthlyReport :many
WITH MonthlyStats AS (
    SELECT 
        EXTRACT(YEAR FROM ngay_tao)::int AS nam,
        EXTRACT(MONTH FROM ngay_tao)::int AS thang,
        COUNT(id) AS so_luong
    FROM nguoi_dung
    WHERE vai_tro = 'khach_hang' 
      AND dang_hoat_dong = TRUE
    GROUP BY 1, 2
)
SELECT 
    nam,
    thang,
    so_luong AS khach_moi_thang_nay,
    COALESCE(LAG(so_luong) OVER (ORDER BY nam, thang), 0) AS khach_moi_thang_truoc,
    -- Tính % tăng trưởng
    ROUND(
        CASE 
            WHEN LAG(so_luong) OVER (ORDER BY nam, thang) IS NULL OR LAG(so_luong) OVER (ORDER BY nam, thang) = 0 THEN 100
            ELSE ((so_luong - LAG(so_luong) OVER (ORDER BY nam, thang))::float / LAG(so_luong) OVER (ORDER BY nam, thang) * 100)
        END::numeric, 2
    ) AS phan_tram_tang_truong
FROM MonthlyStats
WHERE (sqlc.arg('nam')::int = 0 OR nam = sqlc.arg('nam')::int)
ORDER BY nam DESC, thang DESC;