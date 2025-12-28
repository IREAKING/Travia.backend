-- name: CreateFavoriteTour :one
INSERT INTO tour_yeu_thich (
    nguoi_dung_id,
    tour_id
) VALUES (
    $1, $2
)
RETURNING *;

-- name: DeleteFavoriteTour :exec
DELETE FROM tour_yeu_thich
WHERE nguoi_dung_id = $1 AND tour_id = $2;

-- name: GetFavoriteTours :many
SELECT * FROM tour_yeu_thich
WHERE nguoi_dung_id = $1;
