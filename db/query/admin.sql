-- name: GetAdminSummary :one
SELECT
  (SELECT COUNT(*) FROM nguoi_dung)                             AS total_users,
  (SELECT COUNT(*) FROM nguoi_dung WHERE dang_hoat_dong = TRUE) AS active_users,
  (SELECT COUNT(*) FROM tour)                                   AS total_tours,
  (SELECT COUNT(*) FROM tour WHERE dang_hoat_dong = TRUE)       AS active_tours,
  (SELECT COUNT(*) FROM dat_cho_tour)                           AS total_bookings,
  COALESCE((SELECT SUM(so_tien) FROM thanh_toan WHERE trang_thai = 'thanh_cong'), 0) AS total_revenue,
  COALESCE((SELECT AVG(rating)::float FROM danh_gia), 0)        AS avg_rating;

-- name: GetRevenueByMonth :many
SELECT
  EXTRACT(MONTH FROM ngay_thanh_toan) ::int AS month,
  COALESCE(SUM(so_tien), 0)                 AS revenue
FROM thanh_toan
WHERE EXTRACT(YEAR FROM ngay_thanh_toan) = $1
  AND trang_thai = 'thanh_cong'
GROUP BY month
ORDER BY month;

-- name: GetRevenueByYear :many
SELECT
  EXTRACT(YEAR FROM ngay_thanh_toan) ::int AS year,
  COALESCE(SUM(so_tien), 0)                 AS revenue
FROM thanh_toan
WHERE EXTRACT(YEAR FROM ngay_thanh_toan) = $1
  AND trang_thai = 'thanh_cong'
GROUP BY year
ORDER BY year;
-- name: GetBookingsByStatus :many
SELECT
  trang_thai::text AS status,
  COUNT(*)::int    AS total
FROM dat_cho_tour
GROUP BY status
ORDER BY total DESC;

-- name: GetTopToursByBookings :many
SELECT
  t.id                         AS tour_id,
  t.tieu_de                    AS tour_title,
  COUNT(dc.id)::int            AS bookings
FROM dat_cho_tour dc
JOIN khoi_hanh_tour kh ON kh.id = dc.khoi_hanh_id
JOIN tour t          ON t.id  = kh.tour_id
GROUP BY t.id, t.tieu_de
ORDER BY bookings DESC, t.id ASC
LIMIT $1;

-- name: GetNewUsersByMonth :many
SELECT
  EXTRACT(MONTH FROM ngay_tao)::int AS month,
  COUNT(*)::int                     AS new_users
FROM nguoi_dung
WHERE EXTRACT(YEAR FROM ngay_tao) = sqlc.arg(year)::int
GROUP BY month
ORDER BY month;

-- name: GetBookingsByMonth :many
SELECT
  EXTRACT(MONTH FROM ngay_dat)::int AS month,
  COUNT(*)::int                     AS total_bookings
FROM dat_cho_tour
WHERE EXTRACT(YEAR FROM ngay_dat) = sqlc.arg(year)::int
GROUP BY month
ORDER BY month;

-- name: GetTopSuppliers :many
SELECT
  ncc.id                     AS supplier_id,
  ncc.ten                    AS supplier_name,
  COUNT(t.id)::int           AS total_tours,
  COUNT(dc.id)::int          AS total_bookings
FROM nha_cung_cap ncc
LEFT JOIN tour t ON t.nha_cung_cap_id = ncc.id
LEFT JOIN khoi_hanh_tour kh ON kh.tour_id = t.id
LEFT JOIN dat_cho_tour dc ON dc.khoi_hanh_id = kh.id
WHERE ncc.dang_hoat_dong = TRUE
GROUP BY ncc.id, ncc.ten
ORDER BY total_bookings DESC, total_tours DESC
LIMIT $1;

-- name: GetRevenueBySupplier :many
SELECT
  ncc.id                           AS supplier_id,
  ncc.ten                          AS supplier_name,
  COALESCE(SUM(tt.so_tien), 0)     AS total_revenue,
  COUNT(DISTINCT dc.id)::int       AS total_bookings
FROM nha_cung_cap ncc
LEFT JOIN tour t ON t.nha_cung_cap_id = ncc.id
LEFT JOIN khoi_hanh_tour kh ON kh.tour_id = t.id
LEFT JOIN dat_cho_tour dc ON dc.khoi_hanh_id = kh.id
LEFT JOIN thanh_toan tt ON tt.dat_cho_id = dc.id AND tt.trang_thai = 'thanh_cong'
WHERE ncc.dang_hoat_dong = TRUE
GROUP BY ncc.id, ncc.ten
ORDER BY total_revenue DESC
LIMIT $1;

