-- name: GetAllTourCategory :many
select danh_muc_tour.*, total_tours 
from danh_muc_tour
left join (
    select danh_muc_id, COALESCE(COUNT(*), 0) AS total_tours
    from tour
    where dang_hoat_dong = true
    group by danh_muc_id
) as t on t.danh_muc_id = danh_muc_tour.id;
-- name: CreateCategoryTour :one
INSERT INTO danh_muc_tour (ten, mo_ta, anh, dang_hoat_dong, ngay_tao) VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: GetAllTour :many
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
  dm.ten AS danh_muc_ten,
  ncc.ten AS nha_cung_cap_ten,
  (
    SELECT a.duong_dan
    FROM anh_tour a
    WHERE a.tour_id = t.id
    ORDER BY COALESCE(a.la_anh_chinh, false) DESC, COALESCE(a.thu_tu_hien_thi, 0) ASC, a.id ASC
    LIMIT 1
  ) AS anh_chinh,
  dd.diem_den,
  COALESCE(dg.avg_rating, 0) as avg_rating,
  COALESCE(dg.total_reviews, 0) as total_reviews,
  kh.next_departure_date,
  ggt.phan_tram as giam_gia_phan_tram,
  -- Tính giá sau giảm
  CASE 
    WHEN ggt.phan_tram IS NOT NULL THEN 
      ROUND(t.gia_nguoi_lon * (1 - ggt.phan_tram / 100), 2)
    ELSE t.gia_nguoi_lon
  END as gia_sau_giam_nguoi_lon,
  CASE 
    WHEN ggt.phan_tram IS NOT NULL THEN 
      ROUND(t.gia_tre_em * (1 - ggt.phan_tram / 100), 2)
    ELSE t.gia_tre_em
  END as gia_sau_giam_tre_em
FROM tour t
LEFT JOIN danh_muc_tour dm ON dm.id = t.danh_muc_id
LEFT JOIN nha_cung_cap ncc ON ncc.id = t.nha_cung_cap_id
LEFT JOIN (
  SELECT td.tour_id, array_agg(DISTINCT d.ten ORDER BY d.ten) AS diem_den
  FROM tour_diem_den td
  JOIN diem_den d ON d.id = td.diem_den_id
  GROUP BY td.tour_id
) dd ON dd.tour_id = t.id
LEFT JOIN (
  SELECT tour_id, AVG(diem_danh_gia)::float AS avg_rating, COUNT(*)::int AS total_reviews
  FROM danh_gia
  WHERE dang_hoat_dong = TRUE
  GROUP BY tour_id
) dg ON dg.tour_id = t.id
LEFT JOIN (
  SELECT
    tour_id,
    MIN(ngay_khoi_hanh) FILTER (WHERE ngay_khoi_hanh >= CURRENT_DATE AND trang_thai IN ('len_lich','xac_nhan','con_cho')) AS next_departure_date
  FROM khoi_hanh_tour
  GROUP BY tour_id
) kh ON kh.tour_id = t.id
LEFT JOIN (
  SELECT tour_id, MAX(phan_tram) as phan_tram
  FROM giam_gia_tour
  WHERE CURRENT_DATE BETWEEN ngay_bat_dau AND ngay_ket_thuc
  GROUP BY tour_id
) ggt ON ggt.tour_id = t.id
WHERE t.dang_hoat_dong = TRUE
ORDER BY t.noi_bat DESC, t.ngay_tao DESC
LIMIT $1 OFFSET $2;

