-- ==================== ITINERARY QUERIES ====================

-- name: CreateItinerary :one
INSERT INTO lich_trinh (
    tour_id,
    ngay_thu,
    tieu_de,
    mo_ta,
    gio_bat_dau,
    gio_ket_thuc,
    dia_diem,
    thong_tin_luu_tru
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: GetItinerariesByTour :many
SELECT * FROM lich_trinh
WHERE tour_id = $1
ORDER BY ngay_thu ASC;

-- name: GetItineraryByID :one
SELECT * FROM lich_trinh
WHERE id = $1;

-- name: UpdateItinerary :one
UPDATE lich_trinh
SET 
    tieu_de = COALESCE(sqlc.narg('tieu_de'), tieu_de),
    mo_ta = COALESCE(sqlc.narg('mo_ta'), mo_ta),
    gio_bat_dau = COALESCE(sqlc.narg('gio_bat_dau'), gio_bat_dau),
    gio_ket_thuc = COALESCE(sqlc.narg('gio_ket_thuc'), gio_ket_thuc),
    dia_diem = COALESCE(sqlc.narg('dia_diem'), dia_diem),
    thong_tin_luu_tru = COALESCE(sqlc.narg('thong_tin_luu_tru'), thong_tin_luu_tru),
    ngay_cap_nhat = NOW()
WHERE id = $1 AND tour_id = $2
RETURNING *;

-- name: DeleteItinerary :exec
DELETE FROM lich_trinh
WHERE id = $1 AND tour_id = $2;

-- name: DeleteItinerariesByTour :exec
DELETE FROM lich_trinh
WHERE tour_id = $1;

-- ==================== ACTIVITY QUERIES ====================

-- name: CreateActivity :one
INSERT INTO hoat_dong_trong_ngay (
    lich_trinh_id,
    ten,
    gio_bat_dau,
    gio_ket_thuc,
    mo_ta,
    thu_tu
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetActivitiesByItinerary :many
SELECT * FROM hoat_dong_trong_ngay
WHERE lich_trinh_id = $1
ORDER BY thu_tu ASC NULLS LAST, id ASC;

-- name: GetActivityByID :one
SELECT * FROM hoat_dong_trong_ngay
WHERE id = $1;

-- name: UpdateActivity :one
UPDATE hoat_dong_trong_ngay
SET 
    ten = COALESCE(sqlc.narg('ten'), ten),
    gio_bat_dau = COALESCE(sqlc.narg('gio_bat_dau'), gio_bat_dau),
    gio_ket_thuc = COALESCE(sqlc.narg('gio_ket_thuc'), gio_ket_thuc),
    mo_ta = COALESCE(sqlc.narg('mo_ta'), mo_ta),
    thu_tu = COALESCE(sqlc.narg('thu_tu'), thu_tu)
WHERE id = $1 AND lich_trinh_id = $2
RETURNING *;

-- name: DeleteActivity :exec
DELETE FROM hoat_dong_trong_ngay
WHERE id = $1 AND lich_trinh_id = $2;

-- name: DeleteActivitiesByItinerary :exec
DELETE FROM hoat_dong_trong_ngay
WHERE lich_trinh_id = $1;

-- ==================== COMBINED QUERIES ====================

-- name: GetFullItineraryByTour :many
SELECT 
    lt.id as itinerary_id,
    lt.tour_id,
    lt.ngay_thu,
    lt.tieu_de as itinerary_title,
    lt.mo_ta as itinerary_description,
    lt.gio_bat_dau as itinerary_start_time,
    lt.gio_ket_thuc as itinerary_end_time,
    lt.dia_diem,
    lt.thong_tin_luu_tru,
    hd.id as activity_id,
    hd.ten as activity_name,
    hd.gio_bat_dau as activity_start_time,
    hd.gio_ket_thuc as activity_end_time,
    hd.mo_ta as activity_description,
    hd.thu_tu as activity_order
FROM lich_trinh lt
LEFT JOIN hoat_dong_trong_ngay hd ON lt.id = hd.lich_trinh_id
WHERE lt.tour_id = $1
ORDER BY lt.ngay_thu ASC, hd.thu_tu ASC NULLS LAST, hd.id ASC;

-- ==================== GROUP CONFIG QUERIES ====================

-- name: CreateGroupConfig :one
INSERT INTO cau_hinh_nhom_tour (
    tour_id,
    so_nho_nhat,
    so_lon_nhat
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetGroupConfigByTour :one
SELECT * FROM cau_hinh_nhom_tour
WHERE tour_id = $1;

-- name: UpdateGroupConfig :one
UPDATE cau_hinh_nhom_tour
SET 
    so_nho_nhat = COALESCE(sqlc.narg('so_nho_nhat'), so_nho_nhat),
    so_lon_nhat = COALESCE(sqlc.narg('so_lon_nhat'), so_lon_nhat),
    ngay_cap_nhat = NOW()
WHERE tour_id = $1
RETURNING *;

-- name: DeleteGroupConfig :exec
DELETE FROM cau_hinh_nhom_tour
WHERE tour_id = $1;


