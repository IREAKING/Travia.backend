-- name: GetTourContextForAI :many
WITH tour_info AS (
  SELECT
    t.id,
    t.tieu_de,
    t.mo_ta,
    t.danh_muc_id,
    t.so_ngay,
    t.so_dem,
    t.gia_nguoi_lon,
    t.gia_tre_em,
    t.don_vi_tien_te,
    t.trang_thai,
    t.noi_bat,
    t.ngay_tao,
    dm.ten AS danh_muc_ten,
    ncc.ten AS nha_cung_cap_ten
  FROM tour t
  LEFT JOIN danh_muc_tour dm ON dm.id = t.danh_muc_id
  LEFT JOIN nha_cung_cap ncc ON ncc.id = t.nha_cung_cap_id
  WHERE t.dang_hoat_dong = TRUE
),
destinations AS (
  SELECT
    td.tour_id,
    json_agg(
      json_build_object(
        'id', d.id,
        'ten', d.ten,
        'tinh', d.tinh,
        'quoc_gia', d.quoc_gia,
        'khu_vuc', d.khu_vuc
      ) ORDER BY td.thu_tu_tham_quan
    ) AS destinations
  FROM tour_diem_den td
  JOIN diem_den d ON d.id = td.diem_den_id
  GROUP BY td.tour_id
),
itinerary AS (
  SELECT
    lt.tour_id,
    json_agg(
      json_build_object(
        'ngay_thu', lt.ngay_thu,
        'tieu_de', lt.tieu_de,
        'mo_ta', lt.mo_ta,
        'dia_diem', lt.dia_diem,
        'hoat_dong', (
          SELECT json_agg(
            json_build_object(
              'ten', hd.ten,
              'gio_bat_dau', hd.gio_bat_dau,
              'gio_ket_thuc', hd.gio_ket_thuc,
              'mo_ta', hd.mo_ta
            ) ORDER BY hd.thu_tu
          )
          FROM hoat_dong_trong_ngay hd
          WHERE hd.lich_trinh_id = lt.id
        )
      ) ORDER BY lt.ngay_thu
    ) AS itinerary
  FROM lich_trinh lt
  GROUP BY lt.tour_id
),
departures AS (
  SELECT
    k.tour_id,
    json_agg(
      json_build_object(
        'ngay_khoi_hanh', k.ngay_khoi_hanh,
        'ngay_ket_thuc', k.ngay_ket_thuc,
        'suc_chua', k.suc_chua,
        'so_cho_da_dat', k.so_cho_da_dat,
        'trang_thai', k.trang_thai
      ) ORDER BY k.ngay_khoi_hanh
    ) AS departures,
    MIN(k.ngay_khoi_hanh) FILTER (
      WHERE k.ngay_khoi_hanh >= CURRENT_DATE
        AND k.trang_thai IN ('len_lich', 'xac_nhan', 'con_cho')
    ) AS next_departure_date
  FROM khoi_hanh_tour k
  GROUP BY k.tour_id
),
discounts AS (
  SELECT DISTINCT ON (tour_id)
    tour_id,
    phan_tram,
    ngay_bat_dau AS giam_gia_tu,
    ngay_ket_thuc AS giam_gia_den
  FROM giam_gia_tour
  WHERE CURRENT_DATE BETWEEN ngay_bat_dau AND ngay_ket_thuc
  ORDER BY tour_id, phan_tram DESC
),
ratings AS (
  SELECT
    tour_id,
    AVG(diem_danh_gia)::FLOAT AS avg_rating,
    COUNT(*)::INT AS total_reviews
  FROM danh_gia
  WHERE dang_hoat_dong = TRUE
  GROUP BY tour_id
)
SELECT
  ti.id,
  ti.tieu_de,
  ti.mo_ta,
  ti.so_ngay,
  ti.so_dem,
  ti.gia_nguoi_lon,
  ti.gia_tre_em,
  ti.don_vi_tien_te,
  ti.danh_muc_ten,
  ti.nha_cung_cap_ten,
  COALESCE(rt.avg_rating, 0) AS avg_rating,
  COALESCE(rt.total_reviews, 0) AS total_reviews,
  COALESCE(des.destinations, '[]'::json) AS destinations,
  COALESCE(itn.itinerary, '[]'::json) AS itinerary,
  COALESCE(dep.departures, '[]'::json) AS departures,
  dep.next_departure_date,
  dis.phan_tram AS giam_gia_phan_tram,
  dis.giam_gia_tu,
  dis.giam_gia_den