-- name: GetTourDetailByID :one
WITH tour_info AS (
    -- Thông tin cơ bản tour
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
        t.nha_cung_cap_id,
        t.ngay_tao,
        t.ngay_cap_nhat,
        dm.ten as ten_danh_muc,
        ncc.ten as ten_nha_cung_cap,
        ncc.logo as logo_ncc
    FROM tour t
    LEFT JOIN danh_muc_tour dm ON t.danh_muc_id = dm.id
    LEFT JOIN nha_cung_cap ncc ON t.nha_cung_cap_id = ncc.id
    WHERE t.id = $1
),
tour_images AS (
    -- Ảnh tour
    SELECT 
        tour_id,
        json_agg(
            json_build_object(
                'id', id,
                'duong_dan', duong_dan,
                'mo_ta', mo_ta,
                'la_anh_chinh', la_anh_chinh,
                'thu_tu_hien_thi', thu_tu_hien_thi
            ) ORDER BY thu_tu_hien_thi, la_anh_chinh DESC
        ) as images
    FROM anh_tour
    WHERE tour_id = $1
    GROUP BY tour_id
),
tour_destinations AS (
    -- Điểm đến
    SELECT 
        tdd.tour_id,
        json_agg(
            json_build_object(
                'id', dd.id,
                'ten', dd.ten,
                'tinh', dd.tinh,
                'quoc_gia', dd.quoc_gia,
                'khu_vuc', dd.khu_vuc,
                'mo_ta', dd.mo_ta,
                'anh', dd.anh,
                'vi_do', dd.vi_do,
                'kinh_do', dd.kinh_do,
                'thu_tu_tham_quan', tdd.thu_tu_tham_quan
            ) ORDER BY tdd.thu_tu_tham_quan
        ) as destinations
    FROM tour_diem_den tdd
    INNER JOIN diem_den dd ON tdd.diem_den_id = dd.id
    WHERE tdd.tour_id = $1
    GROUP BY tdd.tour_id
),
tour_itinerary AS (
    -- Lịch trình và hoạt động
    SELECT 
        ltt.tour_id,
        json_agg(
            json_build_object(
                'id', ltt.id,
                'ngay_thu', ltt.ngay_thu,
                'tieu_de', ltt.tieu_de,
                'mo_ta', ltt.mo_ta,
                'gio_bat_dau', ltt.gio_bat_dau,
                'gio_ket_thuc', ltt.gio_ket_thuc,
                'dia_diem', ltt.dia_diem,
                'thong_tin_luu_tru', ltt.thong_tin_luu_tru,
                'hoat_dong', (
                    SELECT json_agg(
                        json_build_object(
                            'id', hdlt.id,
                            'ten', hdlt.ten,
                            'gio_bat_dau', hdlt.gio_bat_dau,
                            'gio_ket_thuc', hdlt.gio_ket_thuc,
                            'mo_ta', hdlt.mo_ta,
                            'thu_tu', hdlt.thu_tu
                        ) ORDER BY hdlt.thu_tu
                    )
                    FROM hoat_dong_trong_ngay hdlt
                    WHERE hdlt.lich_trinh_id = ltt.id
                )
            ) ORDER BY ltt.ngay_thu
        ) as itinerary
    FROM lich_trinh ltt
    WHERE ltt.tour_id = $1
    GROUP BY ltt.tour_id
),
tour_departures AS (
    -- Lịch khởi hành
    SELECT 
        tour_id,
        json_agg(
            json_build_object(
                'id', id,
                'ngay_khoi_hanh', ngay_khoi_hanh,
                'ngay_ket_thuc', ngay_ket_thuc,
                'suc_chua', suc_chua,
                'trang_thai', trang_thai,
                'ghi_chu', ghi_chu,
                'so_cho_da_dat', so_cho_da_dat
            ) ORDER BY ngay_khoi_hanh
        ) as departures
    FROM khoi_hanh_tour
    WHERE tour_id = $1
    AND ngay_khoi_hanh >= CURRENT_DATE
    AND trang_thai IN ('len_lich', 'xac_nhan', 'con_cho')
    GROUP BY tour_id
),
tour_discount AS (
    -- Giảm giá hiện tại
    SELECT 
        tour_id,
        phan_tram,
        ngay_bat_dau,
        ngay_ket_thuc
    FROM giam_gia_tour
    WHERE tour_id = $1
    AND CURRENT_DATE BETWEEN ngay_bat_dau AND ngay_ket_thuc
    ORDER BY phan_tram DESC
    LIMIT 1
),
tour_config AS (
    -- Cấu hình nhóm
    SELECT 
        tour_id,
        so_nho_nhat,
        so_lon_nhat
    FROM cau_hinh_nhom_tour
    WHERE tour_id = $1
)
SELECT 
    ti.*,
    COALESCE(timg.images, '[]'::json) as images,
    COALESCE(td.destinations, '[]'::json) as destinations,
    COALESCE(tit.itinerary, '[]'::json) as itinerary,
    COALESCE(tdep.departures, '[]'::json) as departures,
    tdisc.phan_tram as giam_gia_phan_tram,
    tdisc.ngay_bat_dau as giam_gia_tu,
    tdisc.ngay_ket_thuc as giam_gia_den,
    tc.so_nho_nhat,
    tc.so_lon_nhat