-- name: GetToursByCategory :many
SELECT
  dm.id                      AS category_id,
  dm.ten                     AS category_name,
  COUNT(t.id)::int           AS total_tours,
  COUNT(dc.id)::int          AS total_bookings
FROM danh_muc_tour dm
LEFT JOIN tour t ON t.danh_muc_id = dm.id
LEFT JOIN khoi_hanh_tour kh ON kh.tour_id = t.id
LEFT JOIN dat_cho_tour dc ON dc.khoi_hanh_id = kh.id
WHERE dm.dang_hoat_dong = TRUE
GROUP BY dm.id, dm.ten
ORDER BY total_bookings DESC, total_tours DESC;

-- name: GetUpcomingDepartures :many
SELECT
  kh.id                      AS departure_id,
  kh.ngay_khoi_hanh          AS departure_date,
  t.tieu_de                  AS tour_title,
  kh.suc_chua                AS capacity,
  COUNT(dc.id)::int          AS booked,
  kh.suc_chua - COUNT(dc.id)::int AS available
FROM khoi_hanh_tour kh
JOIN tour t ON t.id = kh.tour_id
LEFT JOIN dat_cho_tour dc ON dc.khoi_hanh_id = kh.id 
  AND dc.trang_thai IN ('cho_xac_nhan', 'da_xac_nhan', 'da_thanh_toan')
WHERE kh.ngay_khoi_hanh >= CURRENT_DATE
  AND kh.trang_thai IN ('len_lich', 'xac_nhan')
GROUP BY kh.id, kh.ngay_khoi_hanh, t.tieu_de, kh.suc_chua
ORDER BY kh.ngay_khoi_hanh ASC
LIMIT $1;

-- name: GetTopCustomers :many
SELECT
  nd.id                                AS user_id,
  nd.ho_ten                            AS user_name,
  nd.email                             AS user_email,
  COUNT(dc.id)::int                    AS total_bookings,
  COALESCE(SUM(tt.so_tien), 0)         AS total_spent
FROM nguoi_dung nd
JOIN dat_cho_tour dc ON dc.nguoi_dung_id = nd.id
LEFT JOIN thanh_toan tt ON tt.dat_cho_id = dc.id AND tt.trang_thai = 'thanh_cong'
WHERE nd.dang_hoat_dong = TRUE
GROUP BY nd.id, nd.ho_ten, nd.email
ORDER BY total_spent DESC, total_bookings DESC
LIMIT $1;

-- name: GetReviewStatsByTour :many
SELECT
  t.id                           AS tour_id,
  t.tieu_de                      AS tour_title,
  COUNT(dg.id)::int              AS total_reviews,
  COALESCE(AVG(dg.rating), 0)::float AS avg_rating,
  COUNT(CASE WHEN dg.rating = 5 THEN 1 END)::int AS five_star,
  COUNT(CASE WHEN dg.rating = 4 THEN 1 END)::int AS four_star,
  COUNT(CASE WHEN dg.rating = 3 THEN 1 END)::int AS three_star,
  COUNT(CASE WHEN dg.rating = 2 THEN 1 END)::int AS two_star,
  COUNT(CASE WHEN dg.rating = 1 THEN 1 END)::int AS one_star
FROM tour t
LEFT JOIN danh_gia dg ON dg.tour_id = t.id AND dg.dang_hoat_dong = TRUE
WHERE t.dang_hoat_dong = TRUE
GROUP BY t.id, t.tieu_de
HAVING COUNT(dg.id) > 0
ORDER BY avg_rating DESC, total_reviews DESC
LIMIT $1;

-- name: GetRevenueByDateRange :one
SELECT
  COALESCE(SUM(so_tien), 0) AS total_revenue,
  COUNT(*)::int             AS total_transactions
FROM thanh_toan
WHERE ngay_thanh_toan >= $1
  AND ngay_thanh_toan <= $2
  AND trang_thai = 'thanh_cong';

-- name: GetUserGrowth :many
SELECT
  DATE(ngay_tao) AS date,
  COUNT(*)::int  AS new_users
FROM nguoi_dung
WHERE ngay_tao >= $1 AND ngay_tao <= $2
GROUP BY date
ORDER BY date ASC;

