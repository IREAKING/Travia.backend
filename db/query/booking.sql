-- ===========================================
-- BƯỚC 1: CHỌN TOUR & NGÀY KHỞI HÀNH
-- ===========================================

-- name: GetAvailableDepartures :many
-- Lấy danh sách ngày khởi hành còn chỗ của một tour
SELECT 
    kh.id,
    kh.tour_id,
    kh.ngay_khoi_hanh,
    kh.ngay_ket_thuc,
    kh.suc_chua,
    kh.so_cho_da_dat,
    (kh.suc_chua - kh.so_cho_da_dat) AS so_cho_trong,
    kh.trang_thai,
    kh.ghi_chu
FROM khoi_hanh_tour kh
WHERE kh.tour_id = $1
    AND kh.trang_thai IN ('len_lich', 'xac_nhan', 'con_cho')
    AND kh.ngay_khoi_hanh > CURRENT_DATE
    AND (kh.suc_chua - kh.so_cho_da_dat) > 0
ORDER BY kh.ngay_khoi_hanh ASC;

-- name: GetDepartureById :one
-- Lấy thông tin chi tiết một ngày khởi hành
SELECT 
    kh.*,
    t.tieu_de AS ten_tour,
    t.gia_nguoi_lon,
    t.gia_tre_em,
    t.don_vi_tien_te
FROM khoi_hanh_tour kh
JOIN tour t ON t.id = kh.tour_id
WHERE kh.id = $1;

-- name: CheckDepartureAvailability :one
-- Kiểm tra còn đủ chỗ không
SELECT 
    (suc_chua - so_cho_da_dat) >= $2 AS con_cho,
    (suc_chua - so_cho_da_dat) AS so_cho_trong
FROM khoi_hanh_tour
WHERE id = $1
    AND trang_thai IN ('len_lich', 'xac_nhan', 'con_cho');

-- ===========================================
-- BƯỚC 2: TẠO ĐẶT CHỖ (BOOKING)
-- ===========================================

-- Function giữ chỗ
CREATE OR REPLACE FUNCTION hold_seat(
    p_khoi_hanh_id INT,
    p_so_nguoi_lon INT,
    p_so_tre_em INT
) RETURNS INT AS $$
DECLARE
    v_suc_chua INT;
    v_so_cho_da_dat INT;
    v_tong_so_ghe_dat INT;
    v_so_cho_trong INT;
BEGIN
    v_tong_so_ghe_dat := p_so_nguoi_lon + p_so_tre_em;

    -- Kiểm tra số lượng người hợp lệ
    IF p_so_nguoi_lon < 0 OR p_so_tre_em < 0 OR v_tong_so_ghe_dat <= 0 THEN
        RAISE EXCEPTION 'Số lượng người lớn và trẻ em phải không âm và tổng số người phải lớn hơn 0.';
    END IF;

    SELECT suc_chua, COALESCE(so_cho_da_dat, 0)
    INTO STRICT v_suc_chua, v_so_cho_da_dat
    FROM khoi_hanh_tour
    WHERE id = p_khoi_hanh_id
    FOR UPDATE;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Tour khởi hành ID % không tồn tại.', p_khoi_hanh_id;
    END IF;

    -- Tính số chỗ trống
    v_so_cho_trong := v_suc_chua - v_so_cho_da_dat;

    -- Kiểm tra đủ chỗ
    IF v_tong_so_ghe_dat <= v_so_cho_trong THEN
        UPDATE khoi_hanh_tour
        SET so_cho_da_dat = so_cho_da_dat + v_tong_so_ghe_dat,
            ngay_cap_nhat = CURRENT_TIMESTAMP
        WHERE id = p_khoi_hanh_id;
        RETURN 1; 
    ELSE
        RAISE EXCEPTION 'Không đủ chỗ. Cần % chỗ, chỉ còn % chỗ trống.', 
            v_tong_so_ghe_dat, v_so_cho_trong;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- name: HoldSeat :exec
SELECT hold_seat(sqlc.arg('khoi_hanh_id')::int, sqlc.arg('so_nguoi_lon')::int, sqlc.arg('so_tre_em')::int);

-- Function tính giá tour (có áp dụng giảm giá nếu có)
CREATE OR REPLACE FUNCTION tinh_gia_tour(
    p_khoi_hanh_id INT,
    p_so_nguoi_lon INT,
    p_so_tre_em INT
) RETURNS TABLE (
    tong_tien DECIMAL(12,2),
    gia_goc DECIMAL(12,2),
    phan_tram_giam DECIMAL(5,2),
    tien_giam DECIMAL(12,2),
    don_vi_tien_te VARCHAR(3)
) AS $$
DECLARE
    v_tour_id INT;
    v_gia_nguoi_lon DECIMAL(10,2);
    v_gia_tre_em DECIMAL(10,2);
    v_don_vi_tien_te VARCHAR(3);
    v_phan_tram_giam DECIMAL(5,2) := 0;
    v_gia_goc DECIMAL(12,2);
    v_tien_giam DECIMAL(12,2);
    v_tong_tien DECIMAL(12,2);
