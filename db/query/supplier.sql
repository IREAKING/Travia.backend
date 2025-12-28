-- name: CreateSupplier :one
insert into nha_cung_cap (id, ten, dia_chi, website, mo_ta, logo)
values ($1, $2, $3, $4, $5, $6)
returning *;

-- name: GetSupplierByID :one
SELECT nha_cung_cap.*, nguoi_dung.ho_ten, nguoi_dung.email, nguoi_dung.so_dien_thoai, nguoi_dung.ngay_tao, nguoi_dung.ngay_cap_nhat FROM nha_cung_cap
JOIN nguoi_dung ON nguoi_dung.id = nha_cung_cap.id
WHERE nha_cung_cap.id = $1 AND nguoi_dung.dang_hoat_dong = TRUE;


-- name: GetAllSuppliers :many
SELECT * FROM nha_cung_cap
JOIN nguoi_dung ON nguoi_dung.id = nha_cung_cap.id
WHERE nguoi_dung.dang_hoat_dong = TRUE
ORDER BY nguoi_dung.ngay_tao DESC;


-- name: GetAllSuppliersIncludingDeleted :many
SELECT * FROM nha_cung_cap
JOIN nguoi_dung ON nguoi_dung.id = nha_cung_cap.id
ORDER BY nguoi_dung.ngay_tao DESC;

-- name: GetSuppliersByStatus :many
SELECT * FROM nha_cung_cap
JOIN nguoi_dung ON nguoi_dung.id = nha_cung_cap.id
WHERE nguoi_dung.dang_hoat_dong = TRUE
ORDER BY nguoi_dung.ngay_tao DESC;

-- name: GetActiveSuppliers :many
SELECT * FROM nha_cung_cap
JOIN nguoi_dung ON nguoi_dung.id = nha_cung_cap.id
WHERE nguoi_dung.dang_hoat_dong = TRUE
ORDER BY nguoi_dung.ngay_tao DESC;

-- name: UpdateSupplier :one
UPDATE nha_cung_cap
SET
    ten = COALESCE(sqlc.narg(ten), nha_cung_cap.ten),
    dia_chi = COALESCE(sqlc.narg(dia_chi), nha_cung_cap.dia_chi),
    website = COALESCE(sqlc.narg(website), nha_cung_cap.website),
    mo_ta = COALESCE(sqlc.narg(mo_ta), nha_cung_cap.mo_ta),
    logo = COALESCE(sqlc.narg(logo), nha_cung_cap.logo),
    ngay_cap_nhat = CURRENT_TIMESTAMP
WHERE id = $1 AND nguoi_dung.dang_hoat_dong = TRUE
RETURNING *;

-- name: UpdateSupplierAndUser :one
BEGIN;

WITH updated_ncc AS (
    UPDATE nha_cung_cap
    SET
        ten = COALESCE(sqlc.narg(ten), nha_cung_cap.ten),
        dia_chi = COALESCE(sqlc.narg(dia_chi), nha_canh_cap.dia_chi),
        website = COALESCE(sqlc.narg(website), nha_cung_cap.website),
        mo_ta = COALESCE(sqlc.narg(mo_ta), nha_cung_cap.mo_ta),
        logo = COALESCE(sqlc.narg(logo), nha_cung_cap.logo),
        ngay_cap_nhat = CURRENT_TIMESTAMP
    WHERE id = $1 AND nguoi_dung.dang_hoat_dong = TRUE
    RETURNING id
),
updated_user AS (
    UPDATE nguoi_dung
    SET
        ho_ten = COALESCE(sqlc.narg(ho_ten), nguoi_dung.ho_ten),
        email = COALESCE(sqlc.narg(email), nguoi_dung.email),
        so_dien_thoai = COALESCE(sqlc.narg(so_dien_thoai), nguoi_dung.so_dien_thoai),
        ngay_cap_nhat = CURRENT_TIMESTAMP
    WHERE id = (SELECT id FROM updated_ncc)
      AND dang_hoat_dong = TRUE
    RETURNING *
)
SELECT * FROM updated_user;

COMMIT;


-- name: UpdateSupplierStatus :one
UPDATE nha_cung_cap
SET
    ngay_cap_nhat = CURRENT_TIMESTAMP
WHERE id = $1 AND nha_cung_cap.dang_hoat_dong = TRUE
RETURNING *;

-- name: SoftDeleteSupplier :exec
UPDATE nha_cung_cap
SET
    nha_cung_cap.dang_hoat_dong = FALSE,
    nha_cung_cap.ngay_cap_nhat = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: RestoreSupplier :one
UPDATE nha_cung_cap
SET
    nha_cung_cap.dang_hoat_dong = TRUE,
    nha_cung_cap.ngay_cap_nhat = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteSupplier :exec
DELETE FROM nha_cung_cap
WHERE id = $1 AND nha_cung_cap.dang_hoat_dong = TRUE;

-- name: GetSupplierWithTourCount :many
SELECT
    ncc.*,
    COUNT(t.id)::int AS total_tours,
    COUNT(CASE WHEN t.dang_hoat_dong = TRUE THEN 1 END)::int AS active_tours
