-- ==================== PAYMENT/TRANSACTION QUERIES ====================

-- name: GetAllTransactions :many
-- Lấy tất cả giao dịch với thông tin đầy đủ (dành cho Admin)
SELECT 
    lsgd.id,
    lsgd.dat_cho_id,
    lsgd.nguoi_dung_id,
    lsgd.ma_giao_dich_noi_bo,
    lsgd.ma_tham_chieu_cong_thanh_toan,
    lsgd.cong_thanh_toan_id,
    lsgd.so_tien,
    lsgd.loai_giao_dich,
    lsgd.trang_thai,
    lsgd.noi_dung_chuyen_khoan,
    lsgd.ngay_tao,
    lsgd.ngay_hoan_thanh,
    -- Thông tin booking
    dc.phuong_thuc_thanh_toan,
    -- Thông tin user
    nd.ho_ten AS ten_nguoi_dung,
    nd.email AS email_nguoi_dung,
    -- Thông tin tour
    t.tieu_de AS ten_tour,
    t.id AS tour_id,
    -- Thông tin cổng thanh toán
    ctt.ten_hien_thi AS ten_cong_thanh_toan
FROM lich_su_giao_dich lsgd
LEFT JOIN dat_cho dc ON dc.id = lsgd.dat_cho_id
LEFT JOIN nguoi_dung nd ON nd.id = lsgd.nguoi_dung_id
LEFT JOIN khoi_hanh_tour kh ON kh.id = dc.khoi_hanh_id
LEFT JOIN tour t ON t.id = kh.tour_id
LEFT JOIN cong_thanh_toan ctt ON ctt.id = lsgd.cong_thanh_toan_id
ORDER BY lsgd.ngay_tao DESC
LIMIT $1 OFFSET $2;

-- name: CountAllTransactions :one
-- Đếm tổng số giao dịch
SELECT COUNT(*) FROM lich_su_giao_dich;

-- name: GetTransactionsByStatus :many
-- Lấy giao dịch theo trạng thái
SELECT 
    lsgd.id,
    lsgd.dat_cho_id,
    lsgd.nguoi_dung_id,
    lsgd.ma_giao_dich_noi_bo,
    lsgd.ma_tham_chieu_cong_thanh_toan,
    lsgd.cong_thanh_toan_id,
    lsgd.so_tien,
    lsgd.loai_giao_dich,
    lsgd.trang_thai,
    lsgd.noi_dung_chuyen_khoan,
    lsgd.ngay_tao,
    lsgd.ngay_hoan_thanh,
    dc.phuong_thuc_thanh_toan,
    nd.ho_ten AS ten_nguoi_dung,
    nd.email AS email_nguoi_dung,
    t.tieu_de AS ten_tour,
    t.id AS tour_id,
    ctt.ten_hien_thi AS ten_cong_thanh_toan
FROM lich_su_giao_dich lsgd
LEFT JOIN dat_cho dc ON dc.id = lsgd.dat_cho_id
LEFT JOIN nguoi_dung nd ON nd.id = lsgd.nguoi_dung_id
LEFT JOIN khoi_hanh_tour kh ON kh.id = dc.khoi_hanh_id
LEFT JOIN tour t ON t.id = kh.tour_id
LEFT JOIN cong_thanh_toan ctt ON ctt.id = lsgd.cong_thanh_toan_id
WHERE lsgd.trang_thai = $1::trang_thai_thanh_toan
ORDER BY lsgd.ngay_tao DESC
LIMIT $2 OFFSET $3;

-- name: GetTransactionsByPaymentGateway :many
-- Lấy giao dịch theo cổng thanh toán
SELECT 
    lsgd.id,
    lsgd.dat_cho_id,
    lsgd.nguoi_dung_id,
    lsgd.ma_giao_dich_noi_bo,
    lsgd.ma_tham_chieu_cong_thanh_toan,
    lsgd.cong_thanh_toan_id,
    lsgd.so_tien,
    lsgd.loai_giao_dich,
    lsgd.trang_thai,
    lsgd.noi_dung_chuyen_khoan,
    lsgd.ngay_tao,
    lsgd.ngay_hoan_thanh,
    dc.phuong_thuc_thanh_toan,
    nd.ho_ten AS ten_nguoi_dung,
    nd.email AS email_nguoi_dung,
    t.tieu_de AS ten_tour,
    t.id AS tour_id,
    ctt.ten_hien_thi AS ten_cong_thanh_toan
FROM lich_su_giao_dich lsgd
LEFT JOIN dat_cho dc ON dc.id = lsgd.dat_cho_id
LEFT JOIN nguoi_dung nd ON nd.id = lsgd.nguoi_dung_id
LEFT JOIN khoi_hanh_tour kh ON kh.id = dc.khoi_hanh_id
LEFT JOIN tour t ON t.id = kh.tour_id
LEFT JOIN cong_thanh_toan ctt ON ctt.id = lsgd.cong_thanh_toan_id
WHERE lsgd.cong_thanh_toan_id = $1
ORDER BY lsgd.ngay_tao DESC
LIMIT $2 OFFSET $3;

