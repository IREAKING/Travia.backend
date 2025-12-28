-- ==================== DEPARTURE MANAGEMENT ====================

-- name: CreateDeparture :one
-- tạo lịch khởi hành
INSERT INTO khoi_hanh_tour (
    tour_id,
    ngay_khoi_hanh,
    ngay_ket_thuc,
    suc_chua,
    trang_thai,
    ghi_chu
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetDepartureByID :one
-- lấy thông tin chi tiết của một lịch khởi hành
SELECT 
    kh.*,
    t.tieu_de as ten_tour,
    t.gia_nguoi_lon as gia_nguoi_lon,
    t.gia_tre_em as gia_tre_em,
    t.don_vi_tien_te,
    COALESCE(SUM(dc.so_nguoi_lon + dc.so_tre_em), 0) as so_cho_da_dat,
    kh.suc_chua - COALESCE(SUM(dc.so_nguoi_lon + dc.so_tre_em), 0) as so_cho_con_trong
FROM khoi_hanh_tour kh
JOIN tour t ON kh.tour_id = t.id
LEFT JOIN dat_cho dc ON kh.id = dc.khoi_hanh_id 
    AND dc.trang_thai NOT IN ('da_huy')
WHERE kh.id = $1
GROUP BY kh.id, t.tieu_de, t.gia_nguoi_lon, t.gia_tre_em, t.don_vi_tien_te;

-- name: GetDeparturesByTour :many
-- lấy danh sách lịch khởi hành của một tour
SELECT 
    kh.*,
    COALESCE(SUM(dc.so_nguoi_lon + dc.so_tre_em), 0) as so_cho_da_dat,
    kh.suc_chua - COALESCE(SUM(dc.so_nguoi_lon + dc.so_tre_em), 0) as so_cho_con_trong
FROM khoi_hanh_tour kh
LEFT JOIN dat_cho dc ON kh.id = dc.khoi_hanh_id 
    AND dc.trang_thai NOT IN ('da_huy')
WHERE kh.tour_id = $1
GROUP BY kh.id
ORDER BY kh.ngay_khoi_hanh ASC;


-- name: GetUpcomingDeparturesList :many
SELECT 
    kh.*,
    t.tieu_de as ten_tour,
    t.gia_nguoi_lon,
    t.gia_tre_em,
    COALESCE(SUM(dc.so_nguoi_lon + dc.so_tre_em), 0) as so_cho_da_dat
FROM khoi_hanh_tour kh
JOIN tour t ON kh.tour_id = t.id
LEFT JOIN dat_cho dc ON kh.id = dc.khoi_hanh_id 
    AND dc.trang_thai NOT IN ('da_huy')
WHERE kh.ngay_khoi_hanh BETWEEN CURRENT_DATE AND (CURRENT_DATE + INTERVAL '30 days')
    AND kh.trang_thai IN ('len_lich', 'xac_nhan', 'con_cho')
GROUP BY kh.id, t.tieu_de, t.gia_nguoi_lon, t.gia_tre_em
ORDER BY kh.ngay_khoi_hanh ASC
LIMIT $1;

-- name: UpdateDeparture :one
UPDATE khoi_hanh_tour
SET 
    ngay_khoi_hanh = COALESCE(sqlc.narg('ngay_khoi_hanh'), ngay_khoi_hanh),
    ngay_ket_thuc = COALESCE(sqlc.narg('ngay_ket_thuc'), ngay_ket_thuc),
    suc_chua = COALESCE(sqlc.narg('suc_chua'), suc_chua),
    trang_thai = COALESCE(sqlc.narg('trang_thai'), trang_thai),
    ghi_chu = COALESCE(sqlc.narg('ghi_chu'), ghi_chu),
    ngay_cap_nhat = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteDeparture :exec
DELETE FROM khoi_hanh_tour
WHERE id = $1;

-- name: CancelDeparture :one
UPDATE khoi_hanh_tour
SET 
    trang_thai = 'huy',
    ngay_cap_nhat = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateDepartureCapacity :one
UPDATE khoi_hanh_tour
SET 
    suc_chua = $2,
    ngay_cap_nhat = NOW()
WHERE id = $1
RETURNING *;

-- name: GetDeparturesByDateRange :many
SELECT 
    kh.*,
    t.tieu_de as ten_tour,
    COALESCE(SUM(dc.so_nguoi_lon + dc.so_tre_em), 0) as so_cho_da_dat
FROM khoi_hanh_tour kh
JOIN tour t ON kh.tour_id = t.id
LEFT JOIN dat_cho dc ON kh.id = dc.khoi_hanh_id 
    AND dc.trang_thai NOT IN ('da_huy')
WHERE kh.ngay_khoi_hanh >= $1 AND kh.ngay_khoi_hanh <= $2
GROUP BY kh.id, t.tieu_de
ORDER BY kh.ngay_khoi_hanh ASC;

-- name: GetAllDepartures :many
SELECT 
    kh.*,
    t.tieu_de as ten_tour,
    COALESCE(SUM(dc.so_nguoi_lon + dc.so_tre_em), 0) as so_cho_da_dat
FROM khoi_hanh_tour kh
JOIN tour t ON kh.tour_id = t.id
LEFT JOIN dat_cho dc ON kh.id = dc.khoi_hanh_id 
    AND dc.trang_thai NOT IN ('da_huy')
GROUP BY kh.id, t.tieu_de
ORDER BY kh.ngay_khoi_hanh DESC
LIMIT $1 OFFSET $2;

-- name: CountAllDepartures :one
SELECT COUNT(*) FROM khoi_hanh_tour;

-- name: CountDeparturesByTour :one
SELECT COUNT(*) FROM khoi_hanh_tour
WHERE tour_id = $1;

-- name: GetDeparturesByStatus :many
SELECT 
    kh.*,
    t.tieu_de as ten_tour,
    COALESCE(SUM(dc.so_nguoi_lon + dc.so_tre_em), 0) as so_cho_da_dat
FROM khoi_hanh_tour kh
JOIN tour t ON kh.tour_id = t.id
LEFT JOIN dat_cho dc ON kh.id = dc.khoi_hanh_id 
    AND dc.trang_thai NOT IN ('da_huy')
WHERE kh.trang_thai = $1
GROUP BY kh.id, t.tieu_de
ORDER BY kh.ngay_khoi_hanh ASC
LIMIT $2 OFFSET $3;

-- name: UpdateDepartureStat :one
UPDATE khoi_hanh_tour
SET 
    ngay_cap_nhat = NOW()
WHERE id = $1
RETURNING *;