FROM tour_info ti
LEFT JOIN tour_images timg ON ti.id = timg.tour_id
LEFT JOIN tour_destinations td ON ti.id = td.tour_id
LEFT JOIN tour_itinerary tit ON ti.id = tit.tour_id
LEFT JOIN tour_departures tdep ON ti.id = tdep.tour_id
LEFT JOIN tour_discount tdisc ON ti.id = tdisc.tour_id
LEFT JOIN tour_config tc ON ti.id = tc.tour_id;



-- name: CountAllTours :one
SELECT COUNT(*) 
FROM tour t
WHERE t.dang_hoat_dong = TRUE;


-- ==================== TOUR CRUD OPERATIONS ====================

-- name: CreateTour :one
INSERT INTO tour (
    tieu_de,
    mo_ta,
    danh_muc_id,
    so_ngay,
    so_dem,
    gia_nguoi_lon,
    gia_tre_em,
    don_vi_tien_te,
    trang_thai,
    noi_bat,
    nha_cung_cap_id,
    dang_hoat_dong
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
)
RETURNING *;

-- name: UpdateTour :one
UPDATE tour
SET 
    tieu_de = COALESCE(sqlc.narg('tieu_de'), tieu_de),
    mo_ta = COALESCE(sqlc.narg('mo_ta'), mo_ta),
    danh_muc_id = COALESCE(sqlc.narg('danh_muc_id'), danh_muc_id),
    so_ngay = COALESCE(sqlc.narg('so_ngay'), so_ngay),
    so_dem = COALESCE(sqlc.narg('so_dem'), so_dem),
    gia_nguoi_lon = COALESCE(sqlc.narg('gia_nguoi_lon'), gia_nguoi_lon),
    gia_tre_em = COALESCE(sqlc.narg('gia_tre_em'), gia_tre_em),
    don_vi_tien_te = COALESCE(sqlc.narg('don_vi_tien_te'), don_vi_tien_te),
    trang_thai = COALESCE(sqlc.narg('trang_thai'), trang_thai),
    noi_bat = COALESCE(sqlc.narg('noi_bat'), noi_bat),
    nha_cung_cap_id = COALESCE(sqlc.narg('nha_cung_cap_id'), nha_cung_cap_id),
    dang_hoat_dong = COALESCE(sqlc.narg('dang_hoat_dong'), dang_hoat_dong),
    ngay_cap_nhat = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteTour :exec
DELETE FROM tour
WHERE id = $1;

-- name: ToggleTourActive :one
UPDATE tour
SET 
    dang_hoat_dong = NOT dang_hoat_dong,
    ngay_cap_nhat = NOW()
WHERE id = $1
RETURNING *;


-- name: SearchTours :many
-- Đảm bảo bạn đã định nghĩa các CTE cần thiết (dd, dg, kh, ggt) như trong các bước trước.
-- Tôi sẽ gộp CTE tổng hợp thông tin tour cơ bản vào tour_info.

