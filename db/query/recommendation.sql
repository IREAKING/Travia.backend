-- ===========================================
-- QUERIES CHO AI GỢI Ý TOUR
-- ===========================================

-- name: CreateTourViewHistory :one
-- Lưu lịch sử xem tour
INSERT INTO lich_su_xem_tour (
    nguoi_dung_id,
    tour_id,
    thoi_gian_xem,
    thoi_luong_xem_giay,
    ip_address,
    user_agent
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: UpdateTourViewDuration :one
-- Cập nhật thời lượng xem khi người dùng rời khỏi trang
UPDATE lich_su_xem_tour
SET thoi_luong_xem_giay = $3
WHERE id = $1 AND tour_id = $2
RETURNING *;

-- name: GetTourViewHistoryByUser :many
-- Lấy lịch sử xem tour của người dùng
SELECT * FROM lich_su_xem_tour
WHERE nguoi_dung_id = $1
ORDER BY thoi_gian_xem DESC
LIMIT $2 OFFSET $3;

-- name: GetTourViewHistoryByTour :many
-- Lấy lịch sử xem của một tour cụ thể
SELECT * FROM lich_su_xem_tour
WHERE tour_id = $1
ORDER BY thoi_gian_xem DESC
LIMIT $2 OFFSET $3;

-- name: GetMostViewedTours :many
-- Lấy các tour được xem nhiều nhất
SELECT 
    tour_id,
    COUNT(*) as so_luot_xem,
    AVG(thoi_luong_xem_giay) as thoi_luong_xem_trung_binh
FROM lich_su_xem_tour
WHERE thoi_gian_xem >= NOW() - INTERVAL '30 days'
GROUP BY tour_id
ORDER BY so_luot_xem DESC, thoi_luong_xem_trung_binh DESC
LIMIT $1;

-- name: CreateTourEmbedding :one
-- Tạo hoặc cập nhật embedding cho tour
INSERT INTO tour_embeddings (
    tour_id,
    embedding
) VALUES (
    $1, $2
)
ON CONFLICT (tour_id)
DO UPDATE SET 
    embedding = EXCLUDED.embedding,
    ngay_cap_nhat = NOW()
RETURNING *;

-- name: GetTourEmbedding :one
-- Lấy embedding của một tour
SELECT * FROM tour_embeddings
WHERE tour_id = $1;

-- name: GetSimilarToursByEmbedding :many
-- Tìm các tour tương tự dựa trên embedding (semantic search)
SELECT 
    tour_id,
    1 - (embedding <=> $1) as do_tuong_dong
FROM tour_embeddings
WHERE tour_id != $2
ORDER BY embedding <=> $1
LIMIT $3;

-- name: GetUserPreferences :many
-- Lấy sở thích của người dùng
SELECT * FROM so_thich_nguoi_dung
WHERE nguoi_dung_id = $1
ORDER BY diem_so DESC, ngay_cap_nhat DESC;

-- name: GetUserPreferencesByType :many
-- Lấy sở thích theo loại (danh_muc hoặc diem_den)
SELECT * FROM so_thich_nguoi_dung
WHERE nguoi_dung_id = $1 AND loai_so_thich = $2
ORDER BY diem_so DESC, ngay_cap_nhat DESC;

-- name: GetRecommendedToursByPreferences :many
-- Gợi ý tour dựa trên sở thích người dùng (danh mục và điểm đến)
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
    COALESCE(SUM(st.diem_so), 0) as tong_diem_so_thich,
    COALESCE(dg.avg_rating, 0) as avg_rating,
    COALESCE(dg.total_reviews, 0) as total_reviews
FROM tour t
LEFT JOIN danh_muc_tour dm ON dm.id = t.danh_muc_id
LEFT JOIN nha_cung_cap ncc ON ncc.id = t.nha_cung_cap_id
LEFT JOIN so_thich_nguoi_dung st ON (
    (st.loai_so_thich = 'danh_muc' AND st.gia_tri_id = t.danh_muc_id)
    OR (st.loai_so_thich = 'diem_den' AND st.gia_tri_id IN (
        SELECT diem_den_id FROM tour_diem_den WHERE tour_id = t.id
    ))
) AND st.nguoi_dung_id = $1
LEFT JOIN (
    SELECT tour_id, AVG(diem_danh_gia)::float AS avg_rating, COUNT(*)::int AS total_reviews
    FROM danh_gia
    WHERE dang_hoat_dong = TRUE
    GROUP BY tour_id
) dg ON dg.tour_id = t.id
WHERE t.dang_hoat_dong = TRUE 
    AND t.trang_thai = 'cong_bo'
    AND ($1::uuid IS NULL OR st.nguoi_dung_id IS NOT NULL)
GROUP BY t.id, t.tieu_de, t.mo_ta, t.danh_muc_id, t.so_ngay, t.so_dem, 
         t.gia_nguoi_lon, t.gia_tre_em, t.don_vi_tien_te, t.trang_thai, 
         t.noi_bat, dm.ten, ncc.ten, dg.avg_rating, dg.total_reviews
ORDER BY tong_diem_so_thich DESC, t.noi_bat DESC, avg_rating DESC, t.ngay_tao DESC
LIMIT $2 OFFSET $3;

-- name: GetRecommendedToursByFavoriteDestinations :many
-- Gợi ý tour dựa trên các điểm đến mà người dùng đã yêu thích
WITH SoThichKhachHang AS (
    -- Tìm các địa danh (diem_den) mà khách hàng đã từng đặt hoặc yêu thích
    SELECT DISTINCT dd.id as diem_den_id
    FROM tour_yeu_thich tyf
    JOIN tour_diem_den tdd ON tyf.tour_id = tdd.tour_id
    JOIN diem_den dd ON tdd.diem_den_id = dd.id
    WHERE tyf.nguoi_dung_id = $1
    UNION
    SELECT DISTINCT dd.id as diem_den_id
    FROM dat_cho dc
    JOIN khoi_hanh_tour kh ON dc.khoi_hanh_id = kh.id
    JOIN tour_diem_den tdd ON kh.tour_id = tdd.tour_id
    JOIN diem_den dd ON tdd.diem_den_id = dd.id
    WHERE dc.nguoi_dung_id = $1
)
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
    COUNT(tdd.diem_den_id) as do_phu_hop,
    COALESCE(dg.avg_rating, 0) as avg_rating,
    COALESCE(dg.total_reviews, 0) as total_reviews
FROM tour t
JOIN tour_diem_den tdd ON t.id = tdd.tour_id
LEFT JOIN danh_muc_tour dm ON dm.id = t.danh_muc_id
LEFT JOIN nha_cung_cap ncc ON ncc.id = t.nha_cung_cap_id
LEFT JOIN (
    SELECT tour_id, AVG(diem_danh_gia)::float AS avg_rating, COUNT(*)::int AS total_reviews
    FROM danh_gia
    WHERE dang_hoat_dong = TRUE
    GROUP BY tour_id
) dg ON dg.tour_id = t.id
WHERE tdd.diem_den_id IN (SELECT diem_den_id FROM SoThichKhachHang)
  AND t.dang_hoat_dong = TRUE
  AND t.trang_thai = 'cong_bo'
GROUP BY t.id, t.tieu_de, t.mo_ta, t.danh_muc_id, t.so_ngay, t.so_dem, 
         t.gia_nguoi_lon, t.gia_tre_em, t.don_vi_tien_te, t.trang_thai, 
         t.noi_bat, dm.ten, ncc.ten, dg.avg_rating, dg.total_reviews
ORDER BY do_phu_hop DESC, t.noi_bat DESC, avg_rating DESC
LIMIT $2 OFFSET $3;

-- name: GetRecommendedToursByViewHistory :many
-- Gợi ý tour dựa trên lịch sử xem (ưu tiên tour xem lâu)
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
    AVG(lsx.thoi_luong_xem_giay) as thoi_luong_xem_trung_binh,
    COUNT(lsx.id) as so_lan_xem,
    COALESCE(dg.avg_rating, 0) as avg_rating,
    COALESCE(dg.total_reviews, 0) as total_reviews
FROM tour t
JOIN lich_su_xem_tour lsx ON t.id = lsx.tour_id
LEFT JOIN danh_muc_tour dm ON dm.id = t.danh_muc_id
LEFT JOIN nha_cung_cap ncc ON ncc.id = t.nha_cung_cap_id
LEFT JOIN (
    SELECT tour_id, AVG(diem_danh_gia)::float AS avg_rating, COUNT(*)::int AS total_reviews
    FROM danh_gia
    WHERE dang_hoat_dong = TRUE
    GROUP BY tour_id
) dg ON dg.tour_id = t.id
WHERE lsx.nguoi_dung_id = $1
  AND t.dang_hoat_dong = TRUE
  AND t.trang_thai = 'cong_bo'
  AND lsx.thoi_luong_xem_giay > 30 -- Chỉ lấy tour xem hơn 30 giây
GROUP BY t.id, t.tieu_de, t.mo_ta, t.danh_muc_id, t.so_ngay, t.so_dem, 
         t.gia_nguoi_lon, t.gia_tre_em, t.don_vi_tien_te, t.trang_thai, 
         t.noi_bat, dm.ten, ncc.ten, dg.avg_rating, dg.total_reviews
ORDER BY thoi_luong_xem_trung_binh DESC, so_lan_xem DESC, t.noi_bat DESC
LIMIT $2 OFFSET $3;

