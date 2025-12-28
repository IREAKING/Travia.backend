-- name: ChangePassword :exec
UPDATE nguoi_dung
SET mat_khau_ma_hoa = $1
WHERE id = $2;

-- name: ForgotPassword :one
UPDATE nguoi_dung
SET mat_khau_ma_hoa = $1
WHERE email = $2
RETURNING *;

-- name: ResetPassword :one
UPDATE nguoi_dung
SET mat_khau_ma_hoa = $1
WHERE email = $2
RETURNING *;

-- name: UpdateUser :one
UPDATE nguoi_dung
SET ho_ten = COALESCE(sqlc.narg(ho_ten), nguoi_dung.ho_ten),
    email = COALESCE(sqlc.narg(email), nguoi_dung.email),
    so_dien_thoai = COALESCE(sqlc.narg(so_dien_thoai), nguoi_dung.so_dien_thoai),
    ngay_cap_nhat = CURRENT_TIMESTAMP
WHERE id = $1 AND nguoi_dung.dang_hoat_dong = TRUE
RETURNING *;

-- Note: yeu_thich table doesn't exist yet, will implement in Sprint 3-4
-- -- name: GetUserFavorites :many
-- SELECT * FROM yeu_thich
-- WHERE nguoi_dung_id = $1;

-- name: GetUserPaymentHistory :many
