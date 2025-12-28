-- name: CreateUser :one
insert into nguoi_dung(ho_ten, email, mat_khau_ma_hoa, so_dien_thoai, vai_tro, dang_hoat_dong, xac_thuc, ngay_tao, ngay_cap_nhat)
values
    ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    returning *;
-- name: GetUserByEmail :one
select * from nguoi_dung where email = $1;
-- name: GetUserById :one
SELECT 
    nguoi_dung.*, 
    COUNT(dat_cho.id) AS tong_dat_cho, 
    COUNT(dat_cho.id) FILTER (WHERE dat_cho.trang_thai = 'da_thanh_toan') AS tong_dat_cho_da_thanh_toan, 
    COUNT(dat_cho.id) FILTER (WHERE dat_cho.trang_thai = 'cho_xac_nhan') AS tong_dat_cho_dang_cho_xac_nhan
FROM nguoi_dung
LEFT JOIN dat_cho ON dat_cho.nguoi_dung_id = nguoi_dung.id
WHERE nguoi_dung.id = $1
GROUP BY nguoi_dung.id; -- Thêm dòng này
-- name: UpdateUserById :one
update nguoi_dung
set ho_ten = $1, email = $2, so_dien_thoai = $3, ngay_cap_nhat = $4
where id = $5
returning *;


SELECT 
    nguoi_dung.*,
    nha_cung_cap.id AS nha_cung_cap_id
FROM nguoi_dung
LEFT JOIN nha_cung_cap 
    ON nha_cung_cap.id = nguoi_dung.id
    AND nguoi_dung.vai_tro = 'nha_cung_cap'
WHERE nguoi_dung.email = $1;

-- (
--     SELECT nguoi_dung.*, nha_cung_cap.id AS nha_cung_cap_id
--     FROM nguoi_dung
--     JOIN nha_cung_cap 
--         ON nha_cung_cap.id = nguoi_dung.id
--     WHERE nguoi_dung.email = $1
--       AND nguoi_dung.vai_tro = 'nha_cung_cap'
-- )
-- UNION ALL
-- (
--     SELECT nguoi_dung.*, NULL AS nha_cung_cap_id
--     FROM nguoi_dung
--     WHERE nguoi_dung.email = $1
--       AND nguoi_dung.vai_tro IN ('khach_hang', 'quan_tri')
-- );
