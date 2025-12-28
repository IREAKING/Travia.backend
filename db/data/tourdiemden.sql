-- Insert data cho liên kết tour với điểm đến
-- Lưu ý: Cần chạy tour.sql và diemden.sql trước khi chạy file này

INSERT INTO tour_diem_den (tour_id, diem_den_id, thu_tu_tham_quan)
VALUES
-- Tour 1: Hạ Long - Sapa
(1, 1, 1),  -- Hà Nội
(1, 2, 2),  -- Vịnh Hạ Long
(1, 3, 3),  -- Sapa

-- Tour 2: Đà Nẵng - Hội An - Huế
(2, 5, 1),  -- Đà Nẵng
(2, 6, 2),  -- Hội An
(2, 7, 3),  -- Huế

-- Tour 3: Phú Quốc
(3, 13, 1), -- Phú Quốc

-- Tour 4: Nha Trang
(4, 12, 1), -- Nha Trang

-- Tour 5: Đà Lạt
(5, 11, 1), -- Đà Lạt

-- Tour 6: Mỹ Tho - Cần Thơ
(6, 10, 1), -- TP. Hồ Chí Minh
(6, 14, 2), -- Cần Thơ

-- Tour 7: Ninh Bình - Hạ Long
(7, 4, 1),  -- Ninh Bình
(7, 2, 2),  -- Vịnh Hạ Long

-- Tour 8: Côn Đảo
(8, 10, 1), -- TP. Hồ Chí Minh (điểm khởi hành)

-- Tour 9: Quy Nhơn - Phú Yên
(9, 9, 1),  -- Quy Nhơn

-- Tour 10: Bangkok - Pattaya
(10, 16, 1), -- Bangkok
(10, 17, 2), -- Phuket

-- Tour 11: Singapore - Malaysia
(11, 18, 1), -- Singapore

-- Tour 12: Bali
(12, 19, 1), -- Bali

-- Tour 13: Seoul - Nami
(13, 22, 1), -- Seoul
(13, 23, 2), -- Jeju

-- Tour 14: Tokyo - Osaka
(14, 20, 1), -- Tokyo
(14, 21, 2), -- Kyoto

-- Tour 16: Trăng mật Đà Nẵng
(16, 5, 1),  -- Đà Nẵng
(16, 6, 2),  -- Hội An

-- Tour 17: Gia đình Phú Quốc
(17, 13, 1), -- Phú Quốc

-- Tour 18: Du lịch học tập Hà Nội
(18, 1, 1),  -- Hà Nội

-- Tour 19: Phong Nha - Kẻ Bàng
(19, 8, 1);  -- Phong Nha - Kẻ Bàng