FROM nha_cung_cap ncc
JOIN nguoi_dung ON nguoi_dung.id = nha_cung_cap.id
LEFT JOIN tour t ON t.nha_cung_cap_id = ncc.id
WHERE nguoi_dung.dang_hoat_dong = TRUE AND nguoi_dung.email = $1
GROUP BY ncc.id
ORDER BY total_tours DESC;

-- name: SearchSuppliers :many
SELECT * FROM nha_cung_cap
JOIN nguoi_dung ON nguoi_dung.id = nha_cung_cap.id
WHERE nguoi_dung.dang_hoat_dong = TRUE AND nguoi_dung.email = $1
    AND (
        ten ILIKE '%' || $1 || '%'
        OR email ILIKE '%' || $1 || '%'
        OR nguoi_dai_dien ILIKE '%' || $1 || '%'
        OR dia_chi ILIKE '%' || $1 || '%'
    )
ORDER BY ngay_tao DESC;

-- name: CountSuppliers :one
SELECT COUNT(*)::int FROM nha_cung_cap
JOIN nguoi_dung ON nguoi_dung.id = nha_cung_cap.id;

-- name: CountSuppliersByStatus :many
-- đếm số lượng nhà cung cấp theo trạng thái
SELECT
    nguoi_dung.dang_hoat_dong,
    COUNT(*)::int AS total
FROM nha_cung_cap
JOIN nguoi_dung ON nguoi_dung.id = nha_cung_cap.id
WHERE nguoi_dung.dang_hoat_dong = TRUE AND nguoi_dung.email = $1
GROUP BY nguoi_dung.dang_hoat_dong
ORDER BY nguoi_dung.dang_hoat_dong DESC;

-- name: BulkUpdateSupplierStatus :exec
-- cập nhật trạng thái nhà cung cấp hàng loạt
UPDATE nha_cung_cap
SET
    ngay_cap_nhat = CURRENT_TIMESTAMP
WHERE id = ANY($1::int[]) AND nguoi_dung.dang_hoat_dong = TRUE;

-- name: ApproveSupplier :one
-- phê duyệt nhà cung cấp
UPDATE nguoi_dung
SET 
    dang_hoat_dong = TRUE,
    xac_thuc = TRUE,
    ngay_cap_nhat = CURRENT_TIMESTAMP
WHERE id = $1 
    AND vai_tro = 'nha_cung_cap'
    AND dang_hoat_dong = FALSE
RETURNING *;

-- name: GetPendingSuppliers :many
-- lấy danh sách nhà cung cấp chờ phê duyệt
SELECT 
    nha_cung_cap.*,
    nguoi_dung.ho_ten,
    nguoi_dung.email,
    nguoi_dung.so_dien_thoai,
    nguoi_dung.ngay_tao,
    nguoi_dung.ngay_cap_nhat,
    nguoi_dung.dang_hoat_dong,
    nguoi_dung.xac_thuc
FROM nha_cung_cap
JOIN nguoi_dung ON nguoi_dung.id = nha_cung_cap.id
WHERE nguoi_dung.vai_tro = 'nha_cung_cap'
    AND (nguoi_dung.dang_hoat_dong = FALSE OR nguoi_dung.xac_thuc = FALSE)
ORDER BY nguoi_dung.ngay_tao DESC;

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

-- lấy danh sách tour của nhà cung cấp
-- name: GetMyTours :many
SELECT * FROM tour
WHERE nha_cung_cap_id = $1 
  AND dang_hoat_dong = TRUE 
  AND (sqlc.narg(trang_thai)::TEXT IS NULL OR sqlc.narg(trang_thai)::TEXT = '' OR trang_thai = sqlc.narg(trang_thai)::TEXT)
ORDER BY ngay_tao DESC
LIMIT $2 OFFSET $3;

-- name: UpdateTourStatus :one
UPDATE tour
SET
    trang_thai = sqlc.narg(trang_thai),
    ngay_cap_nhat = CURRENT_TIMESTAMP
WHERE id = $1 AND dang_hoat_dong = TRUE
RETURNING *;

-- name: GetSupplierById :one
SELECT nha_cung_cap.*, nguoi_dung.ho_ten, nguoi_dung.email, nguoi_dung.so_dien_thoai, nguoi_dung.ngay_tao, nguoi_dung.ngay_cap_nhat FROM nha_cung_cap
JOIN nguoi_dung ON nguoi_dung.id = nha_cung_cap.id
WHERE nha_cung_cap.id = $1 AND nguoi_dung.dang_hoat_dong = TRUE;

-- ===========================================
-- DASHBOARD SUPPLIER - TRUY VẤN NÂNG CAO
-- ===========================================

