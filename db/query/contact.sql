-- name: CreateContact :one
INSERT INTO lien_he (
    ho_ten,
    email,
    so_dien_thoai,
    tieu_de,
    noi_dung,
    nguoi_dung_id
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetAllContacts :many
SELECT 
    lh.*,
    nd.ho_ten AS ten_nguoi_dung,
    nd.email AS email_nguoi_dung
FROM lien_he lh
LEFT JOIN nguoi_dung nd ON lh.nguoi_dung_id = nd.id
ORDER BY lh.ngay_tao DESC
LIMIT $1 OFFSET $2;

-- name: GetContactByID :one
SELECT 
    lh.*,
    nd.ho_ten AS ten_nguoi_dung,
    nd.email AS email_nguoi_dung
FROM lien_he lh
LEFT JOIN nguoi_dung nd ON lh.nguoi_dung_id = nd.id
WHERE lh.id = $1;

-- name: UpdateContactStatus :one
UPDATE lien_he
SET 
    trang_thai = $2,
    da_doc = $3,
    ngay_cap_nhat = NOW()
WHERE id = $1
RETURNING *;

-- name: GetContactsByStatus :many
SELECT 
    lh.*,
    nd.ho_ten AS ten_nguoi_dung,
    nd.email AS email_nguoi_dung
FROM lien_he lh
LEFT JOIN nguoi_dung nd ON lh.nguoi_dung_id = nd.id
WHERE lh.trang_thai = $1
ORDER BY lh.ngay_tao DESC
LIMIT $2 OFFSET $3;

-- name: CountContacts :one
SELECT COUNT(*) FROM lien_he;

-- name: CountContactsByStatus :one
SELECT COUNT(*) FROM lien_he WHERE trang_thai = $1;

-- name: GetUnreadContacts :many
SELECT 
    lh.*,
    nd.ho_ten AS ten_nguoi_dung,
    nd.email AS email_nguoi_dung
FROM lien_he lh
LEFT JOIN nguoi_dung nd ON lh.nguoi_dung_id = nd.id
WHERE lh.da_doc = FALSE
ORDER BY lh.ngay_tao DESC
LIMIT $1 OFFSET $2;

-- name: MarkContactAsRead :one
UPDATE lien_he
SET 
    da_doc = TRUE,
    ngay_cap_nhat = NOW()
WHERE id = $1
RETURNING *;

-- ===========================================
-- PHẢN HỒI LIÊN HỆ
-- ===========================================

-- name: CreateContactResponse :one
-- Tạo phản hồi cho liên hệ
INSERT INTO phan_hoi_lien_he (
    lien_he_id,
    nguoi_phan_hoi_id,
    noi_dung
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetContactResponses :many
-- Lấy tất cả phản hồi của một liên hệ
SELECT 
    ph.*,
    nd.ho_ten AS ten_nguoi_phan_hoi,
    nd.email AS email_nguoi_phan_hoi
FROM phan_hoi_lien_he ph
JOIN nguoi_dung nd ON ph.nguoi_phan_hoi_id = nd.id
WHERE ph.lien_he_id = $1
ORDER BY ph.ngay_tao ASC;

-- name: UpdateContactResponse :one
-- Cập nhật phản hồi
UPDATE phan_hoi_lien_he
SET 
    noi_dung = $2,
    ngay_cap_nhat = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteContactResponse :exec
-- Xóa phản hồi
DELETE FROM phan_hoi_lien_he WHERE id = $1;

-- name: GetContactResponseByID :one
-- Lấy phản hồi theo ID
SELECT 
    ph.*,
    nd.ho_ten AS ten_nguoi_phan_hoi,
    nd.email AS email_nguoi_phan_hoi
FROM phan_hoi_lien_he ph
JOIN nguoi_dung nd ON ph.nguoi_phan_hoi_id = nd.id
WHERE ph.id = $1;