BEGIN
    -- 1. Lấy thông tin tour từ khởi hành
    SELECT t.id, t.gia_nguoi_lon, t.gia_tre_em, t.don_vi_tien_te
    INTO STRICT v_tour_id, v_gia_nguoi_lon, v_gia_tre_em, v_don_vi_tien_te
    FROM khoi_hanh_tour kh
    JOIN tour t ON t.id = kh.tour_id
    WHERE kh.id = p_khoi_hanh_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Khởi hành ID % không tồn tại.', p_khoi_hanh_id;
    END IF;

    -- Kiểm tra giá không được NULL
    IF v_gia_nguoi_lon IS NULL OR v_gia_tre_em IS NULL THEN
        RAISE EXCEPTION 'Giá tour không hợp lệ cho khởi hành ID %', p_khoi_hanh_id;
    END IF;

    -- 2. Tính giá gốc
    v_gia_goc := (v_gia_nguoi_lon * p_so_nguoi_lon) + (v_gia_tre_em * p_so_tre_em);

    -- 3. Kiểm tra giảm giá hiện tại (nếu có)
    SELECT COALESCE(gg.phan_tram, 0)
    INTO v_phan_tram_giam
    FROM giam_gia_tour gg
    WHERE gg.tour_id = v_tour_id
        AND CURRENT_DATE BETWEEN gg.ngay_bat_dau AND gg.ngay_ket_thuc
    ORDER BY gg.phan_tram DESC
    LIMIT 1;

    -- 4. Tính tiền giảm và tổng tiền
    v_tien_giam := v_gia_goc * (v_phan_tram_giam / 100);
    v_tong_tien := v_gia_goc - v_tien_giam;

    -- Đảm bảo tổng tiền không NULL
    IF v_tong_tien IS NULL THEN
        v_tong_tien := v_gia_goc;
    END IF;

    -- 5. Trả về kết quả
    RETURN QUERY SELECT 
        v_tong_tien, 
        v_gia_goc, 
        v_phan_tram_giam, 
        v_tien_giam, 
        COALESCE(v_don_vi_tien_te, 'VND');
END;
$$ LANGUAGE plpgsql;

-- name: CalculateTourPrice :one
-- Tính giá tour cho khách hàng xem trước khi đặt
SELECT * FROM tinh_gia_tour(
    sqlc.arg('khoi_hanh_id')::int, 
    sqlc.arg('so_nguoi_lon')::int, 
    sqlc.arg('so_tre_em')::int
);

-- Function tạo booking với tự động tính giá
CREATE OR REPLACE FUNCTION create_booking(
    p_nguoi_dung_id UUID,
    p_khoi_hanh_id INT,
    p_so_nguoi_lon INT,
    p_so_tre_em INT,
    p_phuong_thuc_thanh_toan VARCHAR(50) DEFAULT NULL
) RETURNS dat_cho AS $$
DECLARE
    v_tong_tien DECIMAL(12,2);
    v_don_vi_tien_te VARCHAR(3);
    v_booking dat_cho;
BEGIN
    -- 1. Tính giá tour (lấy row đầu tiên từ TABLE result)
    SELECT tgt.tong_tien, tgt.don_vi_tien_te
    INTO STRICT v_tong_tien, v_don_vi_tien_te
    FROM tinh_gia_tour(p_khoi_hanh_id, p_so_nguoi_lon, p_so_tre_em) tgt
    LIMIT 1;

    -- Kiểm tra nếu không lấy được giá
    IF v_tong_tien IS NULL OR v_tong_tien <= 0 THEN
        RAISE EXCEPTION 'Không thể tính giá tour cho khởi hành ID %. Vui lòng kiểm tra thông tin tour và giá.', p_khoi_hanh_id;
    END IF;

    -- Đảm bảo đơn vị tiền tệ không NULL
    IF v_don_vi_tien_te IS NULL THEN
        v_don_vi_tien_te := 'VND';
    END IF;

    -- 2. Tạo booking
    INSERT INTO dat_cho (
        nguoi_dung_id,
        khoi_hanh_id,
        so_nguoi_lon,
        so_tre_em,
        tong_tien,
        don_vi_tien_te,
        trang_thai,
        phuong_thuc_thanh_toan
    ) VALUES (
        p_nguoi_dung_id,
        p_khoi_hanh_id,
        p_so_nguoi_lon,
        p_so_tre_em,
        v_tong_tien,
        v_don_vi_tien_te,
        'cho_xac_nhan',
        p_phuong_thuc_thanh_toan
    ) RETURNING * INTO v_booking;

    RETURN v_booking;
END;
$$ LANGUAGE plpgsql;

-- name: CreateBooking :one
-- Tạo đặt chỗ mới (tự động tính tổng tiền)
SELECT * FROM create_booking(
    sqlc.arg('nguoi_dung_id')::uuid,
    sqlc.arg('khoi_hanh_id')::int,
    sqlc.arg('so_nguoi_lon')::int,
    sqlc.arg('so_tre_em')::int,
    sqlc.narg('phuong_thuc_thanh_toan')::varchar
) AS booking;


-- name: GetBookingById :one
-- Lấy thông tin đặt chỗ theo ID
SELECT 
    dc.*,
    nd.ho_ten AS ten_nguoi_dat,
    nd.email,
    nd.so_dien_thoai,
    kh.ngay_khoi_hanh,
    kh.ngay_ket_thuc,
    t.tieu_de AS ten_tour,
    t.nha_cung_cap_id
FROM dat_cho dc
JOIN nguoi_dung nd ON nd.id = dc.nguoi_dung_id
JOIN khoi_hanh_tour kh ON kh.id = dc.khoi_hanh_id
JOIN tour t ON t.id = kh.tour_id
WHERE dc.id = $1;

-- name: GetBookingsByUser :many
-- Lấy danh sách đặt chỗ của người dùng
SELECT 
    dc.*,
    kh.ngay_khoi_hanh,
    kh.ngay_ket_thuc,
    kh.trang_thai AS trang_thai_khoi_hanh, -- Trạng thái thực tế của chuyến đi
    t.tieu_de AS ten_tour,
    -- Giả sử bảng anh_tour tồn tại như trong subquery của bạn
    (SELECT duong_dan FROM anh_tour WHERE tour_id = t.id AND la_anh_chinh = TRUE LIMIT 1) AS anh_tour
FROM dat_cho dc
JOIN khoi_hanh_tour kh ON kh.id = dc.khoi_hanh_id
JOIN tour t ON t.id = kh.tour_id
WHERE 
    dc.nguoi_dung_id = $1
    -- Lọc theo trạng thái đặt chỗ (truyền NULL hoặc empty string nếu muốn lấy tất cả)
    AND CASE 
        WHEN COALESCE($4::text, '') = '' THEN TRUE
        ELSE dc.trang_thai = ($4::text)::trang_thai_dat_cho
    END
    -- Lọc theo trạng thái khởi hành (truyền NULL hoặc empty string nếu muốn lấy tất cả)
    -- Hỗ trợ nhiều giá trị phân cách bằng dấu phẩy
    AND CASE 
        WHEN COALESCE($5::text, '') = '' THEN TRUE
        WHEN $5::text LIKE '%,%' THEN 
            kh.trang_thai IN (
                SELECT unnest(string_to_array($5::text, ','))::trang_thai_khoi_hanh
            )
        ELSE kh.trang_thai = ($5::text)::trang_thai_khoi_hanh
    END