-- name: GetSupplierDashboardOverview :one
-- Tổng quan dashboard: tổng doanh thu, số booking, số tour, đánh giá trung bình
SELECT 
    -- Tổng số tour
    COUNT(DISTINCT t.id)::int AS total_tours,
    COUNT(DISTINCT CASE WHEN t.trang_thai = 'cong_bo' THEN t.id END)::int AS published_tours,
    COUNT(DISTINCT CASE WHEN t.trang_thai = 'nhap' THEN t.id END)::int AS draft_tours,
    COUNT(DISTINCT CASE WHEN t.trang_thai = 'luu_tru' THEN t.id END)::int AS archived_tours,
    
    -- Tổng số booking
    COUNT(DISTINCT dc.id)::int AS total_bookings,
    COUNT(DISTINCT CASE WHEN dc.trang_thai = 'cho_xac_nhan' THEN dc.id END)::int AS pending_bookings,
    COUNT(DISTINCT CASE WHEN dc.trang_thai = 'da_xac_nhan' THEN dc.id END)::int AS confirmed_bookings,
    COUNT(DISTINCT CASE WHEN dc.trang_thai = 'da_thanh_toan' THEN dc.id END)::int AS paid_bookings,
    COUNT(DISTINCT CASE WHEN dc.trang_thai = 'hoan_thanh' THEN dc.id END)::int AS completed_bookings,
    COUNT(DISTINCT CASE WHEN dc.trang_thai = 'da_huy' THEN dc.id END)::int AS cancelled_bookings,
    
    -- Doanh thu
    COALESCE(SUM(dc.tong_tien) FILTER (WHERE dc.trang_thai IN ('da_thanh_toan', 'hoan_thanh')), 0)::numeric AS total_revenue,
    COALESCE(SUM(dc.tong_tien) FILTER (WHERE dc.trang_thai IN ('da_thanh_toan', 'hoan_thanh') AND dc.ngay_dat >= CURRENT_DATE - INTERVAL '30 days'), 0)::numeric AS revenue_last_30_days,
    COALESCE(SUM(dc.tong_tien) FILTER (WHERE dc.trang_thai IN ('da_thanh_toan', 'hoan_thanh') AND dc.ngay_dat >= CURRENT_DATE - INTERVAL '7 days'), 0)::numeric AS revenue_last_7_days,
    
    -- Đánh giá
    COALESCE(AVG(dg.diem_danh_gia), 0)::float AS avg_rating,
    COUNT(DISTINCT dg.id)::int AS total_reviews,
    
    -- Khách hàng
    COUNT(DISTINCT dc.nguoi_dung_id)::int AS total_customers,
    
    -- Tỷ lệ hủy
    CASE 
        WHEN COUNT(DISTINCT dc.id) > 0 THEN 
            (COUNT(DISTINCT CASE WHEN dc.trang_thai = 'da_huy' THEN dc.id END)::float / COUNT(DISTINCT dc.id)::float * 100)
        ELSE 0 
    END::float AS cancellation_rate
FROM nha_cung_cap ncc
LEFT JOIN tour t ON t.nha_cung_cap_id = ncc.id AND t.dang_hoat_dong = TRUE
LEFT JOIN khoi_hanh_tour kh ON kh.tour_id = t.id
LEFT JOIN dat_cho dc ON dc.khoi_hanh_id = kh.id
LEFT JOIN danh_gia dg ON dg.tour_id = t.id AND dg.dang_hoat_dong = TRUE
WHERE ncc.id = $1;

-- name: GetSupplierRevenueByTimeRange :many
-- Doanh thu theo khoảng thời gian (ngày, tuần, tháng)
SELECT 
    DATE_TRUNC($2::text, dc.ngay_dat)::timestamp AS period,
    COUNT(DISTINCT dc.id)::int AS booking_count,
    COALESCE(SUM(dc.tong_tien) FILTER (WHERE dc.trang_thai IN ('da_thanh_toan', 'hoan_thanh')), 0)::numeric AS revenue,
    COALESCE(SUM(dc.so_nguoi_lon + dc.so_tre_em) FILTER (WHERE dc.trang_thai IN ('da_thanh_toan', 'hoan_thanh')), 0)::int AS total_passengers
FROM nha_cung_cap ncc
JOIN tour t ON t.nha_cung_cap_id = ncc.id AND t.dang_hoat_dong = TRUE
JOIN khoi_hanh_tour kh ON kh.tour_id = t.id
JOIN dat_cho dc ON dc.khoi_hanh_id = kh.id
WHERE ncc.id = $1
    AND dc.ngay_dat >= COALESCE($3::timestamp, CURRENT_DATE - INTERVAL '30 days')
    AND dc.ngay_dat <= COALESCE($4::timestamp, CURRENT_DATE)
    AND dc.trang_thai IN ('da_thanh_toan', 'hoan_thanh', 'cho_xac_nhan', 'da_xac_nhan')
GROUP BY DATE_TRUNC($2::text, dc.ngay_dat)
ORDER BY period ASC;

