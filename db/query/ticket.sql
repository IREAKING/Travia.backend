-- name: GetTicketsByBookingID :many
SELECT 
    hk.dat_cho_id, 
    dc.tong_tien, 
    dc.don_vi_tien_te,
    hk.ho_ten, 
    hk.ngay_sinh, 
    hk.loai_khach, 
    hk.so_giay_to_tuy_thanh, 
    hk.quoc_tich,
    t.tieu_de AS ten_tour,
    kh.ngay_khoi_hanh, 
    kh.ngay_ket_thuc
FROM hanh_khach hk
JOIN dat_cho dc ON hk.dat_cho_id = dc.id
JOIN khoi_hanh_tour kh ON dc.khoi_hanh_id = kh.id
JOIN tour t ON kh.tour_id = t.id
WHERE dc.id = $1;