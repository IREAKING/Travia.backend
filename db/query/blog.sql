-- Blog API endpoints structure
-- File: db/query/blog.sql

-- ==================== BLOG CATEGORIES ====================

-- Get all active blog categories
-- name: GetAllBlogCategories
-- SELECT id, ten, slug, mo_ta, anh, mau_sac 
-- FROM danh_muc_blog 
-- WHERE dang_hoat_dong = true 
-- ORDER BY ten;

-- Get blog category by slug
-- name: GetBlogCategoryBySlug
-- SELECT id, ten, slug, mo_ta, anh, mau_sac 
-- FROM danh_muc_blog 
-- WHERE slug = $1 AND dang_hoat_dong = true;

-- ==================== BLOG POSTS ====================

-- Get published blog posts with pagination
-- name: GetPublishedBlogPosts
-- SELECT 
--     b.id, b.tieu_de, b.slug, b.tom_tat, b.anh_dai_dien,
--     b.luot_xem, b.luot_thich, b.ngay_cong_bo,
--     c.ten as danh_muc_ten, c.slug as danh_muc_slug, c.mau_sac as danh_muc_mau,
--     u.ho_ten as tac_gia_ten
-- FROM bai_viet_blog b
-- LEFT JOIN danh_muc_blog c ON b.danh_muc_id = c.id
-- LEFT JOIN nguoi_dung u ON b.tac_gia_id = u.id
-- WHERE b.trang_thai = 'cong_bo'
-- ORDER BY b.ngay_cong_bo DESC
-- LIMIT $1 OFFSET $2;

-- Get featured blog posts
-- name: GetFeaturedBlogPosts
-- SELECT 
--     b.id, b.tieu_de, b.slug, b.tom_tat, b.anh_dai_dien,
--     b.luot_xem, b.luot_thich, b.ngay_cong_bo,
--     c.ten as danh_muc_ten, c.slug as danh_muc_slug,
--     u.ho_ten as tac_gia_ten
-- FROM bai_viet_blog b
-- LEFT JOIN danh_muc_blog c ON b.danh_mac_id = c.id
-- LEFT JOIN nguoi_dung u ON b.tac_gia_id = u.id
-- WHERE b.trang_thai = 'cong_bo' AND b.noi_bat = true
-- ORDER BY b.ngay_cong_bo DESC
-- LIMIT $1;

-- Get blog post by slug with full details
-- name: GetBlogPostBySlug
-- SELECT 
--     b.id, b.tieu_de, b.slug, b.tom_tat, b.noi_dung, b.anh_dai_dien,
--     b.luot_xem, b.luot_thich, b.ngay_cong_bo,
--     c.id as danh_mac_id, c.ten as danh_mac_ten, c.slug as danh_mac_slug,
--     u.id as tac_gia_id, u.ho_ten as tac_gia_ten, u.email as tac_gia_email
-- FROM bai_viet_blog b
-- LEFT JOIN danh_muc_blog c ON b.danh_mac_id = c.id
-- LEFT JOIN nguoi_dung u ON b.tac_gia_id = u.id
-- WHERE b.slug = $1 AND b.trang_thai = 'cong_bo';

-- Search blog posts
-- name: SearchBlogPosts
-- SELECT 
--     b.id, b.tieu_de, b.slug, b.tom_tat, b.anh_dai_dien,
--     b.luot_xem, b.luot_thich, b.ngay_cong_bo,
--     c.ten as danh_mac_ten, c.slug as danh_mac_slug,
--     u.ho_ten as tac_gia_ten,
--     ts_rank(b.search_vector, plainto_tsquery('vietnamese', $1)) as rank
-- FROM bai_viet_blog b
-- LEFT JOIN danh_mac_blog c ON b.danh_mac_id = c.id
-- LEFT JOIN nguoi_dung u ON b.tac_gia_id = u.id
-- WHERE b.trang_thai = 'cong_bo' 
--   AND b.search_vector @@ plainto_tsquery('vietnamese', $1)
-- ORDER BY rank DESC, b.ngay_cong_bo DESC
-- LIMIT $2 OFFSET $3;