WITH tour_info AS (
  -- Lấy thông tin cơ bản của Tour và tính toán giá sau giảm
  SELECT
    t.id, t.tieu_de, t.mo_ta, t.danh_muc_id, t.so_ngay, t.so_dem, t.gia_nguoi_lon, t.gia_tre_em, 
    t.don_vi_tien_te, t.trang_thai, t.noi_bat, t.ngay_tao,
    dm.ten AS danh_muc_ten, 
    ncc.ten AS nha_cung_cap_ten,
    -- Subquery lấy Ảnh Chính
    (
      SELECT a.duong_dan FROM anh_tour a 
      WHERE a.tour_id = t.id
      ORDER BY COALESCE(a.la_anh_chinh, FALSE) DESC, COALESCE(a.thu_tu_hien_thi, 0) ASC LIMIT 1
    ) AS anh_chinh
  FROM tour t
  LEFT JOIN danh_muc_tour dm ON dm.id = t.danh_muc_id
  LEFT JOIN nha_cung_cap ncc ON ncc.id = t.nha_cung_cap_id
  WHERE t.dang_hoat_dong = TRUE -- Lọc Tour đang hoạt động
),
dd AS (
  -- Tổng hợp Điểm đến thành mảng (diem_den)
  SELECT td.tour_id, array_agg(DISTINCT d.ten ORDER BY d.ten) AS diem_den
  FROM tour_diem_den td
  JOIN diem_den d ON d.id = td.diem_den_id
  GROUP BY td.tour_id
),
dg AS (
  -- Tính điểm Đánh giá trung bình và Tổng số đánh giá
  SELECT tour_id, AVG(diem_danh_gia)::FLOAT AS avg_rating, COUNT(*)::INT AS total_reviews
  FROM danh_gia
  WHERE dang_hoat_dong = TRUE
  GROUP BY tour_id
),
kh AS (
  -- Tìm Ngày Khởi hành gần nhất và hợp lệ
  SELECT
    tour_id,
    MIN(ngay_khoi_hanh) FILTER (WHERE ngay_khoi_hanh >= CURRENT_DATE AND trang_thai IN ('len_lich','xac_nhan','con_cho')) AS next_departure_date
  FROM khoi_hanh_tour
  GROUP BY tour_id
),
ggt AS (
  -- Tìm Mức Giảm giá cao nhất đang có hiệu lực
  SELECT DISTINCT ON (tour_id)
    tour_id, 
    phan_tram,
    ngay_bat_dau AS giam_gia_tu,
    ngay_ket_thuc AS giam_gia_den
  FROM giam_gia_tour
  WHERE CURRENT_DATE BETWEEN ngay_bat_dau AND ngay_ket_thuc
  ORDER BY tour_id, phan_tram DESC -- Lấy giảm giá % cao nhất
),
search_keyword_match AS (
    -- Lọc Tour theo Từ khóa (trong Tiêu đề HOẶC Hoạt động/Lịch trình)
    SELECT ti.id AS tour_id FROM tour_info ti
    WHERE sqlc.narg('keyword')::TEXT IS NOT NULL AND (
        to_tsvector('vietnamese', ti.tieu_de) @@ plainto_tsquery('vietnamese', sqlc.narg('keyword'))
        OR EXISTS (
            SELECT 1 FROM lich_trinh lt
            JOIN hoat_dong_trong_ngay hd ON hd.lich_trinh_id = lt.id
            WHERE lt.tour_id = ti.id AND to_tsvector('vietnamese', hd.ten) @@ plainto_tsquery('vietnamese', sqlc.narg('keyword'))
        )
    )
),
search_destination_name_match AS (
    -- Lọc Tour theo Tên điểm đến/Tỉnh/Quốc gia (tối ưu hóa việc lọc)
    SELECT tdd.tour_id FROM tour_diem_den tdd
    JOIN diem_den d ON d.id = tdd.diem_den_id
    WHERE sqlc.narg('diem_den_ten')::TEXT IS NOT NULL 
      AND (
          -- Sử dụng unaccent/lower để tìm kiếm linh hoạt (Cần Index gin_trgm trên cột này để tối ưu)
          unaccent(lower(d.ten)) LIKE '%' || unaccent(lower(sqlc.narg('diem_den_ten'))) || '%'
          OR unaccent(lower(d.tinh)) LIKE '%' || unaccent(lower(sqlc.narg('diem_den_ten'))) || '%'
          OR unaccent(lower(d.quoc_gia)) LIKE '%' || unaccent(lower(sqlc.narg('diem_den_ten'))) || '%'
      )
    GROUP BY tdd.tour_id
)
-- Câu truy vấn Chính: Tổng hợp tất cả thông tin và áp dụng các bộ lọc
SELECT
  ti.id, ti.tieu_de, ti.mo_ta, ti.danh_muc_id, ti.so_ngay, ti.so_dem, ti.gia_nguoi_lon, ti.gia_tre_em, ti.don_vi_tien_te,
  ti.trang_thai, ti.noi_bat, ti.danh_muc_ten, ti.nha_cung_cap_ten, ti.anh_chinh,
  
  dd.diem_den, -- Thông tin Điểm đến
  COALESCE(dg.avg_rating, 0) AS avg_rating, -- Đánh giá
  COALESCE(dg.total_reviews, 0) AS total_reviews,
  kh.next_departure_date, -- Ngày khởi hành gần nhất
  
  ggt.phan_tram AS giam_gia_phan_tram, -- Giảm giá
  ggt.giam_gia_tu,
  ggt.giam_gia_den,

  -- Tính giá sau giảm cho Người lớn
  CASE 
    WHEN ggt.phan_tram IS NOT NULL THEN 
      ROUND(ti.gia_nguoi_lon * (1 - ggt.phan_tram / 100), 2)
    ELSE ti.gia_nguoi_lon
  END AS gia_sau_giam_nguoi_lon,
  -- Tính giá sau giảm cho Trẻ em
  CASE 
    WHEN ggt.phan_tram IS NOT NULL THEN 
      ROUND(ti.gia_tre_em * (1 - ggt.phan_tram / 100), 2)
    ELSE ti.gia_tre_em
  END AS gia_sau_giam_tre_em

