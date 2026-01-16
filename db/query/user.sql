-- name: ChangePassword :exec
UPDATE nguoi_dung
SET mat_khau_ma_hoa = $1
WHERE id = $2;

-- name: ForgotPassword :exec
UPDATE nguoi_dung
SET mat_khau_ma_hoa = $1
WHERE email = $2
RETURNING *;

-- name: ResetPassword :one
UPDATE nguoi_dung
SET mat_khau_ma_hoa = $1
WHERE email = $2
RETURNING *;

-- name: CreatePasswordResetOTP :one
INSERT INTO otp_dat_lai_mat_khau (nguoi_dung_id, ma_otp, thoi_han)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetPasswordResetOTP :one
SELECT otp.* FROM otp_dat_lai_mat_khau otp
INNER JOIN nguoi_dung nd ON otp.nguoi_dung_id = nd.id
WHERE nd.email = $1 AND otp.ma_otp = $2 AND otp.thoi_han > NOW()
ORDER BY otp.ngay_tao DESC
LIMIT 1;

-- name: GetUnverifiedPasswordResetOTP :one
SELECT otp.* FROM otp_dat_lai_mat_khau otp
INNER JOIN nguoi_dung nd ON otp.nguoi_dung_id = nd.id
WHERE nd.email = $1 AND otp.ma_otp = $2 AND otp.da_xac_thuc = FALSE AND otp.thoi_han > NOW()
ORDER BY otp.ngay_tao DESC
LIMIT 1;

-- name: VerifyPasswordResetOTP :exec
UPDATE otp_dat_lai_mat_khau otp
SET da_xac_thuc = TRUE, ngay_xac_thuc = CURRENT_TIMESTAMP
FROM nguoi_dung nd
WHERE otp.nguoi_dung_id = nd.id 
  AND nd.email = $1 
  AND otp.ma_otp = $2 
  AND otp.da_xac_thuc = FALSE 
  AND otp.thoi_han > NOW();

-- name: InvalidateAllOTPsForEmail :exec
UPDATE otp_dat_lai_mat_khau otp
SET da_xac_thuc = TRUE
FROM nguoi_dung nd
WHERE otp.nguoi_dung_id = nd.id 
  AND nd.email = $1 
  AND otp.da_xac_thuc = FALSE;

-- name: DeleteExpiredOTPs :exec
DELETE FROM otp_dat_lai_mat_khau
WHERE thoi_han < NOW();


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

-- Function để xóa các OTP đã hết hạn
CREATE OR REPLACE FUNCTION xoa_otp_het_han()
RETURNS INTEGER AS $$
DECLARE
    so_ban_ghi_xoa INTEGER;
BEGIN
    DELETE FROM otp_dat_lai_mat_khau 
    WHERE thoi_han < NOW();
    
    GET DIAGNOSTICS so_ban_ghi_xoa = ROW_COUNT;
    RETURN so_ban_ghi_xoa;
END;
$$ LANGUAGE plpgsql;

-- Có thể tạo scheduled job để chạy function này định kỳ
-- Ví dụ: SELECT cron.schedule('xoa-otp-het-han', '0 * * * *', 'SELECT xoa_otp_het_han()');
-- (Chạy mỗi giờ một lần nếu có extension pg_cron)