-- name: GetSupplierTopTours :many
-- Top tours bán chạy nhất theo số lượng booking và doanh thu
SELECT 
    t.id,
    t.tieu_de,
    t.gia_nguoi_lon,
    t.gia_tre_em,
    t.don_vi_tien_te,
    t.trang_thai,
    (
        SELECT a.duong_dan
        FROM anh_tour a
        WHERE a.tour_id = t.id AND a.la_anh_chinh = TRUE
        LIMIT 1
    ) AS anh_chinh,
    COUNT(DISTINCT dc.id)::int AS total_bookings,
    COALESCE(SUM(dc.tong_tien) FILTER (WHERE dc.trang_thai IN ('da_thanh_toan', 'hoan_thanh')), 0)::numeric AS total_revenue,
    COALESCE(SUM(dc.so_nguoi_lon + dc.so_tre_em) FILTER (WHERE dc.trang_thai IN ('da_thanh_toan', 'hoan_thanh')), 0)::int AS total_passengers,
    COALESCE(AVG(dg.diem_danh_gia), 0)::float AS avg_rating,
    COUNT(DISTINCT dg.id)::int AS total_reviews
FROM nha_cung_cap ncc
JOIN tour t ON t.nha_cung_cap_id = ncc.id AND t.dang_hoat_dong = TRUE
LEFT JOIN khoi_hanh_tour kh ON kh.tour_id = t.id
LEFT JOIN dat_cho dc ON dc.khoi_hanh_id = kh.id
LEFT JOIN danh_gia dg ON dg.tour_id = t.id AND dg.dang_hoat_dong = TRUE
WHERE ncc.id = $1
    AND ($2::timestamp IS NULL OR dc.ngay_dat >= $2)
    AND ($3::timestamp IS NULL OR dc.ngay_dat <= $3)
GROUP BY t.id, t.tieu_de, t.gia_nguoi_lon, t.gia_tre_em, t.don_vi_tien_te, t.trang_thai
ORDER BY 
    CASE WHEN $4::text = 'revenue' THEN COALESCE(SUM(dc.tong_tien) FILTER (WHERE dc.trang_thai IN ('da_thanh_toan', 'hoan_thanh')), 0) END DESC,
    CASE WHEN $4::text = 'bookings' THEN COUNT(DISTINCT dc.id) END DESC,
    CASE WHEN $4::text = 'rating' THEN COALESCE(AVG(dg.diem_danh_gia), 0) END DESC,
    total_bookings DESC
LIMIT $5;

-- name: GetSupplierBookingStatsByStatus :many
-- Thống kê booking theo trạng thái và thời gian
SELECT 
    dc.trang_thai,
    COUNT(DISTINCT dc.id)::int AS booking_count,
    COALESCE(SUM(dc.tong_tien), 0)::numeric AS total_amount,
    COALESCE(SUM(dc.so_nguoi_lon + dc.so_tre_em), 0)::int AS total_passengers,
    MIN(dc.ngay_dat)::timestamp AS first_booking_date,
    MAX(dc.ngay_dat)::timestamp AS last_booking_date
FROM nha_cung_cap ncc
JOIN tour t ON t.nha_cung_cap_id = ncc.id AND t.dang_hoat_dong = TRUE
JOIN khoi_hanh_tour kh ON kh.tour_id = t.id
JOIN dat_cho dc ON dc.khoi_hanh_id = kh.id
WHERE ncc.id = $1
    AND ($2::timestamp IS NULL OR dc.ngay_dat >= $2)
    AND ($3::timestamp IS NULL OR dc.ngay_dat <= $3)
GROUP BY dc.trang_thai
ORDER BY booking_count DESC;

-- name: GetSupplierTourStatsByStatus :many
-- Thống kê tour theo trạng thái
SELECT 
    t.trang_thai,
    COUNT(DISTINCT t.id)::int AS tour_count,
    COUNT(DISTINCT kh.id)::int AS total_departures,
    COUNT(DISTINCT dc.id)::int AS total_bookings,
    COALESCE(SUM(dc.tong_tien) FILTER (WHERE dc.trang_thai IN ('da_thanh_toan', 'hoan_thanh')), 0)::numeric AS total_revenue
FROM nha_cung_cap ncc
JOIN tour t ON t.nha_cung_cap_id = ncc.id AND t.dang_hoat_dong = TRUE
LEFT JOIN khoi_hanh_tour kh ON kh.tour_id = t.id
LEFT JOIN dat_cho dc ON dc.khoi_hanh_id = kh.id
WHERE ncc.id = $1
GROUP BY t.trang_thai
ORDER BY tour_count DESC;

-- name: GetSupplierRevenueChart :many
-- Biểu đồ doanh thu theo thời gian (cho chart)
SELECT 
    DATE_TRUNC($2::text, dc.ngay_dat) AS date,
    COALESCE(SUM(dc.tong_tien) FILTER (WHERE dc.trang_thai IN ('da_thanh_toan', 'hoan_thanh')), 0)::numeric AS revenue,
    COUNT(DISTINCT dc.id) FILTER (WHERE dc.trang_thai IN ('da_thanh_toan', 'hoan_thanh'))::int AS booking_count,
    COUNT(DISTINCT dc.nguoi_dung_id) FILTER (WHERE dc.trang_thai IN ('da_thanh_toan', 'hoan_thanh'))::int AS customer_count