FROM tour_info ti
LEFT JOIN dd ON dd.tour_id = ti.id
LEFT JOIN dg ON dg.tour_id = ti.id
LEFT JOIN kh ON kh.tour_id = ti.id
LEFT JOIN ggt ON ggt.tour_id = ti.id
-- LEFT JOIN các bộ lọc để áp dụng WHERE
LEFT JOIN search_keyword_match sm ON sm.tour_id = ti.id
LEFT JOIN search_destination_name_match dnm ON dnm.tour_id = ti.id

WHERE 
  -- Lọc theo Keyword
  (sqlc.narg('keyword')::TEXT IS NULL OR sm.tour_id IS NOT NULL)
  -- Lọc theo Điểm đến (Tên)
  AND (sqlc.narg('diem_den_ten')::TEXT IS NULL OR dnm.tour_id IS NOT NULL)
  -- Lọc theo Điểm đến (ID)
  AND (
    sqlc.narg('diem_den_id')::INT IS NULL 
    OR EXISTS (
      SELECT 1 FROM tour_diem_den tdd WHERE tdd.tour_id = ti.id AND tdd.diem_den_id = sqlc.narg('diem_den_id')
    )
  )
  -- Lọc theo Số ngày
  AND (sqlc.narg('so_ngay_min')::INT IS NULL OR ti.so_ngay >= sqlc.narg('so_ngay_min'))
  AND (sqlc.narg('so_ngay_max')::INT IS NULL OR ti.so_ngay <= sqlc.narg('so_ngay_max'))
  -- Lọc theo Số đêm  
  AND (sqlc.narg('so_dem_min')::INT IS NULL OR ti.so_dem >= sqlc.narg('so_dem_min'))
  AND (sqlc.narg('so_dem_max')::INT IS NULL OR ti.so_dem <= sqlc.narg('so_dem_max'))

ORDER BY ti.noi_bat DESC, ti.ngay_tao DESC
LIMIT $1 OFFSET $2;