ORDER BY dc.ngay_dat DESC
LIMIT $2 OFFSET $3;

-- name: DeleteBooking :exec
-- Xóa đặt chỗ
DELETE FROM dat_cho WHERE id = $1;

-- name: DeleteBookings :exec
-- Xóa nhiều đặt chỗ
DELETE FROM dat_cho WHERE id IN (SELECT unnest(sqlc.arg('ids')::int[]));

-- name: CountBookingsByUser :one
-- Đếm tổng số đặt chỗ của người dùng (có filter)
SELECT COUNT(*)::int AS total_count
FROM dat_cho dc
JOIN khoi_hanh_tour kh ON kh.id = dc.khoi_hanh_id
WHERE 
    dc.nguoi_dung_id = $1
    -- Lọc theo trạng thái đặt chỗ (truyền NULL hoặc empty string nếu muốn lấy tất cả)
    AND CASE 
        WHEN COALESCE($2::text, '') = '' THEN TRUE
        ELSE dc.trang_thai = ($2::text)::trang_thai_dat_cho
    END
    -- Lọc theo trạng thái khởi hành (truyền NULL hoặc empty string nếu muốn lấy tất cả)
    -- Hỗ trợ nhiều giá trị phân cách bằng dấu phẩy
    AND CASE 
        WHEN COALESCE($3::text, '') = '' THEN TRUE
        WHEN $3::text LIKE '%,%' THEN 
            kh.trang_thai IN (
                SELECT unnest(string_to_array($3::text, ','))::trang_thai_khoi_hanh
            )
        ELSE kh.trang_thai = ($3::text)::trang_thai_khoi_hanh
    END;

-- ===========================================
-- BƯỚC 3: NHẬP THÔNG TIN HÀNH KHÁCH
-- ===========================================

