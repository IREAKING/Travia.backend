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
SELECT id, quoc_gia FROM diem_den
GROUP BY quoc_gia, id;

-- name: GetProvinceByCountry :many
SELECT id, tinh FROM diem_den
WHERE quoc_gia = $1
GROUP BY tinh, id;

-- name: GetCityByProvince :many
SELECT id, ten FROM diem_den
WHERE tinh = $1
GROUP BY ten, id;

-- name: GetPopularDestinations :many
-- Lấy các điểm đến phổ biến nhất (được nhiều tour sử dụng nhất)
SELECT 
    dd.id,
    dd.ten,
    dd.tinh,
    dd.quoc_gia,
    dd.khu_vuc,
    dd.mo_ta,
    dd.anh,
    dd.vi_do,
    dd.kinh_do,
    COUNT(DISTINCT CASE WHEN t.trang_thai = 'cong_bo' AND t.dang_hoat_dong = TRUE THEN tdd.tour_id END) as so_luong_tour
FROM diem_den dd
LEFT JOIN tour_diem_den tdd ON dd.id = tdd.diem_den_id
LEFT JOIN tour t ON tdd.tour_id = t.id
GROUP BY dd.id, dd.ten, dd.tinh, dd.quoc_gia, dd.khu_vuc, dd.mo_ta, dd.anh, dd.vi_do, dd.kinh_do
HAVING COUNT(DISTINCT CASE WHEN t.trang_thai = 'cong_bo' AND t.dang_hoat_dong = TRUE THEN tdd.tour_id END) > 0
ORDER BY so_luong_tour DESC, dd.ten ASC
LIMIT $1;

-- name: GetTopPopularDestinations :many
-- Lấy top N điểm đến phổ biến nhất với thông tin chi tiết (có số tour nổi bật)
SELECT 
    dd.id,
    dd.ten,
    dd.tinh,
    dd.quoc_gia,
    dd.khu_vuc,
    dd.mo_ta,
    dd.anh,
    dd.vi_do,
    dd.kinh_do,
    COUNT(DISTINCT tdd.tour_id) as so_luong_tour,
    COUNT(DISTINCT CASE WHEN t.noi_bat = TRUE THEN t.id END) as so_tour_noi_bat
FROM diem_den dd
INNER JOIN tour_diem_den tdd ON dd.id = tdd.diem_den_id
INNER JOIN tour t ON tdd.tour_id = t.id
WHERE t.trang_thai = 'cong_bo' 
    AND t.dang_hoat_dong = TRUE
GROUP BY dd.id, dd.ten, dd.tinh, dd.quoc_gia, dd.khu_vuc, dd.mo_ta, dd.anh, dd.vi_do, dd.kinh_do
HAVING COUNT(DISTINCT tdd.tour_id) > 0
ORDER BY so_luong_tour DESC, so_tour_noi_bat DESC, dd.ten ASC
LIMIT $1;

-- name: GetDestinationByID :one
-- Lấy thông tin chi tiết điểm đến theo ID
SELECT 
    id,
    ten,
    tinh,
    quoc_gia,
    khu_vuc,
    iso2,
    iso3,
    mo_ta,
    anh,
    vi_do,
    kinh_do,
    ngay_tao,
    ngay_cap_nhat
FROM diem_den
WHERE id = $1;

-- name: UpdateDestinationImage :one
-- Cập nhật hình ảnh cho điểm đến
UPDATE diem_den
SET anh = $2,
    ngay_cap_nhat = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;