-- name: CountSearchTours :one
SELECT COUNT(*)
FROM tour t
WHERE t.dang_hoat_dong = TRUE
  -- Tìm kiếm từ khoá trong tiêu đề và hoạt động
  AND (
    sqlc.narg('keyword')::text IS NULL 
    OR to_tsvector('vietnamese', t.tieu_de) @@ plainto_tsquery('vietnamese', sqlc.narg('keyword'))
    OR EXISTS (
      SELECT 1 FROM lich_trinh lt
      JOIN hoat_dong_trong_ngay hd ON hd.lich_trinh_id = lt.id
      WHERE lt.tour_id = t.id 
        AND to_tsvector('vietnamese', hd.ten) @@ plainto_tsquery('vietnamese', sqlc.narg('keyword'))
    )
  )
  -- Lọc theo điểm đến (tên hoặc ID)
  AND (
    sqlc.narg('diem_den_id')::int IS NULL 
    OR EXISTS (
      SELECT 1 FROM tour_diem_den tdd
      WHERE tdd.tour_id = t.id AND tdd.diem_den_id = sqlc.narg('diem_den_id')
    )
  )
  AND (
    sqlc.narg('diem_den_ten')::text IS NULL 
    OR EXISTS (
      SELECT 1 FROM tour_diem_den tdd
      JOIN diem_den d ON d.id = tdd.diem_den_id
      WHERE tdd.tour_id = t.id 
        AND (
          unaccent(lower(d.ten)) LIKE '%' || unaccent(lower(sqlc.narg('diem_den_ten'))) || '%'
          OR unaccent(lower(d.tinh)) LIKE '%' || unaccent(lower(sqlc.narg('diem_den_ten'))) || '%'
          OR unaccent(lower(d.quoc_gia)) LIKE '%' || unaccent(lower(sqlc.narg('diem_den_ten'))) || '%'
        )
    )
  )
  -- Lọc theo số ngày
  AND (sqlc.narg('so_ngay_min')::int IS NULL OR t.so_ngay >= sqlc.narg('so_ngay_min'))
  AND (sqlc.narg('so_ngay_max')::int IS NULL OR t.so_ngay <= sqlc.narg('so_ngay_max'))
  -- Lọc theo số đêm  
  AND (sqlc.narg('so_dem_min')::int IS NULL OR t.so_dem >= sqlc.narg('so_dem_min'))
  AND (sqlc.narg('so_dem_max')::int IS NULL OR t.so_dem <= sqlc.narg('so_dem_max'));

-- name: FilterTours :many
SELECT
  t.id,
  t.tieu_de,
  t.mo_ta,
  t.so_ngay,
  t.so_dem,
  t.gia_nguoi_lon,
  t.gia_tre_em,
  t.don_vi_tien_te,
  t.noi_bat,
  dm.ten AS danh_muc_ten,
  (
    SELECT a.duong_dan
    FROM anh_tour a
    WHERE a.tour_id = t.id AND a.la_anh_chinh = true
    LIMIT 1
  ) AS anh_chinh,
  COALESCE(AVG(dg.diem_danh_gia), 0) as avg_rating
FROM tour t
LEFT JOIN danh_muc_tour dm ON dm.id = t.danh_muc_id
LEFT JOIN danh_gia dg ON t.id = dg.tour_id AND dg.dang_hoat_dong = true
WHERE t.dang_hoat_dong = true
  AND (sqlc.narg('danh_muc_id')::int IS NULL OR t.danh_muc_id = sqlc.narg('danh_muc_id'))
  AND (sqlc.narg('gia_min')::numeric IS NULL OR t.gia_nguoi_lon >= sqlc.narg('gia_min'))
  AND (sqlc.narg('gia_max')::numeric IS NULL OR t.gia_nguoi_lon <= sqlc.narg('gia_max'))
  AND (sqlc.narg('so_ngay_min')::int IS NULL OR t.so_ngay >= sqlc.narg('so_ngay_min'))
  AND (sqlc.narg('so_ngay_max')::int IS NULL OR t.so_ngay <= sqlc.narg('so_ngay_max'))
GROUP BY t.id, dm.ten
HAVING (sqlc.narg('rating_min')::float IS NULL OR COALESCE(AVG(dg.diem_danh_gia), 0) >= sqlc.narg('rating_min'))
ORDER BY 
  CASE WHEN sqlc.narg('sort_by')::text = 'price_asc' THEN t.gia_nguoi_lon END ASC,
  CASE WHEN sqlc.narg('sort_by')::text = 'price_desc' THEN t.gia_nguoi_lon END DESC,
  CASE WHEN sqlc.narg('sort_by')::text = 'rating' THEN AVG(dg.diem_danh_gia) END DESC NULLS LAST,
  CASE WHEN sqlc.narg('sort_by')::text = 'newest' THEN t.ngay_tao END DESC,
  t.noi_bat DESC,
  t.ngay_tao DESC
