-- name: GetAllTourCategory :many
select * from danh_muc_tour;

-- name: GetAllTour :many
SELECT
  id,
  tieu_de,
  mo_ta,
  danh_muc_id,
  so_ngay,
  so_dem,
  gia_moi_nguoi,
  don_vi_tien_te,
  trang_thai,
  noi_bat,
  nguoi_tao_id,
  nha_cung_cap_id,
  dang_hoat_dong,
  ngay_tao,
  ngay_cap_nhat
FROM tour;

-- name: GetAllTourWithRelations :many
SELECT
  t.id,
  t.tieu_de,
  t.mo_ta,
  t.danh_muc_id,
  t.so_ngay,
  t.so_dem,
  t.gia_moi_nguoi,
  t.don_vi_tien_te,
  t.trang_thai,
  t.noi_bat,
  t.nguoi_tao_id,
  t.nha_cung_cap_id,
  t.dang_hoat_dong,
  t.ngay_tao,
  t.ngay_cap_nhat,
  dm.ten AS danh_muc_ten,
  ncc.ten AS nha_cung_cap_ten,
  (
    SELECT a.link
    FROM anh_tour a
    WHERE a.tour_id = t.id
    ORDER BY COALESCE(a.la_anh_chinh, false) DESC, COALESCE(a.thu_tu_hien_thi, 0) ASC, a.id ASC
    LIMIT 1
  ) AS anh_chinh,
  dd.diem_den,
  dg.avg_rating,
  dg.total_reviews,
  kh.next_departure_date,
  kh.min_price
FROM tour t
LEFT JOIN danh_muc_tour dm ON dm.id = t.danh_muc_id
LEFT JOIN nha_cung_cap ncc ON ncc.id = t.nha_cung_cap_id
LEFT JOIN (
  SELECT td.tour_id, array_agg(DISTINCT d.ten) AS diem_den
  FROM tour_diem_den td
  JOIN diem_den d ON d.id = td.diem_den_id
  GROUP BY td.tour_id
) dd ON dd.tour_id = t.id
LEFT JOIN (
  SELECT tour_id, AVG(rating)::float AS avg_rating, COUNT(*)::int AS total_reviews
  FROM danh_gia
  WHERE dang_hoat_dong = TRUE
  GROUP BY tour_id
) dg ON dg.tour_id = t.id
LEFT JOIN (
  SELECT
    tour_id,
    MIN(ngay_khoi_hanh) FILTER (WHERE ngay_khoi_hanh >= CURRENT_DATE AND trang_thai IN ('len_lich','xac_nhan')) AS next_departure_date,
    MIN(gia_dac_biet) FILTER (WHERE ngay_khoi_hanh >= CURRENT_DATE AND trang_thai IN ('len_lich','xac_nhan') AND gia_dac_biet IS NOT NULL) AS min_price
  FROM khoi_hanh_tour
  GROUP BY tour_id
) kh ON kh.tour_id = t.id

ORDER BY t.ngay_tao DESC;
-- name: GetTourDetailByID :one
SELECT
  t.id,
  t.tieu_de,
  t.mo_ta,
  t.danh_muc_id,
  t.so_ngay,
  t.so_dem,
  t.gia_moi_nguoi,
  t.don_vi_tien_te,
  t.trang_thai,
  t.noi_bat,
  t.nguoi_tao_id,
  t.nha_cung_cap_id,
  t.dang_hoat_dong,
  t.ngay_tao,
  t.ngay_cap_nhat
FROM tour t
LEFT JOIN danh_muc_tour dm ON dm.id = t.danh_muc_id
LEFT JOIN nha_cung_cap ncc ON ncc.id = t.nha_cung_cap_id
LEFT JOIN (
  SELECT td.tour_id, array_agg(DISTINCT d.ten) AS diem_den
  FROM tour_diem_den td
  JOIN diem_den d ON d.id = td.diem_den_id
  GROUP BY td.tour_id
) dd ON dd.tour_id = t.id
LEFT JOIN (
  SELECT tour_id, AVG(rating)::float AS avg_rating, COUNT(*)::int AS total_reviews
  FROM danh_gia
  WHERE dang_hoat_dong = TRUE
  GROUP BY tour_id
) dg ON dg.tour_id = t.id
LEFT JOIN (
  SELECT tour_id, MIN(ngay_khoi_hanh) FILTER (WHERE ngay_khoi_hanh >= CURRENT_DATE AND trang_thai IN ('len_lich','xac_nhan')) AS next_departure_date, MIN(gia_dac_biet) FILTER (WHERE ngay_khoi_hanh >= CURRENT_DATE AND trang_thai IN ('len_lich','xac_nhan') AND gia_dac_biet IS NOT NULL) AS min_price
  FROM khoi_hanh_tour
  GROUP BY tour_id
) kh ON kh.tour_id = t.id
WHERE t.id = $1
ORDER BY t.ngay_tao DESC;

-- name: GetTourDetailByIDWithRelations :one
SELECT
  t.id,
  t.tieu_de,
  t.mo_ta,
  t.danh_muc_id,
  t.so_ngay,
  t.so_dem,
  t.gia_moi_nguoi,
  t.don_vi_tien_te,
  t.trang_thai,
  t.noi_bat,
  t.nguoi_tao_id,
  t.nha_cung_cap_id,
  t.dang_hoat_dong,
  t.ngay_tao,
  t.ngay_cap_nhat
FROM tour t
LEFT JOIN danh_muc_tour dm ON dm.id = t.danh_muc_id
LEFT JOIN nha_cung_cap ncc ON ncc.id = t.nha_cung_cap_id
LEFT JOIN (
  SELECT td.tour_id, array_agg(DISTINCT d.ten) AS diem_den
  FROM tour_diem_den td
  JOIN diem_den d ON d.id = td.diem_den_id
  GROUP BY td.tour_id
) dd ON dd.tour_id = t.id
LEFT JOIN (
  SELECT tour_id, AVG(rating)::float AS avg_rating, COUNT(*)::int AS total_reviews
  FROM danh_gia
  WHERE dang_hoat_dong = TRUE
  GROUP BY tour_id
) dg ON dg.tour_id = t.id
LEFT JOIN (
  SELECT tour_id, MIN(ngay_khoi_hanh) FILTER (WHERE ngay_khoi_hanh >= CURRENT_DATE AND trang_thai IN ('len_lich','xac_nhan')) AS next_departure_date, MIN(gia_dac_biet) FILTER (WHERE ngay_khoi_hanh >= CURRENT_DATE AND trang_thai IN ('len_lich','xac_nhan') AND gia_dac_biet IS NOT NULL) AS min_price
  FROM khoi_hanh_tour
  GROUP BY tour_id
) kh ON kh.tour_id = t.id
WHERE t.id = $1
ORDER BY t.ngay_tao DESC;
