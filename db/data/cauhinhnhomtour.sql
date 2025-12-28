-- Insert data cho cấu hình nhóm tour
-- Lưu ý: Cần chạy tour.sql trước khi chạy file này

INSERT INTO cau_hinh_nhom_tour (tour_id, so_nho_nhat, so_lon_nhat)
VALUES
-- Các tour nội địa thông thường
(1, 10, 30),  -- Hạ Long - Sapa
(2, 10, 35),  -- Đà Nẵng - Hội An - Huế
(3, 2, 40),   -- Phú Quốc nghỉ dưỡng
(4, 5, 30),   -- Nha Trang
(5, 5, 25),   -- Đà Lạt
(6, 10, 30),  -- Mỹ Tho - Cần Thơ
(7, 8, 25),   -- Ninh Bình - Hạ Long
(8, 5, 20),   -- Côn Đảo (giới hạn vì bay)
(9, 5, 25),   -- Quy Nhơn - Phú Yên

-- Các tour quốc tế
(10, 15, 35), -- Bangkok - Pattaya
(11, 15, 30), -- Singapore - Malaysia
(12, 10, 25), -- Bali
(13, 15, 30), -- Hàn Quốc
(14, 10, 25), -- Nhật Bản (visa và chi phí cao)

-- Tour đặc biệt
(15, 30, 100), -- Team Building (nhóm lớn)
(16, 2, 2),    -- Trăng mật (cặp đôi)
(17, 2, 50),   -- Gia đình Phú Quốc
(18, 20, 50),  -- Du lịch học tập (nhóm học sinh)
(19, 6, 20),   -- Phong Nha - Kẻ Bàng (mạo hiểm)

-- Tour nháp và lưu trữ
(20, 15, 30), -- Tour Châu Âu (đang soạn)
(21, 10, 30); -- Tour lưu trữ