LIMIT $1 OFFSET $2;

-- name: GetToursByCategoryFilter :many
SELECT
  t.id,
  t.tieu_de,
  t.mo_ta,
  t.so_ngay,
  t.so_dem,
  t.gia_nguoi_lon,
  t.gia_tre_em,
  t.don_vi_tien_te,
  (
    SELECT a.duong_dan
    FROM anh_tour a
    WHERE a.tour_id = t.id AND a.la_anh_chinh = true
    LIMIT 1
  ) AS anh_chinh
FROM tour t
WHERE t.danh_muc_id = $1 AND t.dang_hoat_dong = true
ORDER BY t.noi_bat DESC, t.ngay_tao DESC
LIMIT $2 OFFSET $3;

-- name: GetFeaturedTours :many
SELECT
  t.id,
  t.tieu_de,
  t.mo_ta,
  t.so_ngay,
  t.so_dem,
  t.gia_nguoi_lon,
  t.gia_tre_em,
  t.don_vi_tien_te,
  dm.ten AS danh_muc_ten,
  (
    SELECT a.duong_dan
    FROM anh_tour a
    WHERE a.tour_id = t.id AND a.la_anh_chinh = true
    LIMIT 1
  ) AS anh_chinh,
  COALESCE(AVG(dg.diem_danh_gia), 0) as avg_rating,
  COUNT(dg.id) as total_reviews
FROM tour t
LEFT JOIN danh_muc_tour dm ON dm.id = t.danh_muc_id
LEFT JOIN danh_gia dg ON t.id = dg.tour_id AND dg.dang_hoat_dong = true
WHERE t.noi_bat = true AND t.dang_hoat_dong = true
GROUP BY t.id, dm.ten
ORDER BY t.ngay_tao DESC
LIMIT $1;

-- name: GetToursBySupplier :many
SELECT * FROM tour
WHERE nha_cung_cap_id = $1
ORDER BY ngay_tao DESC
LIMIT $2 OFFSET $3;

-- name: CountToursBySupplier :one
SELECT COUNT(*) FROM tour
WHERE nha_cung_cap_id = $1;

-- name: GetTourImages :many
SELECT * FROM anh_tour
WHERE tour_id = $1
ORDER BY 
    la_anh_chinh DESC NULLS LAST,
    thu_tu_hien_thi ASC NULLS LAST,
    id ASC;

