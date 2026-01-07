-- ===========================================
-- THÔNG BÁO (NOTIFICATIONS)
-- ===========================================

-- name: CreateNotification :one
-- Tạo thông báo mới
INSERT INTO thong_bao (
    nguoi_dung_id,
    tieu_de,
    noi_dung,
    loai,
    lien_ket
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetNotificationsByUser :many
-- Lấy thông báo của người dùng (có phân trang)
SELECT * FROM thong_bao
WHERE nguoi_dung_id = $1
ORDER BY ngay_tao DESC
LIMIT $2 OFFSET $3;

-- name: GetUnreadNotificationsByUser :many
-- Lấy thông báo chưa đọc của người dùng
SELECT * FROM thong_bao
WHERE nguoi_dung_id = $1 AND da_doc = FALSE
ORDER BY ngay_tao DESC
LIMIT $2 OFFSET $3;

-- name: CountUnreadNotificationsByUser :one
-- Đếm số thông báo chưa đọc
SELECT COUNT(*)::int FROM thong_bao
WHERE nguoi_dung_id = $1 AND da_doc = FALSE;

-- name: MarkNotificationAsRead :one
-- Đánh dấu thông báo đã đọc
UPDATE thong_bao
SET da_doc = TRUE
WHERE id = $1
RETURNING *;

-- name: MarkAllNotificationsAsRead :exec
-- Đánh dấu tất cả thông báo của user đã đọc
UPDATE thong_bao
SET da_doc = TRUE
WHERE nguoi_dung_id = $1 AND da_doc = FALSE;

-- name: DeleteNotification :exec
-- Xóa thông báo
DELETE FROM thong_bao WHERE id = $1;

-- name: GetNotificationByID :one
-- Lấy thông báo theo ID
SELECT * FROM thong_bao WHERE id = $1;

-- name: CreateNotificationForContactResponse :one
-- Tạo thông báo khi admin phản hồi liên hệ
-- Function này sẽ được gọi tự động khi có phản hồi
INSERT INTO thong_bao (
    nguoi_dung_id,
    tieu_de,
    noi_dung,
    loai,
    lien_ket
) 
SELECT 
    lh.nguoi_dung_id,
    'Phản hồi liên hệ: ' || lh.tieu_de,
    'Chúng tôi đã phản hồi yêu cầu của bạn: ' || ph.noi_dung,
    'system',
    '/contact'
FROM lien_he lh
JOIN phan_hoi_lien_he ph ON ph.lien_he_id = lh.id
WHERE lh.id = sqlc.arg('lien_he_id') AND ph.id = sqlc.arg('phan_hoi_id') AND lh.nguoi_dung_id IS NOT NULL
RETURNING *;