FROM nha_cung_cap ncc
JOIN tour t ON t.nha_cung_cap_id = ncc.id AND t.dang_hoat_dong = TRUE
JOIN khoi_hanh_tour kh ON kh.tour_id = t.id
JOIN dat_cho dc ON dc.khoi_hanh_id = kh.id
WHERE ncc.id = $1
    AND dc.ngay_dat >= COALESCE($3::timestamp, CURRENT_DATE - INTERVAL '30 days')
    AND dc.ngay_dat <= COALESCE($4::timestamp, CURRENT_DATE)
    AND dc.trang_thai IN ('da_thanh_toan', 'hoan_thanh')
GROUP BY DATE_TRUNC($2::text, dc.ngay_dat)
ORDER BY date ASC;

-- name: GetSupplierCustomerStats :many
-- Thống kê khách hàng: top khách hàng, số lần đặt, tổng tiền
SELECT 
    nd.id AS customer_id,
    nd.ho_ten AS customer_name,
    nd.email AS customer_email,
    COUNT(DISTINCT dc.id)::int AS total_bookings,
    COALESCE(SUM(dc.tong_tien) FILTER (WHERE dc.trang_thai IN ('da_thanh_toan', 'hoan_thanh')), 0)::numeric AS total_spent,
    COALESCE(SUM(dc.so_nguoi_lon + dc.so_tre_em) FILTER (WHERE dc.trang_thai IN ('da_thanh_toan', 'hoan_thanh')), 0)::int AS total_passengers,
    MIN(dc.ngay_dat)::timestamp AS first_booking_date,
    MAX(dc.ngay_dat)::timestamp AS last_booking_date
FROM nha_cung_cap ncc
JOIN tour t ON t.nha_cung_cap_id = ncc.id AND t.dang_hoat_dong = TRUE
JOIN khoi_hanh_tour kh ON kh.tour_id = t.id
JOIN dat_cho dc ON dc.khoi_hanh_id = kh.id
JOIN nguoi_dung nd ON nd.id = dc.nguoi_dung_id
WHERE ncc.id = $1
    AND ($2::timestamp IS NULL OR dc.ngay_dat >= $2)
    AND ($3::timestamp IS NULL OR dc.ngay_dat <= $3)
GROUP BY nd.id, nd.ho_ten, nd.email
ORDER BY 
    CASE WHEN $4::text = 'spent' THEN COALESCE(SUM(dc.tong_tien) FILTER (WHERE dc.trang_thai IN ('da_thanh_toan', 'hoan_thanh')), 0) END DESC,
    CASE WHEN $4::text = 'bookings' THEN COUNT(DISTINCT dc.id) END DESC,
    total_spent DESC
LIMIT $5;

-- name: GetSupplierCancellationAnalysis :one
-- Phân tích tỷ lệ hủy booking
SELECT 
    COUNT(DISTINCT dc.id)::int AS total_bookings,
    COUNT(DISTINCT CASE WHEN dc.trang_thai = 'da_huy' THEN dc.id END)::int AS cancelled_bookings,
    CASE 
        WHEN COUNT(DISTINCT dc.id) > 0 THEN 
            (COUNT(DISTINCT CASE WHEN dc.trang_thai = 'da_huy' THEN dc.id END)::float / COUNT(DISTINCT dc.id)::float * 100)
        ELSE 0 
    END::float AS cancellation_rate,
    COALESCE(SUM(dc.tong_tien) FILTER (WHERE dc.trang_thai = 'da_huy'), 0)::numeric AS lost_revenue,
    COUNT(DISTINCT CASE WHEN dc.trang_thai = 'da_huy' AND dc.ngay_dat >= CURRENT_DATE - INTERVAL '30 days' THEN dc.id END)::int AS cancelled_last_30_days
FROM nha_cung_cap ncc
JOIN tour t ON t.nha_cung_cap_id = ncc.id AND t.dang_hoat_dong = TRUE
JOIN khoi_hanh_tour kh ON kh.tour_id = t.id
JOIN dat_cho dc ON dc.khoi_hanh_id = kh.id
WHERE ncc.id = $1
    AND ($2::timestamp IS NULL OR dc.ngay_dat >= $2)
    AND ($3::timestamp IS NULL OR dc.ngay_dat <= $3);

-- name: GetSupplierRatingAnalysis :one
-- Phân tích đánh giá tour
SELECT 
    COUNT(DISTINCT dg.id)::int AS total_reviews,
    COALESCE(AVG(dg.diem_danh_gia), 0)::float AS avg_rating,
    COUNT(DISTINCT CASE WHEN dg.diem_danh_gia = 5 THEN dg.id END)::int AS five_star_count,
    COUNT(DISTINCT CASE WHEN dg.diem_danh_gia = 4 THEN dg.id END)::int AS four_star_count,
    COUNT(DISTINCT CASE WHEN dg.diem_danh_gia = 3 THEN dg.id END)::int AS three_star_count,
    COUNT(DISTINCT CASE WHEN dg.diem_danh_gia = 2 THEN dg.id END)::int AS two_star_count,
    COUNT(DISTINCT CASE WHEN dg.diem_danh_gia = 1 THEN dg.id END)::int AS one_star_count,
    COUNT(DISTINCT t.id)::int AS tours_with_reviews