-- name: AddTourImage :one
INSERT INTO anh_tour (
    tour_id,
    duong_dan,
    mo_ta,
    la_anh_chinh,
    thu_tu_hien_thi
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: UpdateTourImage :one
UPDATE anh_tour
SET 
    duong_dan = COALESCE(sqlc.narg('duong_dan'), duong_dan),
    mo_ta = COALESCE(sqlc.narg('mo_ta'), mo_ta),
    la_anh_chinh = COALESCE(sqlc.narg('la_anh_chinh'), la_anh_chinh),
    thu_tu_hien_thi = COALESCE(sqlc.narg('thu_tu_hien_thi'), thu_tu_hien_thi)
WHERE id = $1 AND tour_id = $2
RETURNING *;

-- name: DeleteTourImage :exec
DELETE FROM anh_tour
WHERE id = $1 AND tour_id = $2;

-- name: SetPrimaryTourImage :exec
UPDATE anh_tour
SET la_anh_chinh = (id = $2)
WHERE tour_id = $1;

-- name: GetTourDestinations :many
SELECT 
    dd.*,
    td.thu_tu_tham_quan
FROM tour_diem_den td
JOIN diem_den dd ON td.diem_den_id = dd.id
WHERE td.tour_id = $1
ORDER BY td.thu_tu_tham_quan ASC NULLS LAST;



-- name: GetDiscountsByTourID :many
SELECT id, tour_id, phan_tram, ngay_bat_dau, ngay_ket_thuc, ngay_tao, ngay_cap_nhat FROM giam_gia_tour
WHERE tour_id = $1
ORDER BY ngay_bat_dau DESC;

-- name: CreateDiscountTour :one
INSERT INTO giam_gia_tour (
    tour_id,
    phan_tram,
    ngay_bat_dau,
    ngay_ket_thuc,
    ngay_tao,
    ngay_cap_nhat
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: UpdateDiscountTour :one
UPDATE giam_gia_tour
SET
    tour_id = COALESCE($2, tour_id),
    phan_tram = COALESCE(sqlc.narg('phan_tram'), phan_tram),
    ngay_bat_dau = COALESCE(sqlc.narg('ngay_bat_dau'), ngay_bat_dau),
    ngay_ket_thuc = COALESCE(sqlc.narg('ngay_ket_thuc'), ngay_ket_thuc),
    ngay_cap_nhat = NOW()
WHERE id = $1 AND tour_id = $2
RETURNING *;

-- name: DeleteDiscountTour :exec
DELETE FROM giam_gia_tour
WHERE id = $1 AND tour_id = $2;

-- name: GetToursByCountryCode :many
-- Lấy tour quốc nội (iso2 = country_code) hoặc quốc tế (iso2 != country_code) sắp xếp theo số lượt đặt
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
  dm.ten AS danh_muc_ten,
  ncc.ten AS nha_cung_cap_ten,
  (
    SELECT a.duong_dan
    FROM anh_tour a
    WHERE a.tour_id = t.id
    ORDER BY COALESCE(a.la_anh_chinh, false) DESC, COALESCE(a.thu_tu_hien_thi, 0) ASC, a.id ASC
    LIMIT 1
  ) AS anh_chinh,
  dd.diem_den,
  COALESCE(dg.avg_rating, 0) as avg_rating,
  COALESCE(dg.total_reviews, 0) as total_reviews,
  COUNT(dc.id) AS so_booking
FROM tour t
LEFT JOIN danh_muc_tour dm ON dm.id = t.danh_muc_id
LEFT JOIN nha_cung_cap ncc ON ncc.id = t.nha_cung_cap_id
LEFT JOIN (
  SELECT td.tour_id, array_agg(DISTINCT d.ten ORDER BY d.ten) AS diem_den
  FROM tour_diem_den td
  JOIN diem_den d ON d.id = td.diem_den_id
  GROUP BY td.tour_id
) dd ON dd.tour_id = t.id
LEFT JOIN (
  SELECT tour_id, AVG(diem_danh_gia)::float AS avg_rating, COUNT(*)::int AS total_reviews
  FROM danh_gia
  WHERE dang_hoat_dong = TRUE
  GROUP BY tour_id
) dg ON dg.tour_id = t.id
LEFT JOIN khoi_hanh_tour kh ON kh.tour_id = t.id
LEFT JOIN dat_cho dc ON dc.khoi_hanh_id = kh.id
WHERE t.dang_hoat_dong = TRUE 
  AND t.trang_thai = 'cong_bo'
  AND EXISTS (
    SELECT 1 
    FROM tour_diem_den tdd
    JOIN diem_den d ON d.id = tdd.diem_den_id
    WHERE tdd.tour_id = t.id
      AND (
        ($1 = 'domestic' AND d.iso2 = $2)
        OR ($1 = 'international' AND d.iso2 IS NOT NULL AND d.iso2 != $2)
      )
  )
GROUP BY t.id, t.tieu_de, t.mo_ta, t.danh_muc_id, t.so_ngay, t.so_dem, 
         t.gia_nguoi_lon, t.gia_tre_em, t.don_vi_tien_te, t.trang_thai, 
         t.noi_bat, dm.ten, ncc.ten, dd.diem_den, dg.avg_rating, dg.total_reviews
ORDER BY so_booking DESC, t.noi_bat DESC, t.ngay_tao DESC
LIMIT $3 OFFSET $4;


