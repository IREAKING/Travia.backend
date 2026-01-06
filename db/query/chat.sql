-- name: CreateChatHistory :one
INSERT INTO lich_su_chat (
    nguoi_dung_id,
    ma_phien,
    cau_hoi,
    cau_tra_loi
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetChatHistoryByUserID :many
SELECT * FROM lich_su_chat
WHERE nguoi_dung_id = $1
ORDER BY ngay_tao DESC
LIMIT $2 OFFSET $3;

-- name: GetChatHistoryBySessionID :many
SELECT * FROM lich_su_chat
WHERE ma_phien = $1
ORDER BY ngay_tao DESC
LIMIT $2 OFFSET $3;

-- name: GetRecentChatHistory :many
SELECT * FROM lich_su_chat
WHERE (($1::uuid IS NOT NULL AND nguoi_dung_id = $1) OR ($1::uuid IS NULL AND ma_phien = $2))
ORDER BY ngay_tao DESC
LIMIT $3;

-- name: DeleteChatHistoryBySessionID :exec
DELETE FROM lich_su_chat
WHERE ma_phien = $1;

-- name: DeleteChatHistoryByUserID :exec
DELETE FROM lich_su_chat
WHERE nguoi_dung_id = $1;