FROM nha_cung_cap ncc
JOIN tour t ON t.nha_cung_cap_id = ncc.id AND t.dang_hoat_dong = TRUE
LEFT JOIN danh_gia dg ON dg.tour_id = t.id AND dg.dang_hoat_dong = TRUE
WHERE ncc.id = $1;

-- name: GetSupplierUpcomingDepartures :many
-- Lịch khởi hành sắp tới
SELECT 
    kh.id AS departure_id,
    kh.ngay_khoi_hanh,
    kh.ngay_ket_thuc,
    kh.suc_chua,
    kh.so_cho_da_dat,
    (kh.suc_chua - kh.so_cho_da_dat) AS available_seats,
    kh.trang_thai,
    t.id AS tour_id,
    t.tieu_de AS tour_title,
    (
        SELECT a.duong_dan
        FROM anh_tour a
        WHERE a.tour_id = t.id AND a.la_anh_chinh = TRUE
        LIMIT 1
    ) AS tour_image,
    COUNT(DISTINCT dc.id)::int AS booking_count,
    COALESCE(SUM(dc.tong_tien) FILTER (WHERE dc.trang_thai IN ('da_thanh_toan', 'hoan_thanh')), 0)::numeric AS revenue
FROM nha_cung_cap ncc
JOIN tour t ON t.nha_cung_cap_id = ncc.id AND t.dang_hoat_dong = TRUE
JOIN khoi_hanh_tour kh ON kh.tour_id = t.id
LEFT JOIN dat_cho dc ON dc.khoi_hanh_id = kh.id
WHERE ncc.id = $1
    AND kh.ngay_khoi_hanh >= CURRENT_DATE
    AND kh.trang_thai IN ('len_lich', 'xac_nhan', 'con_cho')
GROUP BY kh.id, kh.ngay_khoi_hanh, kh.ngay_ket_thuc, kh.suc_chua, kh.so_cho_da_dat, kh.trang_thai, t.id, t.tieu_de
ORDER BY kh.ngay_khoi_hanh ASC
LIMIT $2;

-- name: GetSupplierRecentBookings :many
-- Booking gần đây
SELECT 
    dc.id AS booking_id,
    dc.ngay_dat,
    dc.trang_thai,
    dc.tong_tien,
    dc.don_vi_tien_te,
    dc.so_nguoi_lon,
    dc.so_tre_em,
    nd.ho_ten AS customer_name,
    nd.email AS customer_email,
    t.id AS tour_id,
    t.tieu_de AS tour_title,
    kh.ngay_khoi_hanh
FROM nha_cung_cap ncc
JOIN tour t ON t.nha_cung_cap_id = ncc.id AND t.dang_hoat_dong = TRUE
JOIN khoi_hanh_tour kh ON kh.tour_id = t.id
JOIN dat_cho dc ON dc.khoi_hanh_id = kh.id
JOIN nguoi_dung nd ON nd.id = dc.nguoi_dung_id
WHERE ncc.id = $1
ORDER BY dc.ngay_dat DESC
LIMIT $2;