-- name: AddPassenger :one
-- Thêm hành khách vào booking
INSERT INTO hanh_khach (
    dat_cho_id,
    ho_ten,
    ngay_sinh,
    loai_khach,
    gioi_tinh,
    so_giay_to_tuy_thanh,
    quoc_tich,
    ghi_chu
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: AddPassengers :copyfrom
-- Thêm nhiều hành khách cùng lúc
INSERT INTO hanh_khach (
    dat_cho_id,
    ho_ten,
    ngay_sinh,
    loai_khach,
    gioi_tinh,
    so_giay_to_tuy_thanh,
    quoc_tich,
    ghi_chu
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
);

-- name: GetPassengersByBooking :many
-- Lấy danh sách hành khách của một booking
SELECT * FROM hanh_khach
WHERE dat_cho_id = $1
ORDER BY loai_khach DESC, id ASC;

-- name: UpdatePassenger :one
-- Cập nhật thông tin hành khách
UPDATE hanh_khach
SET 
    ho_ten = COALESCE(sqlc.narg('ho_ten'), ho_ten),
    ngay_sinh = COALESCE(sqlc.narg('ngay_sinh'), ngay_sinh),
    loai_khach = COALESCE(sqlc.narg('loai_khach'), loai_khach),
    gioi_tinh = COALESCE(sqlc.narg('gioi_tinh'), gioi_tinh),
    so_giay_to_tuy_thanh = COALESCE(sqlc.narg('so_giay_to_tuy_thanh'), so_giay_to_tuy_thanh),
    quoc_tich = COALESCE(sqlc.narg('quoc_tich'), quoc_tich),
    ghi_chu = COALESCE(sqlc.narg('ghi_chu'), ghi_chu)
WHERE id = $1
RETURNING *;

-- name: DeletePassenger :exec
-- Xóa hành khách
DELETE FROM hanh_khach WHERE id = $1;

-- ===========================================
-- BƯỚC 4: XÁC NHẬN ĐẶT CHỖ
-- ===========================================

-- name: ConfirmBooking :one
-- Admin/Hệ thống xác nhận đặt chỗ
UPDATE dat_cho
SET 
    trang_thai = 'da_xac_nhan',
    ngay_cap_nhat = CURRENT_TIMESTAMP
WHERE id = $1 AND trang_thai = 'cho_xac_nhan'
RETURNING *;

-- name: GetPendingBookings :many
-- Lấy danh sách booking chờ xác nhận (dành cho Admin/NCC)
SELECT 
    dc.*,
    nd.ho_ten AS ten_nguoi_dat,
    nd.email,
    nd.so_dien_thoai,
    kh.ngay_khoi_hanh,
    t.tieu_de AS ten_tour,
    t.nha_cung_cap_id
FROM dat_cho dc
JOIN nguoi_dung nd ON nd.id = dc.nguoi_dung_id
JOIN khoi_hanh_tour kh ON kh.id = dc.khoi_hanh_id
JOIN tour t ON t.id = kh.tour_id
WHERE dc.trang_thai = 'cho_xac_nhan'
    AND ($1::uuid IS NULL OR t.nha_cung_cap_id = $1)
ORDER BY dc.ngay_dat ASC
LIMIT $2 OFFSET $3;

-- ===========================================
-- BƯỚC 5: THANH TOÁN
-- ===========================================

-- name: CreateTransaction :one
-- Tạo giao dịch thanh toán mới
INSERT INTO lich_su_giao_dich (
    dat_cho_id,
    nguoi_dung_id,
    ma_giao_dich_noi_bo,
    cong_thanh_toan_id,
    so_tien,
    loai_giao_dich,
    trang_thai,
    noi_dung_chuyen_khoan
) VALUES (
    $1, $2, $3, $4, $5, 'thanh_toan', 'dang_cho_thanh_toan', $6
) RETURNING *;

-- name: UpdateTransactionStatus :one
-- Cập nhật trạng thái giao dịch (callback từ cổng thanh toán)
UPDATE lich_su_giao_dich
SET 
    trang_thai = $2,
    ma_tham_chieu_cong_thanh_toan = sqlc.narg('ma_tham_chieu'),
    ngay_hoan_thanh = CASE WHEN $2 = 'thanh_cong' THEN CURRENT_TIMESTAMP ELSE NULL END
WHERE id = $1
RETURNING *;

-- name: UpdateBookingPaymentStatus :one
-- Cập nhật trạng thái booking sau khi thanh toán thành công
UPDATE dat_cho
SET 
    trang_thai = 'da_thanh_toan',
    phuong_thuc_thanh_toan = $2,
    ngay_cap_nhat = CURRENT_TIMESTAMP
WHERE id = $1 AND trang_thai IN ('cho_xac_nhan', 'da_xac_nhan')
RETURNING *;

-- name: GetTransactionsByBooking :many
-- Lấy lịch sử giao dịch của một booking
SELECT 
    lsgd.*,
    ctt.ten_hien_thi AS ten_cong_thanh_toan
FROM lich_su_giao_dich lsgd
LEFT JOIN cong_thanh_toan ctt ON ctt.id = lsgd.cong_thanh_toan_id
WHERE lsgd.dat_cho_id = $1
ORDER BY lsgd.ngay_tao DESC;

-- name: GetTransactionByCode :one
-- Tìm giao dịch theo mã nội bộ
SELECT * FROM lich_su_giao_dich
WHERE ma_giao_dich_noi_bo = $1;

-- name: GetPaymentGateways :many
-- Lấy danh sách cổng thanh toán đang hoạt động
SELECT * FROM cong_thanh_toan
WHERE hoat_dong = TRUE;

-- ===========================================
-- BƯỚC 6: HOÀN THÀNH TOUR
-- ===========================================

-- name: CompleteBooking :one
-- Đánh dấu booking hoàn thành (sau khi tour kết thúc)
UPDATE dat_cho
SET 
    trang_thai = 'hoan_thanh',
    ngay_cap_nhat = CURRENT_TIMESTAMP
WHERE id = $1 AND trang_thai = 'da_thanh_toan'
RETURNING *;

-- name: AutoCompleteBookings :exec
-- Tự động hoàn thành các booking sau khi tour kết thúc (chạy bằng cron job)
UPDATE dat_cho dc
SET 
    trang_thai = 'hoan_thanh',
    ngay_cap_nhat = CURRENT_TIMESTAMP
FROM khoi_hanh_tour kh
WHERE dc.khoi_hanh_id = kh.id
    AND dc.trang_thai = 'da_thanh_toan'
    AND kh.ngay_ket_thuc < CURRENT_DATE;

-- ===========================================
-- HỦY BOOKING & HOÀN TIỀN
-- ===========================================

-- Function tính số tiền hoàn lại dựa trên chính sách hoàn tiền
CREATE OR REPLACE FUNCTION tinh_tien_hoan_lai(
    p_booking_id INT
) RETURNS TABLE (
    tong_tien DECIMAL(12,2),
    so_tien_hoan DECIMAL(12,2),
    phan_tram_hoan DECIMAL(5,2),
    so_ngay_truoc_khoi_hanh INT,
    ly_do TEXT
) AS $$
DECLARE
    v_tong_tien DECIMAL(12,2);
    v_ngay_khoi_hanh DATE;
    v_so_ngay_truoc_khoi_hanh INT;
    v_phan_tram_hoan DECIMAL(5,2);
    v_so_tien_hoan DECIMAL(12,2);
    v_ly_do TEXT;
BEGIN
    -- Lấy thông tin booking và ngày khởi hành
    SELECT dc.tong_tien, kh.ngay_khoi_hanh
    INTO v_tong_tien, v_ngay_khoi_hanh
    FROM dat_cho dc
    JOIN khoi_hanh_tour kh ON kh.id = dc.khoi_hanh_id
    WHERE dc.id = p_booking_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Booking ID % không tồn tại.', p_booking_id;
    END IF;

    -- Tính số ngày trước ngày khởi hành
    v_so_ngay_truoc_khoi_hanh := v_ngay_khoi_hanh - CURRENT_DATE;

    -- Áp dụng chính sách hoàn tiền
    IF v_so_ngay_truoc_khoi_hanh >= 15 THEN
        -- Trước 15 ngày: hoàn 100%
        v_phan_tram_hoan := 100.00;
        v_ly_do := 'Hủy trước 15 ngày - hoàn 100%';
    ELSIF v_so_ngay_truoc_khoi_hanh >= 7 THEN
        -- Trước 7 ngày: hoàn 90%
        v_phan_tram_hoan := 90.00;
        v_ly_do := 'Hủy trước 7 ngày - hoàn 90%';
    ELSIF v_so_ngay_truoc_khoi_hanh >= 3 THEN
        -- Trước 3 ngày: hoàn 70%
        v_phan_tram_hoan := 70.00;
        v_ly_do := 'Hủy trước 3 ngày - hoàn 70%';
    ELSIF v_so_ngay_truoc_khoi_hanh >= 1 THEN
        -- Trước 24 giờ (1 ngày): hoàn 50%
        v_phan_tram_hoan := 50.00;
        v_ly_do := 'Hủy trước 24 giờ - hoàn 50%';
    ELSE
        -- Trong 24 giờ: không hoàn
        v_phan_tram_hoan := 0.00;
        v_ly_do := 'Hủy trong 24 giờ - không hoàn tiền';
    END IF;

    -- Tính số tiền hoàn lại
    v_so_tien_hoan := (v_tong_tien * v_phan_tram_hoan) / 100.00;

    RETURN QUERY SELECT v_tong_tien, v_so_tien_hoan, v_phan_tram_hoan, v_so_ngay_truoc_khoi_hanh, v_ly_do;
END;
$$ LANGUAGE plpgsql;

-- Function hủy booking và trả lại chỗ
CREATE OR REPLACE FUNCTION cancel_booking(
    p_booking_id INT
) RETURNS TABLE (
    so_tien_hoan DECIMAL(12,2),
    phan_tram_hoan DECIMAL(5,2),
    so_ngay_truoc_khoi_hanh INT,
    ly_do TEXT
) AS $$
DECLARE
    v_khoi_hanh_id INT;
    v_so_nguoi_lon INT;
    v_so_tre_em INT;
    v_trang_thai trang_thai_dat_cho;
    v_refund_info RECORD;
BEGIN
    -- Lấy thông tin booking
    SELECT khoi_hanh_id, so_nguoi_lon, so_tre_em, trang_thai
    INTO v_khoi_hanh_id, v_so_nguoi_lon, v_so_tre_em, v_trang_thai
    FROM dat_cho
    WHERE id = p_booking_id
    FOR UPDATE;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Booking ID % không tồn tại.', p_booking_id;
    END IF;

    IF v_trang_thai = 'da_huy' THEN
        RAISE EXCEPTION 'Booking đã bị hủy trước đó.';
    END IF;

    IF v_trang_thai = 'hoan_thanh' THEN
        RAISE EXCEPTION 'Không thể hủy booking đã hoàn thành.';
    END IF;

    -- Tính số tiền hoàn lại
    SELECT * INTO v_refund_info
    FROM tinh_tien_hoan_lai(p_booking_id);

    -- Cập nhật trạng thái booking
    UPDATE dat_cho
    SET trang_thai = 'da_huy', ngay_cap_nhat = CURRENT_TIMESTAMP
    WHERE id = p_booking_id;

    -- Tính lại số chỗ đã đặt từ tổng số booking thực tế (không bao gồm booking đã hủy)
    -- Điều này đảm bảo dữ liệu luôn đồng bộ và không bị lệch
    UPDATE khoi_hanh_tour
    SET so_cho_da_dat = COALESCE((
        SELECT SUM(COALESCE(so_nguoi_lon, 0) + COALESCE(so_tre_em, 0))
        FROM dat_cho
        WHERE khoi_hanh_id = v_khoi_hanh_id
            AND trang_thai != 'da_huy'
    ), 0),
        ngay_cap_nhat = CURRENT_TIMESTAMP
    WHERE id = v_khoi_hanh_id;

    -- Trả về thông tin hoàn tiền
    RETURN QUERY SELECT 
        v_refund_info.so_tien_hoan,
        v_refund_info.phan_tram_hoan,
        v_refund_info.so_ngay_truoc_khoi_hanh,
        v_refund_info.ly_do;
END;
$$ LANGUAGE plpgsql;

-- name: CalculateRefundAmount :one
WITH booking_info AS (
    SELECT 
        dc.tong_tien,
        kh.ngay_khoi_hanh,
        (kh.ngay_khoi_hanh - CURRENT_DATE)::INT AS so_ngay_truoc_khoi_hanh
    FROM dat_cho dc
    JOIN khoi_hanh_tour kh ON kh.id = dc.khoi_hanh_id
    WHERE dc.id = sqlc.arg('booking_id')::int
),
refund_calc AS (
    SELECT 
        bi.tong_tien,
        bi.so_ngay_truoc_khoi_hanh,
        CASE 
            WHEN bi.so_ngay_truoc_khoi_hanh >= 15 THEN 100.00
            WHEN bi.so_ngay_truoc_khoi_hanh >= 7 THEN 90.00
            WHEN bi.so_ngay_truoc_khoi_hanh >= 3 THEN 70.00
            WHEN bi.so_ngay_truoc_khoi_hanh >= 1 THEN 50.00
            ELSE 0.00
        END::DECIMAL(5,2) AS phan_tram_hoan,
        CASE 
            WHEN bi.so_ngay_truoc_khoi_hanh >= 15 THEN 'Hủy trước 15 ngày - hoàn 100%'
            WHEN bi.so_ngay_truoc_khoi_hanh >= 7 THEN 'Hủy trước 7 ngày - hoàn 90%'
            WHEN bi.so_ngay_truoc_khoi_hanh >= 3 THEN 'Hủy trước 3 ngày - hoàn 70%'
            WHEN bi.so_ngay_truoc_khoi_hanh >= 1 THEN 'Hủy trước 24 giờ - hoàn 50%'
            ELSE 'Hủy trong 24 giờ - không hoàn tiền'
        END::TEXT AS ly_do
    FROM booking_info bi
)
SELECT 
    rc.tong_tien,
    (rc.tong_tien * rc.phan_tram_hoan / 100.00)::DECIMAL(12,2) AS so_tien_hoan,
    rc.phan_tram_hoan,
    rc.so_ngay_truoc_khoi_hanh,
    rc.ly_do
FROM refund_calc rc;

-- name: CancelBooking :one
SELECT 
    so_tien_hoan::DECIMAL(12,2) as so_tien_hoan,
    phan_tram_hoan::DECIMAL(5,2) as phan_tram_hoan,
    so_ngay_truoc_khoi_hanh::INT as so_ngay_truoc_khoi_hanh,
    ly_do::TEXT as ly_do
FROM cancel_booking(sqlc.arg('booking_id')::int)
LIMIT 1;

-- name: GetCancelledBookings :many
-- Lấy danh sách booking đã hủy
SELECT 
    dc.*,
    nd.ho_ten AS ten_nguoi_dat,
    kh.ngay_khoi_hanh,
    t.tieu_de AS ten_tour
FROM dat_cho dc
JOIN nguoi_dung nd ON nd.id = dc.nguoi_dung_id
JOIN khoi_hanh_tour kh ON kh.id = dc.khoi_hanh_id
JOIN tour t ON t.id = kh.tour_id
WHERE dc.trang_thai = 'da_huy'
    AND ($1::uuid IS NULL OR t.nha_cung_cap_id = $1)
ORDER BY dc.ngay_cap_nhat DESC
LIMIT $2 OFFSET $3;

-- ===========================================
-- THỐNG KÊ & BÁO CÁO
-- ===========================================

-- name: GetBookingStats :one
-- Thống kê booking theo nhà cung cấp
SELECT 
    COUNT(*) FILTER (WHERE dc.trang_thai = 'cho_xac_nhan') AS cho_xac_nhan,
    COUNT(*) FILTER (WHERE dc.trang_thai = 'da_xac_nhan') AS da_xac_nhan,
    COUNT(*) FILTER (WHERE dc.trang_thai = 'da_thanh_toan') AS da_thanh_toan,
    COUNT(*) FILTER (WHERE dc.trang_thai = 'hoan_thanh') AS hoan_thanh,
    COUNT(*) FILTER (WHERE dc.trang_thai = 'da_huy') AS da_huy,
    COALESCE(SUM(dc.tong_tien) FILTER (WHERE dc.trang_thai IN ('da_thanh_toan', 'hoan_thanh')), 0) AS tong_doanh_thu
FROM dat_cho dc
JOIN khoi_hanh_tour kh ON kh.id = dc.khoi_hanh_id
JOIN tour t ON t.id = kh.tour_id
WHERE ($1::uuid IS NULL OR t.nha_cung_cap_id = $1)
    AND ($2::timestamp IS NULL OR dc.ngay_dat >= $2)
    AND ($3::timestamp IS NULL OR dc.ngay_dat <= $3);

-- name: GetBookingsByStatus :many
-- Lấy booking theo trạng thái
SELECT 
    dc.*,
    nd.ho_ten AS ten_nguoi_dat,
    nd.email,
    kh.ngay_khoi_hanh,
    kh.ngay_ket_thuc,
    t.tieu_de AS ten_tour
FROM dat_cho dc
JOIN nguoi_dung nd ON nd.id = dc.nguoi_dung_id
JOIN khoi_hanh_tour kh ON kh.id = dc.khoi_hanh_id
JOIN tour t ON t.id = kh.tour_id
WHERE dc.trang_thai = $1::trang_thai_dat_cho
    AND ($2::uuid IS NULL OR t.nha_cung_cap_id = $2)
ORDER BY dc.ngay_dat DESC
LIMIT $3 OFFSET $4;

-- name: GetBookingsByUserId :many
SELECT 
    dc.*,
    nd.ho_ten AS ten_nguoi_dat,
    kh.ngay_khoi_hanh,
    kh.ngay_ket_thuc,
    t.tieu_de AS ten_tour
FROM dat_cho dc
JOIN nguoi_dung nd ON nd.id = dc.nguoi_dung_id
JOIN khoi_hanh_tour kh ON kh.id = dc.khoi_hanh_id
JOIN tour t ON t.id = kh.tour_id
WHERE dc.nguoi_dung_id = $1;

-- ===========================================
-- QUẢN LÝ HOÀN TIỀN (REFUND MANAGEMENT)
-- ===========================================

-- name: GetAllRefunds :many
-- Lấy tất cả refund cho admin (tất cả booking đã hủy với thông tin refund)
WITH refund_info AS (
    SELECT 
        dc.id AS booking_id,
        dc.tong_tien,
        kh.ngay_khoi_hanh,
        -- Tính số ngày trước khởi hành tại thời điểm hủy (ngay_cap_nhat)
        (kh.ngay_khoi_hanh - DATE(dc.ngay_cap_nhat))::INT AS so_ngay_truoc_khoi_hanh
    FROM dat_cho dc
    JOIN khoi_hanh_tour kh ON kh.id = dc.khoi_hanh_id
    WHERE dc.trang_thai = 'da_huy'
)
SELECT 
    dc.id AS booking_id,
    dc.ngay_dat,
    dc.ngay_cap_nhat AS ngay_huy,
    dc.tong_tien,
    dc.don_vi_tien_te,
    dc.so_nguoi_lon,
    dc.so_tre_em,
    dc.phuong_thuc_thanh_toan,
    dc.trang_thai,
    
    -- Thông tin khách hàng
    nd.id AS customer_id,
    nd.ho_ten AS customer_name,
    nd.email AS customer_email,
    nd.so_dien_thoai AS customer_phone,
    
    -- Thông tin tour
    t.id AS tour_id,
    t.tieu_de AS tour_title,
    ncc.id AS supplier_id,
    ncc.ten AS supplier_name,
    
    -- Thông tin khởi hành
    kh.id AS departure_id,
    kh.ngay_khoi_hanh,
    kh.ngay_ket_thuc,
    
    -- Thông tin refund
    ri.so_ngay_truoc_khoi_hanh,
    CASE 
        WHEN ri.so_ngay_truoc_khoi_hanh >= 15 THEN 100.00
        WHEN ri.so_ngay_truoc_khoi_hanh >= 7 THEN 90.00
        WHEN ri.so_ngay_truoc_khoi_hanh >= 3 THEN 70.00
        WHEN ri.so_ngay_truoc_khoi_hanh >= 1 THEN 50.00
        ELSE 0.00
    END::DECIMAL(5,2) AS phan_tram_hoan,
    (ri.tong_tien * 
        CASE 
            WHEN ri.so_ngay_truoc_khoi_hanh >= 15 THEN 100.00
            WHEN ri.so_ngay_truoc_khoi_hanh >= 7 THEN 90.00
            WHEN ri.so_ngay_truoc_khoi_hanh >= 3 THEN 70.00
            WHEN ri.so_ngay_truoc_khoi_hanh >= 1 THEN 50.00
            ELSE 0.00
        END / 100.00)::DECIMAL(12,2) AS so_tien_hoan,
    CASE 
        WHEN ri.so_ngay_truoc_khoi_hanh >= 15 THEN 'Hủy trước 15 ngày - hoàn 100%'
        WHEN ri.so_ngay_truoc_khoi_hanh >= 7 THEN 'Hủy trước 7 ngày - hoàn 90%'
        WHEN ri.so_ngay_truoc_khoi_hanh >= 3 THEN 'Hủy trước 3 ngày - hoàn 70%'
        WHEN ri.so_ngay_truoc_khoi_hanh >= 1 THEN 'Hủy trước 24 giờ - hoàn 50%'
        ELSE 'Hủy trong 24 giờ - không hoàn tiền'
    END::TEXT AS ly_do
FROM dat_cho dc
JOIN nguoi_dung nd ON nd.id = dc.nguoi_dung_id
JOIN khoi_hanh_tour kh ON kh.id = dc.khoi_hanh_id
JOIN tour t ON t.id = kh.tour_id
JOIN nha_cung_cap ncc ON ncc.id = t.nha_cung_cap_id
JOIN refund_info ri ON ri.booking_id = dc.id
WHERE dc.trang_thai = 'da_huy'
    AND ($1::timestamp IS NULL OR dc.ngay_cap_nhat >= $1::timestamp)
    AND ($2::timestamp IS NULL OR dc.ngay_cap_nhat <= $2::timestamp)
    AND ($3::uuid IS NULL OR ncc.id = $3::uuid)
    AND ($4::text IS NULL OR nd.ho_ten ILIKE '%' || $4::text || '%' OR nd.email ILIKE '%' || $4::text || '%' OR t.tieu_de ILIKE '%' || $4::text || '%')
ORDER BY dc.ngay_cap_nhat DESC
LIMIT $5 OFFSET $6;

-- name: GetSupplierRefunds :many
-- Lấy refund cho supplier (chỉ tour của họ)
WITH refund_info AS (
    SELECT 
        dc.id AS booking_id,
        dc.tong_tien,
        kh.ngay_khoi_hanh,
        -- Tính số ngày trước khởi hành tại thời điểm hủy (ngay_cap_nhat)
        (kh.ngay_khoi_hanh - DATE(dc.ngay_cap_nhat))::INT AS so_ngay_truoc_khoi_hanh
    FROM dat_cho dc
    JOIN khoi_hanh_tour kh ON kh.id = dc.khoi_hanh_id
    JOIN tour t ON t.id = kh.tour_id
    WHERE dc.trang_thai = 'da_huy'
        AND t.nha_cung_cap_id = $1::uuid
)
SELECT 
    dc.id AS booking_id,
    dc.ngay_dat,
    dc.ngay_cap_nhat AS ngay_huy,
    dc.tong_tien,
    dc.don_vi_tien_te,
    dc.so_nguoi_lon,
    dc.so_tre_em,
    dc.phuong_thuc_thanh_toan,
    dc.trang_thai,
    
    -- Thông tin khách hàng
    nd.id AS customer_id,
    nd.ho_ten AS customer_name,
    nd.email AS customer_email,
    nd.so_dien_thoai AS customer_phone,
    
    -- Thông tin tour
    t.id AS tour_id,
    t.tieu_de AS tour_title,
    
    -- Thông tin khởi hành
    kh.id AS departure_id,
    kh.ngay_khoi_hanh,
    kh.ngay_ket_thuc,
    
    -- Thông tin refund
    ri.so_ngay_truoc_khoi_hanh,
    CASE 
        WHEN ri.so_ngay_truoc_khoi_hanh >= 15 THEN 100.00
        WHEN ri.so_ngay_truoc_khoi_hanh >= 7 THEN 90.00
        WHEN ri.so_ngay_truoc_khoi_hanh >= 3 THEN 70.00
        WHEN ri.so_ngay_truoc_khoi_hanh >= 1 THEN 50.00
        ELSE 0.00
    END::DECIMAL(5,2) AS phan_tram_hoan,
    (ri.tong_tien * 
        CASE 
            WHEN ri.so_ngay_truoc_khoi_hanh >= 15 THEN 100.00
            WHEN ri.so_ngay_truoc_khoi_hanh >= 7 THEN 90.00
            WHEN ri.so_ngay_truoc_khoi_hanh >= 3 THEN 70.00
            WHEN ri.so_ngay_truoc_khoi_hanh >= 1 THEN 50.00
            ELSE 0.00
        END / 100.00)::DECIMAL(12,2) AS so_tien_hoan,
    CASE 
        WHEN ri.so_ngay_truoc_khoi_hanh >= 15 THEN 'Hủy trước 15 ngày - hoàn 100%'
        WHEN ri.so_ngay_truoc_khoi_hanh >= 7 THEN 'Hủy trước 7 ngày - hoàn 90%'
        WHEN ri.so_ngay_truoc_khoi_hanh >= 3 THEN 'Hủy trước 3 ngày - hoàn 70%'
        WHEN ri.so_ngay_truoc_khoi_hanh >= 1 THEN 'Hủy trước 24 giờ - hoàn 50%'
        ELSE 'Hủy trong 24 giờ - không hoàn tiền'
    END::TEXT AS ly_do
FROM dat_cho dc
JOIN nguoi_dung nd ON nd.id = dc.nguoi_dung_id
JOIN khoi_hanh_tour kh ON kh.id = dc.khoi_hanh_id
JOIN tour t ON t.id = kh.tour_id
JOIN refund_info ri ON ri.booking_id = dc.id
WHERE dc.trang_thai = 'da_huy'
    AND t.nha_cung_cap_id = $1::uuid
    AND ($2::timestamp IS NULL OR dc.ngay_cap_nhat >= $2::timestamp)
    AND ($3::timestamp IS NULL OR dc.ngay_cap_nhat <= $3::timestamp)
    AND ($4::text IS NULL OR nd.ho_ten ILIKE '%' || $4::text || '%' OR nd.email ILIKE '%' || $4::text || '%' OR t.tieu_de ILIKE '%' || $4::text || '%')
ORDER BY dc.ngay_cap_nhat DESC
LIMIT $5 OFFSET $6;

-- name: GetRefundStats :one
-- Thống kê refund cho admin
WITH refund_info AS (
    SELECT 
        dc.id AS booking_id,
        dc.tong_tien,
        dc.ngay_cap_nhat AS ngay_huy,
        kh.ngay_khoi_hanh,
        (kh.ngay_khoi_hanh - CURRENT_DATE)::INT AS so_ngay_truoc_khoi_hanh,
        t.nha_cung_cap_id
    FROM dat_cho dc
    JOIN khoi_hanh_tour kh ON kh.id = dc.khoi_hanh_id
    JOIN tour t ON t.id = kh.tour_id
    WHERE dc.trang_thai = 'da_huy'
        AND ($1::timestamp IS NULL OR dc.ngay_cap_nhat >= $1::timestamp)
        AND ($2::timestamp IS NULL OR dc.ngay_cap_nhat <= $2::timestamp)
)
SELECT 
    COUNT(*)::int AS tong_so_refund,
    COALESCE(SUM(ri.tong_tien), 0)::numeric AS tong_tien_goc,
    COALESCE(SUM(
        ri.tong_tien * 
        CASE 
            WHEN ri.so_ngay_truoc_khoi_hanh >= 15 THEN 100.00
            WHEN ri.so_ngay_truoc_khoi_hanh >= 7 THEN 90.00
            WHEN ri.so_ngay_truoc_khoi_hanh >= 3 THEN 70.00
            WHEN ri.so_ngay_truoc_khoi_hanh >= 1 THEN 50.00
            ELSE 0.00
        END / 100.00
    ), 0)::numeric AS tong_tien_hoan,
    COALESCE(SUM(
        ri.tong_tien * 
        (100.00 - 
            CASE 
                WHEN ri.so_ngay_truoc_khoi_hanh >= 15 THEN 100.00
                WHEN ri.so_ngay_truoc_khoi_hanh >= 7 THEN 90.00
                WHEN ri.so_ngay_truoc_khoi_hanh >= 3 THEN 70.00
                WHEN ri.so_ngay_truoc_khoi_hanh >= 1 THEN 50.00
                ELSE 0.00
            END
        ) / 100.00
    ), 0)::numeric AS tong_tien_phat,
    COUNT(*) FILTER (WHERE ri.so_ngay_truoc_khoi_hanh >= 15)::int AS hoan_100_percent,
    COUNT(*) FILTER (WHERE ri.so_ngay_truoc_khoi_hanh >= 7 AND ri.so_ngay_truoc_khoi_hanh < 15)::int AS hoan_90_percent,
    COUNT(*) FILTER (WHERE ri.so_ngay_truoc_khoi_hanh >= 3 AND ri.so_ngay_truoc_khoi_hanh < 7)::int AS hoan_70_percent,
    COUNT(*) FILTER (WHERE ri.so_ngay_truoc_khoi_hanh >= 1 AND ri.so_ngay_truoc_khoi_hanh < 3)::int AS hoan_50_percent,
    COUNT(*) FILTER (WHERE ri.so_ngay_truoc_khoi_hanh < 1)::int AS khong_hoan
FROM refund_info ri;

-- name: GetSupplierRefundStats :one
-- Thống kê refund cho supplier
WITH refund_info AS (
    SELECT 
        dc.id AS booking_id,
        dc.tong_tien,
        dc.ngay_cap_nhat AS ngay_huy,
        kh.ngay_khoi_hanh,
        (kh.ngay_khoi_hanh - CURRENT_DATE)::INT AS so_ngay_truoc_khoi_hanh
    FROM dat_cho dc
    JOIN khoi_hanh_tour kh ON kh.id = dc.khoi_hanh_id
    JOIN tour t ON t.id = kh.tour_id
    WHERE dc.trang_thai = 'da_huy'
        AND t.nha_cung_cap_id = $1::uuid
        AND ($2::timestamp IS NULL OR dc.ngay_cap_nhat >= $2::timestamp)
        AND ($3::timestamp IS NULL OR dc.ngay_cap_nhat <= $3::timestamp)
)
SELECT 
    COUNT(*)::int AS tong_so_refund,
    COALESCE(SUM(ri.tong_tien), 0)::numeric AS tong_tien_goc,
    COALESCE(SUM(
        ri.tong_tien * 
        CASE 
            WHEN ri.so_ngay_truoc_khoi_hanh >= 15 THEN 100.00
            WHEN ri.so_ngay_truoc_khoi_hanh >= 7 THEN 90.00
            WHEN ri.so_ngay_truoc_khoi_hanh >= 3 THEN 70.00
            WHEN ri.so_ngay_truoc_khoi_hanh >= 1 THEN 50.00
            ELSE 0.00
        END / 100.00
    ), 0)::numeric AS tong_tien_hoan,
    COALESCE(SUM(
        ri.tong_tien * 
        (100.00 - 
            CASE 
                WHEN ri.so_ngay_truoc_khoi_hanh >= 15 THEN 100.00
                WHEN ri.so_ngay_truoc_khoi_hanh >= 7 THEN 90.00
                WHEN ri.so_ngay_truoc_khoi_hanh >= 3 THEN 70.00
                WHEN ri.so_ngay_truoc_khoi_hanh >= 1 THEN 50.00
                ELSE 0.00
            END
        ) / 100.00
    ), 0)::numeric AS tong_tien_phat,
    COUNT(*) FILTER (WHERE ri.so_ngay_truoc_khoi_hanh >= 15)::int AS hoan_100_percent,
    COUNT(*) FILTER (WHERE ri.so_ngay_truoc_khoi_hanh >= 7 AND ri.so_ngay_truoc_khoi_hanh < 15)::int AS hoan_90_percent,
    COUNT(*) FILTER (WHERE ri.so_ngay_truoc_khoi_hanh >= 3 AND ri.so_ngay_truoc_khoi_hanh < 7)::int AS hoan_70_percent,
    COUNT(*) FILTER (WHERE ri.so_ngay_truoc_khoi_hanh >= 1 AND ri.so_ngay_truoc_khoi_hanh < 3)::int AS hoan_50_percent,
    COUNT(*) FILTER (WHERE ri.so_ngay_truoc_khoi_hanh < 1)::int AS khong_hoan
FROM refund_info ri;