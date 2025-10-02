-- name: CreateUser :one
insert into nguoi_dung(ho_ten, email, mat_khau_ma_hoa, so_dien_thoai, vai_tro, dang_hoat_dong, xac_thuc, ngay_tao, ngay_cap_nhat)
values
    ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    returning *;
-- name: GetUserByEmail :one
select * from nguoi_dung
where email = $1;
-- name: GetUserById :one
select * from nguoi_dung
where id = $1;
-- name: UpdateUserById :one
update nguoi_dung
set ho_ten = $1, email = $2, mat_khau_ma_hoa = $3, so_dien_thoai = $4, ngay_cap_nhat = $5
where id = $6
returning *;