-- name: GetSupplierMonthlyComparison :one
-- So sánh tháng hiện tại với tháng trước
SELECT 
    -- Tháng hiện tại
    COUNT(DISTINCT dc.id) FILTER (WHERE DATE_TRUNC('month', dc.ngay_dat) = DATE_TRUNC('month', CURRENT_DATE))::int AS current_month_bookings,
    COALESCE(SUM(dc.tong_tien) FILTER (WHERE dc.trang_thai IN ('da_thanh_toan', 'hoan_thanh') AND DATE_TRUNC('month', dc.ngay_dat) = DATE_TRUNC('month', CURRENT_DATE)), 0)::numeric AS current_month_revenue,
    
    -- Tháng trước
    COUNT(DISTINCT dc.id) FILTER (WHERE DATE_TRUNC('month', dc.ngay_dat) = DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 month'))::int AS previous_month_bookings,
    COALESCE(SUM(dc.tong_tien) FILTER (WHERE dc.trang_thai IN ('da_thanh_toan', 'hoan_thanh') AND DATE_TRUNC('month', dc.ngay_dat) = DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 month')), 0)::numeric AS previous_month_revenue,
    
    -- Tính phần trăm thay đổi
    CASE 
        WHEN COUNT(DISTINCT dc.id) FILTER (WHERE DATE_TRUNC('month', dc.ngay_dat) = DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 month')) > 0 THEN
            ((COUNT(DISTINCT dc.id) FILTER (WHERE DATE_TRUNC('month', dc.ngay_dat) = DATE_TRUNC('month', CURRENT_DATE))::float - 
              COUNT(DISTINCT dc.id) FILTER (WHERE DATE_TRUNC('month', dc.ngay_dat) = DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 month'))::float) /
             COUNT(DISTINCT dc.id) FILTER (WHERE DATE_TRUNC('month', dc.ngay_dat) = DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 month'))::float * 100)
        ELSE 0
    END::float AS booking_change_percent,
    
    CASE 
        WHEN COALESCE(SUM(dc.tong_tien) FILTER (WHERE dc.trang_thai IN ('da_thanh_toan', 'hoan_thanh') AND DATE_TRUNC('month', dc.ngay_dat) = DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 month')), 0) > 0 THEN
            ((COALESCE(SUM(dc.tong_tien) FILTER (WHERE dc.trang_thai IN ('da_thanh_toan', 'hoan_thanh') AND DATE_TRUNC('month', dc.ngay_dat) = DATE_TRUNC('month', CURRENT_DATE)), 0)::float - 
              COALESCE(SUM(dc.tong_tien) FILTER (WHERE dc.trang_thai IN ('da_thanh_toan', 'hoan_thanh') AND DATE_TRUNC('month', dc.ngay_dat) = DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 month')), 0)::float) /
             COALESCE(SUM(dc.tong_tien) FILTER (WHERE dc.trang_thai IN ('da_thanh_toan', 'hoan_thanh') AND DATE_TRUNC('month', dc.ngay_dat) = DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 month')), 0)::float * 100)
        ELSE 0
    END::float AS revenue_change_percent
FROM nha_cung_cap ncc
JOIN tour t ON t.nha_cung_cap_id = ncc.id AND t.dang_hoat_dong = TRUE
JOIN khoi_hanh_tour kh ON kh.tour_id = t.id
JOIN dat_cho dc ON dc.khoi_hanh_id = kh.id
WHERE ncc.id = $1;

-- name: GetSupplierBookingsByStatusAdvanced :many
-- Lấy danh sách đặt chỗ theo trạng thái với nhiều filter nâng cao
SELECT 
    dc.id AS booking_id,
    dc.ngay_dat,
    dc.trang_thai,
    dc.tong_tien,
    dc.don_vi_tien_te,
    dc.so_nguoi_lon,
    dc.so_tre_em,
    dc.phuong_thuc_thanh_toan,
    dc.ngay_cap_nhat,
    
    -- Thông tin khách hàng
    nd.id AS customer_id,
    nd.ho_ten AS customer_name,
    nd.email AS customer_email,
    nd.so_dien_thoai AS customer_phone,
    
    -- Thông tin tour
    t.id AS tour_id,
    t.tieu_de AS tour_title,
    t.gia_nguoi_lon AS tour_price_adult,
    t.gia_tre_em AS tour_price_child,
    t.don_vi_tien_te AS tour_currency,
    (
        SELECT a.duong_dan
        FROM anh_tour a
        WHERE a.tour_id = t.id AND a.la_anh_chinh = TRUE
        LIMIT 1
    ) AS tour_image,
    
    -- Thông tin khởi hành
    kh.id AS departure_id,
    kh.ngay_khoi_hanh,
    kh.ngay_ket_thuc,
    kh.suc_chua AS departure_capacity,
    kh.so_cho_da_dat AS departure_booked,
    (kh.suc_chua - kh.so_cho_da_dat) AS departure_available,
    kh.trang_thai AS departure_status,
    
    -- Thông tin hành khách
    (
        SELECT COUNT(*)
        FROM hanh_khach hk
        WHERE hk.dat_cho_id = dc.id
    )::int AS passenger_count,
    
    -- Thông tin thanh toán
    (
        SELECT COUNT(*)
        FROM lich_su_giao_dich lsgd
        WHERE lsgd.dat_cho_id = dc.id
            AND lsgd.trang_thai = 'thanh_cong'
    )::int AS successful_payments,
    
    -- Tính toán
    (dc.so_nguoi_lon + dc.so_tre_em) AS total_passengers,
    CASE 
        WHEN dc.trang_thai IN ('da_thanh_toan', 'hoan_thanh') THEN dc.tong_tien
        ELSE 0
    END::numeric AS confirmed_revenue
FROM nha_cung_cap ncc
JOIN tour t ON t.nha_cung_cap_id = ncc.id AND t.dang_hoat_dong = TRUE
JOIN khoi_hanh_tour kh ON kh.tour_id = t.id
JOIN dat_cho dc ON dc.khoi_hanh_id = kh.id
JOIN nguoi_dung nd ON nd.id = dc.nguoi_dung_id
WHERE ncc.id = $1
    -- Filter theo trạng thái
    AND (
        sqlc.narg('trang_thai')::text IS NULL 
        OR sqlc.narg('trang_thai')::text = ''
        OR dc.trang_thai::text = sqlc.narg('trang_thai')::text
    )
    -- Filter theo tour ID
    AND (sqlc.narg('tour_id')::int IS NULL OR t.id = sqlc.narg('tour_id')::int)
    -- Filter theo khoảng thời gian đặt chỗ
    AND (sqlc.narg('start_date')::timestamp IS NULL OR dc.ngay_dat >= sqlc.narg('start_date')::timestamp)
    AND (sqlc.narg('end_date')::timestamp IS NULL OR dc.ngay_dat <= sqlc.narg('end_date')::timestamp)
    -- Filter theo khoảng thời gian khởi hành
    AND (sqlc.narg('departure_start_date')::date IS NULL OR kh.ngay_khoi_hanh >= sqlc.narg('departure_start_date')::date)
    AND (sqlc.narg('departure_end_date')::date IS NULL OR kh.ngay_khoi_hanh <= sqlc.narg('departure_end_date')::date)
    -- Search theo tên khách hàng hoặc email
    AND (
        sqlc.narg('search_keyword')::text IS NULL 
        OR sqlc.narg('search_keyword')::text = ''
        OR nd.ho_ten ILIKE '%' || sqlc.narg('search_keyword')::text || '%'
        OR nd.email ILIKE '%' || sqlc.narg('search_keyword')::text || '%'
        OR t.tieu_de ILIKE '%' || sqlc.narg('search_keyword')::text || '%'
    )
    -- Filter theo phương thức thanh toán
    AND (
        sqlc.narg('phuong_thuc_thanh_toan')::text IS NULL 
        OR sqlc.narg('phuong_thuc_thanh_toan')::text = ''
        OR dc.phuong_thuc_thanh_toan::text = sqlc.narg('phuong_thuc_thanh_toan')::text
    )
    -- Filter theo khoảng giá
    AND (sqlc.narg('min_amount')::numeric IS NULL OR dc.tong_tien >= sqlc.narg('min_amount')::numeric)
    AND (sqlc.narg('max_amount')::numeric IS NULL OR dc.tong_tien <= sqlc.narg('max_amount')::numeric)
ORDER BY 
    CASE WHEN sqlc.narg('sort_by')::text = 'ngay_dat_asc' THEN dc.ngay_dat END ASC,
    CASE WHEN sqlc.narg('sort_by')::text = 'ngay_dat_desc' THEN dc.ngay_dat END DESC,
    CASE WHEN sqlc.narg('sort_by')::text = 'tong_tien_asc' THEN dc.tong_tien END ASC,
    CASE WHEN sqlc.narg('sort_by')::text = 'tong_tien_desc' THEN dc.tong_tien END DESC,
    CASE WHEN sqlc.narg('sort_by')::text = 'ngay_khoi_hanh_asc' THEN kh.ngay_khoi_hanh END ASC,
    CASE WHEN sqlc.narg('sort_by')::text = 'ngay_khoi_hanh_desc' THEN kh.ngay_khoi_hanh END DESC,
    dc.ngay_dat DESC
LIMIT $2 OFFSET $3;

-- name: CountSupplierBookingsByStatusAdvanced :one
-- Đếm tổng số booking theo các filter nâng cao
SELECT COUNT(*)::int AS total_count
FROM nha_cung_cap ncc
JOIN tour t ON t.nha_cung_cap_id = ncc.id AND t.dang_hoat_dong = TRUE
JOIN khoi_hanh_tour kh ON kh.tour_id = t.id
JOIN dat_cho dc ON dc.khoi_hanh_id = kh.id
JOIN nguoi_dung nd ON nd.id = dc.nguoi_dung_id
WHERE ncc.id = $1
    -- Filter theo trạng thái
    AND (
        sqlc.narg('trang_thai')::text IS NULL 
        OR sqlc.narg('trang_thai')::text = ''
        OR dc.trang_thai::text = sqlc.narg('trang_thai')::text
    )
    -- Filter theo tour ID
    AND (sqlc.narg('tour_id')::int IS NULL OR t.id = sqlc.narg('tour_id')::int)
    -- Filter theo khoảng thời gian đặt chỗ
    AND (sqlc.narg('start_date')::timestamp IS NULL OR dc.ngay_dat >= sqlc.narg('start_date')::timestamp)
    AND (sqlc.narg('end_date')::timestamp IS NULL OR dc.ngay_dat <= sqlc.narg('end_date')::timestamp)
    -- Filter theo khoảng thời gian khởi hành
    AND (sqlc.narg('departure_start_date')::date IS NULL OR kh.ngay_khoi_hanh >= sqlc.narg('departure_start_date')::date)
    AND (sqlc.narg('departure_end_date')::date IS NULL OR kh.ngay_khoi_hanh <= sqlc.narg('departure_end_date')::date)
    -- Search theo tên khách hàng hoặc email
    AND (
        sqlc.narg('search_keyword')::text IS NULL 
        OR sqlc.narg('search_keyword')::text = ''
        OR nd.ho_ten ILIKE '%' || sqlc.narg('search_keyword')::text || '%'
        OR nd.email ILIKE '%' || sqlc.narg('search_keyword')::text || '%'
        OR t.tieu_de ILIKE '%' || sqlc.narg('search_keyword')::text || '%'
    )
    -- Filter theo phương thức thanh toán
    AND (
        sqlc.narg('phuong_thuc_thanh_toan')::text IS NULL 
        OR sqlc.narg('phuong_thuc_thanh_toan')::text = ''
        OR dc.phuong_thuc_thanh_toan::text = sqlc.narg('phuong_thuc_thanh_toan')::text
    )
    -- Filter theo khoảng giá
    AND (sqlc.narg('min_amount')::numeric IS NULL OR dc.tong_tien >= sqlc.narg('min_amount')::numeric)
    AND (sqlc.narg('max_amount')::numeric IS NULL OR dc.tong_tien <= sqlc.narg('max_amount')::numeric);




