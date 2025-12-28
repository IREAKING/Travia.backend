-- name: CreateDestination :one
INSERT INTO diem_den (
    ten,
    quoc_gia,
    khu_vuc,
    mo_ta,
    anh,
    vi_do,
    kinh_do
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetCountry :many
SELECT quoc_gia FROM diem_den
GROUP BY quoc_gia;

-- name: GetProvinceByCountry :many
SELECT tinh FROM diem_den
WHERE quoc_gia = $1
GROUP BY tinh;

-- name: GetCityByProvince :many
SELECT ten FROM diem_den
WHERE tinh = $1
GROUP BY ten;


-- name: GetDestinationByID :one
SELECT * FROM diem_den
WHERE id = $1;

-- name: GetAllDestinations :many
SELECT * FROM diem_den
ORDER BY ngay_tao DESC;

-- name: GetDestinationsByCountry :many
SELECT * FROM diem_den
WHERE quoc_gia = $1
ORDER BY ten ASC;

-- name: GetDestinationsByRegion :many
SELECT * FROM diem_den
WHERE khu_vuc = $1
ORDER BY ten ASC;

-- name: GetDestinationsByCountryAndRegion :many
SELECT * FROM diem_den
WHERE quoc_gia = $1 AND khu_vuc = $2
ORDER BY ten ASC;

-- name: UpdateDestination :one
UPDATE diem_den
SET
    ten = COALESCE(sqlc.narg(ten), ten),
    quoc_gia = COALESCE(sqlc.narg(quoc_gia), quoc_gia),
    khu_vuc = COALESCE(sqlc.narg(khu_vuc), khu_vuc),
    mo_ta = COALESCE(sqlc.narg(mo_ta), mo_ta),
    anh = COALESCE(sqlc.narg(anh), anh),
    vi_do = COALESCE(sqlc.narg(vi_do), vi_do),
    kinh_do = COALESCE(sqlc.narg(kinh_do), kinh_do),
    ngay_cap_nhat = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteDestination :exec
DELETE FROM diem_den
WHERE id = $1;

-- name: SearchDestinations :many
SELECT * FROM diem_den
WHERE 
    ten ILIKE '%' || $1 || '%'
    OR quoc_gia ILIKE '%' || $1 || '%'
    OR khu_vuc ILIKE '%' || $1 || '%'
    OR mo_ta ILIKE '%' || $1 || '%'
ORDER BY ten ASC;

-- name: GetDestinationWithTourCount :many
SELECT
    d.*,
    COUNT(td.tour_id)::int AS total_tours
FROM diem_den d
LEFT JOIN tour_diem_den td ON td.diem_den_id = d.id
GROUP BY d.id
ORDER BY total_tours DESC, d.ngay_tao DESC;

-- name: GetDestinationWithTourCountByID :one
SELECT
    d.*,
    COUNT(td.tour_id)::int AS total_tours
FROM diem_den d
LEFT JOIN tour_diem_den td ON td.diem_den_id = d.id
WHERE d.id = $1
GROUP BY d.id;

-- name: GetPopularDestinations :many
SELECT
    d.*,
    COUNT(td.tour_id)::int AS total_tours
FROM diem_den d
LEFT JOIN tour_diem_den td ON td.diem_den_id = d.id
LEFT JOIN tour t ON t.id = td.tour_id AND t.dang_hoat_dong = TRUE
where d.anh is not null
GROUP BY d.id
HAVING COUNT(td.tour_id) > 0
ORDER BY total_tours DESC
LIMIT $1;

-- name: GetDestinationsWithPagination :many
SELECT * FROM diem_den
ORDER BY ngay_tao DESC
LIMIT $1 OFFSET $2;

-- name: CountDestinations :one
SELECT COUNT(*)::int FROM diem_den;

-- name: CountDestinationsByCountry :many
SELECT
    quoc_gia,
    COUNT(*)::int AS total
FROM diem_den
GROUP BY quoc_gia
ORDER BY total DESC;

-- name: GetDestinationsByTourID :many
SELECT 
    d.id,
    d.ten,
    d.quoc_gia,
    d.khu_vuc,
    d.mo_ta,
    d.anh,
    d.vi_do,
    d.kinh_do,
    td.thu_tu_tham_quan
FROM diem_den d
JOIN tour_diem_den td ON td.diem_den_id = d.id
WHERE td.tour_id = $1
ORDER BY td.thu_tu_tham_quan ASC;

-- name: CheckDestinationExists :one
SELECT EXISTS(
    SELECT 1 FROM diem_den
    WHERE ten = $1 AND quoc_gia = $2
) AS exists;

-- name: GetUniqueCountries :many
SELECT DISTINCT quoc_gia
FROM diem_den
WHERE quoc_gia IS NOT NULL
ORDER BY quoc_gia ASC;

-- name: BulkDeleteDestinations :exec
DELETE FROM diem_den
WHERE id = ANY($1::int[]);


-- name: GetDestinationsWithoutImage :many
SELECT id, tinh
FROM diem_den
WHERE (anh IS NULL OR anh = '')
LIMIT $1;

-- name: UpdateDestinationImage :exec
UPDATE diem_den
SET anh = $1, ngay_cap_nhat = NOW()
WHERE id = $2;