-- Get blog posts by category
-- name: GetBlogPostsByCategory
-- SELECT 
--     b.id, b.tieu_de, b.slug, b.tom_tat, b.anh_dai_dien,
--     b.luot_xem, b.luot_thich, b.ngay_cong_bo,
--     c.ten as danh_mac_ten, c.slug as danh_mac_slug,
--     u.ho_ten as tac_gia_ten
-- FROM bai_viet_blog b
-- LEFT JOIN danh_mac_blog c ON b.danh_mac_id = c.id
-- LEFT JOIN nguoi_dung u ON b.tac_gia_id = u.id
-- WHERE b.trang_thai = 'cong_bo' AND c.slug = $1
-- ORDER BY b.ngay_cong_bo DESC
-- LIMIT $2 OFFSET $3;

-- Increment blog post view count
-- name: IncrementBlogPostViews
-- UPDATE bai_viet_blog 
-- SET luot_xem = luot_xem + 1 
-- WHERE id = $1;

-- ==================== BLOG TAGS ====================

-- Get all blog tags
-- name: GetAllBlogTags
-- SELECT id, ten, slug, mau_sac 
-- FROM the_blog 
-- ORDER BY ten;

-- Get tags for a blog post
-- name: GetBlogPostTags
-- SELECT t.id, t.ten, t.slug, t.mau_sac
-- FROM the_blog t
-- JOIN bai_viet_the bt ON t.id = bt.the_id
-- WHERE bt.bai_viet_id = $1;

-- ==================== BLOG COMMENTS ====================

-- Get approved comments for a blog post
-- name: GetBlogPostComments
-- SELECT 
--     bc.id, bc.noi_dung, bc.ngay_tao,
--     u.ho_ten as nguoi_dung_ten, u.email as nguoi_dung_email,
--     bc.phan_hoi_id
-- FROM binh_luan_blog bc
-- JOIN nguoi_dung u ON bc.nguoi_dung_id = u.id
-- WHERE bc.bai_viet_id = $1 AND bc.trang_thai = 'da_duyet'
-- ORDER BY bc.ngay_tao ASC;

-- Create blog comment
-- name: CreateBlogComment
-- INSERT INTO binh_luan_blog (bai_viet_id, nguoi_dung_id, noi_dung, phan_hoi_id)
-- VALUES ($1, $2, $3, $4)
-- RETURNING id, ngay_tao;

-- ==================== ADMIN BLOG MANAGEMENT ====================

-- Get all blog posts for admin (with status)
-- name: GetAllBlogPostsAdmin
-- SELECT 
--     b.id, b.tieu_de, b.slug, b.trang_thai, b.noi_bat,
--     b.luot_xem, b.luot_thich, b.ngay_cong_bo, b.ngay_tao,
--     c.ten as danh_mac_ten,
--     u.ho_ten as tac_gia_ten
-- FROM bai_viet_blog b
-- LEFT JOIN danh_mac_blog c ON b.danh_mac_id = c.id
-- LEFT JOIN nguoi_dung u ON b.tac_gia_id = u.id
-- ORDER BY b.ngay_tao DESC
-- LIMIT $1 OFFSET $2;

-- Create blog post
-- name: CreateBlogPost
-- INSERT INTO bai_viet_blog (
--     tieu_de, slug, tom_tat, noi_dung, anh_dai_dien,
--     danh_mac_id, tac_gia_id, trang_thai, noi_bat
-- ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
-- RETURNING id, ngay_tao;

-- Update blog post
-- name: UpdateBlogPost
-- UPDATE bai_viet_blog 
-- SET tieu_de = $2, slug = $3, tom_tat = $4, noi_dung = $5,
--     anh_dai_dien = $6, danh_mac_id = $7, trang_thai = $8,
--     noi_bat = $9, ngay_cap_nhat = CURRENT_TIMESTAMP
-- WHERE id = $1
-- RETURNING ngay_cap_nhat;

-- Delete blog post
-- name: DeleteBlogPost
-- DELETE FROM bai_viet_blog WHERE id = $1;
