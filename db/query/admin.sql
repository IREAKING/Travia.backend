-- ===========================================
-- ADMIN STATISTICS QUERIES
-- ===========================================

-- =====================
-- 1. DASHBOARD OVERVIEW
-- =====================

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

-- name: GetDashboardOverviewWithComparison :one
-- Tổng quan dashboard với so sánh tháng trước
SELECT 
    -- Tổng số
    (SELECT COUNT(*) FROM nguoi_dung WHERE dang_hoat_dong = TRUE) AS tong_nguoi_dung,
    (SELECT COUNT(*) FROM tour WHERE dang_hoat_dong = TRUE AND trang_thai = 'cong_bo') AS tong_tour,
    (SELECT COUNT(*) FROM dat_cho WHERE trang_thai IN ('da_thanh_toan', 'hoan_thanh')) AS tong_dat_cho_thanh_cong,
    
    -- Doanh thu tháng này
    (SELECT COALESCE(SUM(tong_tien), 0) FROM dat_cho 
     WHERE trang_thai IN ('da_thanh_toan', 'hoan_thanh') 
     AND DATE_TRUNC('month', ngay_dat) = DATE_TRUNC('month', CURRENT_DATE)) AS doanh_thu_thang_nay,
    
    -- Doanh thu tháng trước
    (SELECT COALESCE(SUM(tong_tien), 0) FROM dat_cho 
     WHERE trang_thai IN ('da_thanh_toan', 'hoan_thanh') 
     AND DATE_TRUNC('month', ngay_dat) = DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 month')) AS doanh_thu_thang_truoc,
    
    -- Người dùng mới tháng này
    (SELECT COUNT(*) FROM nguoi_dung 
     WHERE DATE_TRUNC('month', ngay_tao) = DATE_TRUNC('month', CURRENT_DATE)) AS nguoi_dung_moi_thang_nay,
    
    -- Người dùng mới tháng trước
    (SELECT COUNT(*) FROM nguoi_dung 
     WHERE DATE_TRUNC('month', ngay_tao) = DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 month')) AS nguoi_dung_moi_thang_truoc,
    
    -- Booking tháng này
    (SELECT COUNT(*) FROM dat_cho 
     WHERE DATE_TRUNC('month', ngay_dat) = DATE_TRUNC('month', CURRENT_DATE)) AS booking_thang_nay,
    
    -- Booking tháng trước
    (SELECT COUNT(*) FROM dat_cho 
     WHERE DATE_TRUNC('month', ngay_dat) = DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 month')) AS booking_thang_truoc;

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

-- name: GetUserGrowthByMonth :many
-- Tăng trưởng người dùng theo tháng (12 tháng gần nhất)
SELECT 
    DATE_TRUNC('month', ngay_tao) AS thang,
    COUNT(*) AS tong_dang_ky,
    COUNT(*) FILTER (WHERE vai_tro = 'khach_hang') AS khach_hang_moi,
    COUNT(*) FILTER (WHERE vai_tro = 'nha_cung_cap') AS nha_cung_cap_moi
FROM nguoi_dung
WHERE ngay_tao >= NOW() - INTERVAL '12 months'
GROUP BY DATE_TRUNC('month', ngay_tao)
ORDER BY thang ASC;

-- name: GetUserGrowthByDay :many
-- Tăng trưởng người dùng theo ngày (30 ngày gần nhất)
SELECT 
    DATE(ngay_tao) AS ngay,
    COUNT(*) AS so_luong
FROM nguoi_dung
WHERE ngay_tao >= NOW() - INTERVAL '30 days'
GROUP BY DATE(ngay_tao)
ORDER BY ngay ASC;

-- name: GetNewUsersToday :one
-- Số người dùng mới hôm nay
SELECT 
    COUNT(*) AS nguoi_dung_moi_hom_nay,
    COUNT(*) FILTER (WHERE vai_tro = 'khach_hang') AS khach_hang_moi,
    COUNT(*) FILTER (WHERE vai_tro = 'nha_cung_cap') AS nha_cung_cap_moi
FROM nguoi_dung
WHERE DATE(ngay_tao) = CURRENT_DATE;

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
LIMIT 10;

-- =====================
-- 3. TOUR STATISTICS
-- =====================

-- name: GetTourStatsByCategory :many
-- Thống kê tour theo danh mục
SELECT 
    dmt.id AS danh_muc_id,
    dmt.ten AS ten_danh_muc,
    COUNT(t.id) AS tong_tour,
    COUNT(t.id) FILTER (WHERE t.trang_thai = 'cong_bo') AS tour_cong_bo,
    COUNT(t.id) FILTER (WHERE t.noi_bat = TRUE) AS tour_noi_bat,
    COALESCE(AVG(t.gia_nguoi_lon), 0) AS gia_trung_binh
FROM danh_muc_tour dmt
LEFT JOIN tour t ON dmt.id = t.danh_muc_id AND t.dang_hoat_dong = TRUE
GROUP BY dmt.id, dmt.ten
ORDER BY tong_tour DESC;

-- name: GetTourStatsByStatus :many
-- Thống kê tour theo trạng thái
SELECT 
    trang_thai,
    COUNT(*) AS so_luong
FROM tour
WHERE dang_hoat_dong = TRUE
GROUP BY trang_thai
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

-- name: GetToursCreatedByMonth :many
-- Số tour mới theo tháng
SELECT 
    DATE_TRUNC('month', ngay_tao) AS thang,
    COUNT(*) AS so_luong
FROM tour
WHERE ngay_tao >= NOW() - INTERVAL '12 months'
GROUP BY DATE_TRUNC('month', ngay_tao)
ORDER BY thang ASC;

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

-- =====================
-- 4. BOOKING STATISTICS
-- =====================

-- name: GetBookingStatsByStatus :many
-- Thống kê booking theo trạng thái
SELECT 
    trang_thai,
    COUNT(*) AS so_luong,
    COALESCE(SUM(tong_tien), 0) AS tong_tien
FROM dat_cho
GROUP BY trang_thai
ORDER BY so_luong DESC;

-- name: GetRevenueByDay :many
-- Doanh thu theo ngày (30 ngày gần nhất)
SELECT 
    DATE(ngay_dat) AS ngay,
    COUNT(*) AS so_booking,
    COALESCE(SUM(tong_tien), 0) AS doanh_thu
FROM dat_cho
WHERE ngay_dat >= NOW() - INTERVAL '30 days'
  AND trang_thai IN ('da_thanh_toan', 'hoan_thanh')
GROUP BY DATE(ngay_dat)
ORDER BY ngay ASC;

-- name: GetRevenueByMonth :many
-- Doanh thu theo tháng (12 tháng gần nhất)
SELECT 
    DATE_TRUNC('month', ngay_dat) AS thang,
    COUNT(*) AS so_booking,
    COALESCE(SUM(tong_tien), 0) AS doanh_thu,
    COALESCE(AVG(tong_tien), 0) AS trung_binh_booking
FROM dat_cho
WHERE ngay_dat >= NOW() - INTERVAL '12 months'
  AND trang_thai IN ('da_thanh_toan', 'hoan_thanh')
GROUP BY DATE_TRUNC('month', ngay_dat)
ORDER BY thang ASC;

-- name: GetRevenueByYear :many
-- Doanh thu theo năm
SELECT 
    EXTRACT(YEAR FROM ngay_dat) AS nam,
    COUNT(*) AS so_booking,
    COALESCE(SUM(tong_tien), 0) AS doanh_thu
FROM dat_cho
WHERE trang_thai IN ('da_thanh_toan', 'hoan_thanh')
GROUP BY EXTRACT(YEAR FROM ngay_dat)
ORDER BY nam ASC;

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

-- name: GetBookingConversionRate :one
-- Tỷ lệ chuyển đổi booking
SELECT 
    COUNT(*) AS tong_booking,
    COUNT(*) FILTER (WHERE trang_thai = 'cho_xac_nhan') AS dang_cho,
    COUNT(*) FILTER (WHERE trang_thai = 'da_xac_nhan') AS da_xac_nhan,
    COUNT(*) FILTER (WHERE trang_thai = 'da_thanh_toan') AS da_thanh_toan,
    COUNT(*) FILTER (WHERE trang_thai = 'hoan_thanh') AS hoan_thanh,
    COUNT(*) FILTER (WHERE trang_thai = 'da_huy') AS da_huy,
    ROUND(
        COUNT(*) FILTER (WHERE trang_thai IN ('da_thanh_toan', 'hoan_thanh'))::DECIMAL / 
        NULLIF(COUNT(*), 0) * 100, 2
    ) AS ty_le_thanh_cong
FROM dat_cho;

-- =====================
-- 5. PAYMENT STATISTICS
-- =====================

-- name: GetPaymentStatsByMethod :many
-- Thống kê thanh toán theo phương thức
SELECT 
    phuong_thuc_thanh_toan,
    COUNT(*) AS so_giao_dich,
    COALESCE(SUM(tong_tien), 0) AS tong_tien
FROM dat_cho
WHERE phuong_thuc_thanh_toan IS NOT NULL
  AND trang_thai IN ('da_thanh_toan', 'hoan_thanh')
GROUP BY phuong_thuc_thanh_toan
ORDER BY tong_tien DESC;

-- name: GetTransactionStatsByStatus :many
-- Thống kê giao dịch theo trạng thái
SELECT 
    trang_thai,
    COUNT(*) AS so_giao_dich,
    COALESCE(SUM(so_tien), 0) AS tong_tien
FROM lich_su_giao_dich
GROUP BY trang_thai
ORDER BY so_giao_dich DESC;

-- name: GetTransactionsByDay :many
-- Giao dịch theo ngày (30 ngày gần nhất)
SELECT 
    DATE(ngay_tao) AS ngay,
    COUNT(*) AS so_giao_dich,
    COALESCE(SUM(so_tien), 0) AS tong_tien,
    COUNT(*) FILTER (WHERE trang_thai = 'thanh_cong') AS thanh_cong,
    COUNT(*) FILTER (WHERE trang_thai = 'that_bai') AS that_bai
FROM lich_su_giao_dich
WHERE ngay_tao >= NOW() - INTERVAL '30 days'
GROUP BY DATE(ngay_tao)
ORDER BY ngay ASC;

-- name: GetPaymentGatewayStats :many
-- Thống kê theo cổng thanh toán
SELECT 
    ctt.id AS cong_id,
    ctt.ten_hien_thi,
    COUNT(lsg.id) AS so_giao_dich,
    COALESCE(SUM(lsg.so_tien), 0) AS tong_tien,
    COUNT(*) FILTER (WHERE lsg.trang_thai = 'thanh_cong') AS thanh_cong,
    COUNT(*) FILTER (WHERE lsg.trang_thai = 'that_bai') AS that_bai
FROM cong_thanh_toan ctt
LEFT JOIN lich_su_giao_dich lsg ON ctt.id = lsg.cong_thanh_toan_id
GROUP BY ctt.id, ctt.ten_hien_thi
ORDER BY so_giao_dich DESC;

-- =====================
-- 6. SUPPLIER STATISTICS
-- =====================

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

-- =====================
-- 7. REVIEW STATISTICS
-- =====================

-- name: GetReviewStats :one
-- Thống kê đánh giá tổng quan
SELECT 
    COUNT(*) AS tong_danh_gia,
    COALESCE(AVG(diem_danh_gia), 0) AS diem_trung_binh,
    COUNT(*) FILTER (WHERE diem_danh_gia = 5) AS so_5_sao,
    COUNT(*) FILTER (WHERE diem_danh_gia = 4) AS so_4_sao,
    COUNT(*) FILTER (WHERE diem_danh_gia = 3) AS so_3_sao,
    COUNT(*) FILTER (WHERE diem_danh_gia = 2) AS so_2_sao,
    COUNT(*) FILTER (WHERE diem_danh_gia = 1) AS so_1_sao
FROM danh_gia
WHERE dang_hoat_dong = TRUE;

-- name: GetReviewDistribution :many
-- Phân bố đánh giá
SELECT 
    diem_danh_gia,
    COUNT(*) AS so_luong,
    ROUND(COUNT(*)::DECIMAL / NULLIF((SELECT COUNT(*) FROM danh_gia WHERE dang_hoat_dong = TRUE), 0) * 100, 2) AS phan_tram
FROM danh_gia
WHERE dang_hoat_dong = TRUE
GROUP BY diem_danh_gia
ORDER BY diem_danh_gia DESC;

-- name: GetReviewsByMonth :many
-- Đánh giá theo tháng
SELECT 
    DATE_TRUNC('month', ngay_tao) AS thang,
    COUNT(*) AS so_luong,
    COALESCE(AVG(diem_danh_gia), 0) AS diem_trung_binh
FROM danh_gia
WHERE ngay_tao >= NOW() - INTERVAL '12 months'
  AND dang_hoat_dong = TRUE
GROUP BY DATE_TRUNC('month', ngay_tao)
ORDER BY thang ASC;

-- name: GetRecentReviews :many
-- Đánh giá gần đây
SELECT 
    dg.id,
    dg.diem_danh_gia,
    dg.tieu_de,
    dg.noi_dung,
    dg.ngay_tao,
    nd.ho_ten AS ten_nguoi_danh_gia,
    nd.email AS email_nguoi_danh_gia,
    t.tieu_de AS ten_tour
FROM danh_gia dg
JOIN nguoi_dung nd ON dg.nguoi_dung_id = nd.id
JOIN tour t ON dg.tour_id = t.id
WHERE dg.dang_hoat_dong = TRUE
ORDER BY dg.ngay_tao DESC
LIMIT $1;

-- name: GetToursWithLowestRating :many
-- Tours có điểm đánh giá thấp nhất (cần cải thiện)
SELECT 
    t.id,
    t.tieu_de,
    ncc.ten AS ten_nha_cung_cap,
    COUNT(dg.id) AS so_danh_gia,
    COALESCE(AVG(dg.diem_danh_gia), 0) AS diem_trung_binh
FROM tour t
LEFT JOIN danh_gia dg ON t.id = dg.tour_id AND dg.dang_hoat_dong = TRUE
LEFT JOIN nha_cung_cap ncc ON t.nha_cung_cap_id = ncc.id
WHERE t.dang_hoat_dong = TRUE
GROUP BY t.id, t.tieu_de, ncc.ten
HAVING COUNT(dg.id) > 0
ORDER BY diem_trung_binh ASC
LIMIT $1;

-- =====================
-- 8. DESTINATION STATISTICS
-- =====================

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

-- name: GetDepartureCapacityStats :one
-- Thống kê công suất khởi hành
SELECT 
    COUNT(*) AS tong_khoi_hanh,
    SUM(suc_chua) AS tong_suc_chua,
    SUM(so_cho_da_dat) AS tong_da_dat,
    ROUND(
        SUM(so_cho_da_dat)::DECIMAL / NULLIF(SUM(suc_chua), 0) * 100, 2
    ) AS ty_le_lap_day
FROM khoi_hanh_tour
WHERE ngay_khoi_hanh >= CURRENT_DATE
  AND trang_thai IN ('len_lich', 'xac_nhan');

-- name: GetDeparturesByMonth :many
-- Khởi hành theo tháng
SELECT 
    DATE_TRUNC('month', ngay_khoi_hanh) AS thang,
    COUNT(*) AS so_khoi_hanh,
    SUM(suc_chua) AS tong_suc_chua,
    SUM(so_cho_da_dat) AS tong_da_dat
FROM khoi_hanh_tour
WHERE ngay_khoi_hanh >= NOW() - INTERVAL '6 months'
  AND ngay_khoi_hanh <= NOW() + INTERVAL '6 months'
GROUP BY DATE_TRUNC('month', ngay_khoi_hanh)
ORDER BY thang ASC;

-- =====================
-- 10. NOTIFICATION & ACTIVITY
-- =====================

-- name: GetUnreadNotificationCount :one
-- Số thông báo chưa đọc (cho admin xem tổng quan)
SELECT 
    COUNT(*) AS tong_thong_bao_chua_doc
FROM thong_bao
WHERE da_doc = FALSE;

-- name: GetSystemActivity :many
-- Hoạt động hệ thống gần đây
SELECT 
    'booking' AS loai,
    dc.id AS id_doi_tuong,
    dc.ngay_dat AS thoi_gian,
    CONCAT('Booking mới #', dc.id, ' - ', t.tieu_de) AS mo_ta
FROM dat_cho dc
JOIN khoi_hanh_tour kh ON dc.khoi_hanh_id = kh.id
JOIN tour t ON kh.tour_id = t.id
WHERE dc.ngay_dat >= NOW() - INTERVAL '24 hours'

UNION ALL

SELECT 
    'user' AS loai,
    NULL AS id_doi_tuong,
    ngay_tao AS thoi_gian,
    CONCAT('Người dùng mới: ', ho_ten) AS mo_ta
FROM nguoi_dung
WHERE ngay_tao >= NOW() - INTERVAL '24 hours'

UNION ALL

SELECT 
    'review' AS loai,
    dg.id AS id_doi_tuong,
    dg.ngay_tao AS thoi_gian,
    CONCAT('Đánh giá mới ', dg.diem_danh_gia, ' sao cho tour: ', t.tieu_de) AS mo_ta
FROM danh_gia dg
JOIN tour t ON dg.tour_id = t.id
WHERE dg.ngay_tao >= NOW() - INTERVAL '24 hours'

ORDER BY thoi_gian DESC
LIMIT 50;

-- =====================
-- 11. FAVORITES STATISTICS
-- =====================

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

