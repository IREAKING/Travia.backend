-- ===========================================
-- BLOG QUERIES
-- ===========================================

-- name: CreateBlog :one
INSERT INTO blog (
    tieu_de,
    slug,
    tom_tat,
    noi_dung,
    anh_dai_dien,
    tac_gia_id,
    danh_muc,
    tu_khoa,
    trang_thai,
    noi_bat,
    ngay_dang
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING *;

-- name: GetBlogByID :one
SELECT 
    b.*,
    nd.ho_ten AS ten_tac_gia,
    nd.email AS email_tac_gia
FROM blog b
LEFT JOIN nguoi_dung nd ON b.tac_gia_id = nd.id
WHERE b.id = $1;

-- name: GetBlogBySlug :one
SELECT 
    b.*,
    nd.ho_ten AS ten_tac_gia,
    nd.email AS email_tac_gia
FROM blog b
LEFT JOIN nguoi_dung nd ON b.tac_gia_id = nd.id
WHERE b.slug = $1;

-- name: GetAllBlogs :many
SELECT 
    b.*,
    nd.ho_ten AS ten_tac_gia,
    nd.email AS email_tac_gia
FROM blog b
LEFT JOIN nguoi_dung nd ON b.tac_gia_id = nd.id
ORDER BY 
    CASE WHEN b.noi_bat = TRUE THEN 0 ELSE 1 END,
    COALESCE(b.ngay_dang, b.ngay_tao) DESC
LIMIT $1 OFFSET $2;

-- name: GetPublishedBlogs :many
SELECT 
    b.*,
    nd.ho_ten AS ten_tac_gia,
    nd.email AS email_tac_gia
FROM blog b
LEFT JOIN nguoi_dung nd ON b.tac_gia_id = nd.id
WHERE b.trang_thai = 'cong_bo'
    AND (b.ngay_dang IS NULL OR b.ngay_dang <= NOW())
ORDER BY 
    CASE WHEN b.noi_bat = TRUE THEN 0 ELSE 1 END,
    COALESCE(b.ngay_dang, b.ngay_tao) DESC
LIMIT $1 OFFSET $2;

-- name: GetBlogsByCategory :many
SELECT 
    b.*,
    nd.ho_ten AS ten_tac_gia,
    nd.email AS email_tac_gia
FROM blog b
LEFT JOIN nguoi_dung nd ON b.tac_gia_id = nd.id
WHERE b.danh_muc = $1
    AND b.trang_thai = 'cong_bo'
    AND (b.ngay_dang IS NULL OR b.ngay_dang <= NOW())
ORDER BY 
    CASE WHEN b.noi_bat = TRUE THEN 0 ELSE 1 END,
    COALESCE(b.ngay_dang, b.ngay_tao) DESC
LIMIT $2 OFFSET $3;

-- name: GetFeaturedBlogs :many
SELECT 
    b.*,
    nd.ho_ten AS ten_tac_gia,
    nd.email AS email_tac_gia
FROM blog b
LEFT JOIN nguoi_dung nd ON b.tac_gia_id = nd.id
WHERE b.noi_bat = TRUE
    AND b.trang_thai = 'cong_bo'
    AND (b.ngay_dang IS NULL OR b.ngay_dang <= NOW())
ORDER BY COALESCE(b.ngay_dang, b.ngay_tao) DESC
LIMIT $1;

-- name: SearchBlogs :many
SELECT 
    b.*,
    nd.ho_ten AS ten_tac_gia,
    nd.email AS email_tac_gia,
    ts_rank(
        to_tsvector('vietnamese', COALESCE(b.tieu_de, '') || ' ' || COALESCE(b.tom_tat, '') || ' ' || COALESCE(b.noi_dung, '')),
        plainto_tsquery('vietnamese', $1)
    ) AS rank
FROM blog b
LEFT JOIN nguoi_dung nd ON b.tac_gia_id = nd.id
WHERE b.trang_thai = 'cong_bo'
    AND (b.ngay_dang IS NULL OR b.ngay_dang <= NOW())
    AND (
        to_tsvector('vietnamese', COALESCE(b.tieu_de, '') || ' ' || COALESCE(b.tom_tat, '') || ' ' || COALESCE(b.noi_dung, '')) 
        @@ plainto_tsquery('vietnamese', $1)
        OR b.tieu_de ILIKE '%' || $1 || '%'
        OR b.tom_tat ILIKE '%' || $1 || '%'
    )
ORDER BY rank DESC, COALESCE(b.ngay_dang, b.ngay_tao) DESC
LIMIT $2 OFFSET $3;

-- name: GetBlogsByAuthor :many
SELECT 
    b.*,
    nd.ho_ten AS ten_tac_gia,
    nd.email AS email_tac_gia
FROM blog b
LEFT JOIN nguoi_dung nd ON b.tac_gia_id = nd.id
WHERE b.tac_gia_id = $1
ORDER BY COALESCE(b.ngay_dang, b.ngay_tao) DESC
LIMIT $2 OFFSET $3;

-- name: UpdateBlog :one
UPDATE blog
SET 
    tieu_de = COALESCE($2, tieu_de),
    slug = COALESCE($3, slug),
    tom_tat = COALESCE($4, tom_tat),
    noi_dung = COALESCE($5, noi_dung),
    anh_dai_dien = COALESCE($6, anh_dai_dien),
    danh_muc = COALESCE($7, danh_muc),
    tu_khoa = COALESCE($8, tu_khoa),
    trang_thai = COALESCE($9, trang_thai),
    noi_bat = COALESCE($10, noi_bat),
    ngay_dang = COALESCE($11, ngay_dang),
    ngay_cap_nhat = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteBlog :exec
DELETE FROM blog WHERE id = $1;

-- name: IncrementBlogViews :exec
UPDATE blog
SET luot_xem = luot_xem + 1
WHERE id = $1;

-- name: IncrementBlogLikes :exec
UPDATE blog
SET luot_thich = luot_thich + 1
WHERE id = $1;

-- name: DecrementBlogLikes :exec
UPDATE blog
SET luot_thich = GREATEST(0, luot_thich - 1)
WHERE id = $1;

-- name: GetRelatedBlogs :many
SELECT 
    b.*,
    nd.ho_ten AS ten_tac_gia,
    nd.email AS email_tac_gia
FROM blog b
LEFT JOIN nguoi_dung nd ON b.tac_gia_id = nd.id
WHERE b.id != $1
    AND b.trang_thai = 'cong_bo'
    AND (b.ngay_dang IS NULL OR b.ngay_dang <= NOW())
    AND (
        b.danh_muc = (SELECT danh_muc FROM blog WHERE id = $1)
        OR b.tu_khoa && (SELECT tu_khoa FROM blog WHERE id = $1)
    )
ORDER BY COALESCE(b.ngay_dang, b.ngay_tao) DESC
LIMIT $2;

-- name: GetBlogStats :one
SELECT 
    COUNT(*) FILTER (WHERE trang_thai = 'cong_bo') AS tong_so_da_dang,
    COUNT(*) FILTER (WHERE trang_thai = 'nhap') AS tong_so_nhap,
    COUNT(*) FILTER (WHERE trang_thai = 'luu_tru') AS tong_so_luu_tru,
    COUNT(*) FILTER (WHERE noi_bat = TRUE) AS tong_so_noi_bat,
    SUM(luot_xem) AS tong_luot_xem,
    SUM(luot_thich) AS tong_luot_thich,
    COUNT(*) FILTER (WHERE ngay_dang >= NOW() - INTERVAL '30 days') AS so_bai_trong_30_ngay
FROM blog;

-- name: CountBlogs :one
SELECT COUNT(*) FROM blog
WHERE ($1::text IS NULL OR trang_thai = $1::text)
    AND ($2::text IS NULL OR danh_muc = $2::text)
    AND ($3::text IS NULL OR tac_gia_id::text = $3::text);

-- name: CountPublishedBlogs :one
SELECT COUNT(*) FROM blog
WHERE trang_thai = 'cong_bo'
    AND (ngay_dang IS NULL OR ngay_dang <= NOW());

-- ===========================================
-- BLOG AI HISTORY QUERIES
-- ===========================================

-- name: CreateBlogAIHistory :one
INSERT INTO lich_su_ai_blog (
    blog_id,
    prompt,
    phan_hoi_ai,
    mo_hinh_ai,
    so_luong_token
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetBlogAIHistory :many
SELECT * FROM lich_su_ai_blog
WHERE blog_id = $1
ORDER BY ngay_tao DESC;

-- ===========================================
-- BLOG COMMENT QUERIES
-- ===========================================

-- name: CreateBlogComment :one
INSERT INTO binh_luan_blog (
    blog_id,
    nguoi_dung_id,
    noi_dung,
    binh_luan_cha_id,
    da_duyet
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetBlogComments :many
SELECT 
    bl.*,
    nd.ho_ten AS ten_nguoi_dung,
    nd.email AS email_nguoi_dung
FROM binh_luan_blog bl
LEFT JOIN nguoi_dung nd ON bl.nguoi_dung_id = nd.id
WHERE bl.blog_id = $1
    AND bl.da_duyet = TRUE
    AND bl.binh_luan_cha_id IS NULL -- Chỉ lấy comment gốc
ORDER BY bl.ngay_tao DESC
LIMIT $2 OFFSET $3;

-- name: GetBlogCommentReplies :many
SELECT 
    bl.*,
    nd.ho_ten AS ten_nguoi_dung,
    nd.email AS email_nguoi_dung
FROM binh_luan_blog bl
LEFT JOIN nguoi_dung nd ON bl.nguoi_dung_id = nd.id
WHERE bl.binh_luan_cha_id = $1
    AND bl.da_duyet = TRUE
ORDER BY bl.ngay_tao ASC;

-- name: GetPendingComments :many
SELECT 
    bl.*,
    nd.ho_ten AS ten_nguoi_dung,
    nd.email AS email_nguoi_dung,
    b.tieu_de AS ten_blog
FROM binh_luan_blog bl
LEFT JOIN nguoi_dung nd ON bl.nguoi_dung_id = nd.id
LEFT JOIN blog b ON bl.blog_id = b.id
WHERE bl.da_duyet = FALSE
ORDER BY bl.ngay_tao DESC
LIMIT $1 OFFSET $2;

-- name: ApproveComment :one
UPDATE binh_luan_blog
SET 
    da_duyet = TRUE,
    ngay_cap_nhat = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteComment :exec
DELETE FROM binh_luan_blog WHERE id = $1;

-- name: CountBlogComments :one
SELECT COUNT(*) FROM binh_luan_blog
WHERE blog_id = $1
    AND da_duyet = TRUE;