FROM tour_info ti
LEFT JOIN destinations des ON des.tour_id = ti.id
LEFT JOIN itinerary itn ON itn.tour_id = ti.id
LEFT JOIN departures dep ON dep.tour_id = ti.id
LEFT JOIN discounts dis ON dis.tour_id = ti.id
LEFT JOIN ratings rt ON rt.tour_id = ti.id
WHERE
  (
    sqlc.narg('keyword')::TEXT IS NULL OR
    to_tsvector('vietnamese', ti.tieu_de) @@ plainto_tsquery('vietnamese', sqlc.narg('keyword')) OR
    to_tsvector('vietnamese', coalesce(ti.mo_ta, '')) @@ plainto_tsquery('vietnamese', sqlc.narg('keyword')) OR
    unaccent(lower(ti.tieu_de)) LIKE '%' || unaccent(lower(sqlc.narg('keyword'))) || '%' OR
    unaccent(lower(coalesce(ti.mo_ta, ''))) LIKE '%' || unaccent(lower(sqlc.narg('keyword'))) || '%' OR
    EXISTS (
      SELECT 1
      FROM tour_diem_den tdd
      JOIN diem_den dd ON dd.id = tdd.diem_den_id
      WHERE tdd.tour_id = ti.id
        AND unaccent(lower(dd.ten)) LIKE '%' || unaccent(lower(sqlc.narg('keyword'))) || '%'
    )
  )
  AND (
    sqlc.narg('diem_den_ten')::TEXT IS NULL OR
    EXISTS (
      SELECT 1
      FROM tour_diem_den tdd
      JOIN diem_den dd ON dd.id = tdd.diem_den_id
      WHERE tdd.tour_id = ti.id
        AND unaccent(lower(dd.ten)) LIKE '%' || unaccent(lower(sqlc.narg('diem_den_ten'))) || '%'
    )
  )
  AND (sqlc.narg('so_ngay_min')::INT IS NULL OR ti.so_ngay >= sqlc.narg('so_ngay_min'))
  AND (sqlc.narg('so_ngay_max')::INT IS NULL OR ti.so_ngay <= sqlc.narg('so_ngay_max'))
  AND (sqlc.narg('gia_min')::NUMERIC IS NULL OR ti.gia_nguoi_lon >= sqlc.narg('gia_min'))
  AND (sqlc.narg('gia_max')::NUMERIC IS NULL OR ti.gia_nguoi_lon <= sqlc.narg('gia_max'))
  AND (
    sqlc.narg('ngay_khoi_hanh_tu')::DATE IS NULL OR
    dep.next_departure_date >= sqlc.narg('ngay_khoi_hanh_tu')
  )
  AND (
    sqlc.narg('ngay_khoi_hanh_den')::DATE IS NULL OR
    dep.next_departure_date <= sqlc.narg('ngay_khoi_hanh_den')
  )
  AND (
    sqlc.narg('chi_co_giam_gia')::BOOL IS NULL OR
    (sqlc.narg('chi_co_giam_gia') = TRUE AND dis.phan_tram IS NOT NULL) OR
    (sqlc.narg('chi_co_giam_gia') = FALSE)
  )
ORDER BY ti.noi_bat DESC, ti.ngay_tao DESC
LIMIT $1 OFFSET $2;
