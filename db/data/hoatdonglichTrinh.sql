-- Insert data cho hoạt động trong lịch trình
-- Lưu ý: Cần chạy lichtrinhtour.sql trước khi chạy file này

-- Ngày 1 - Tour 1: Hà Nội
INSERT INTO hoat_dong_trong_ngay (lich_trinh_id, ten, gio_bat_dau, gio_ket_thuc, mo_ta, thu_tu)
VALUES
(1, 'Đón khách tại sân bay Nội Bài', '14:00:00', '15:30:00', 'HDV đón khách tại sảnh đến quốc tế, đưa về khách sạn', 1),
(1, 'Nhận phòng khách sạn', '15:30:00', '16:00:00', 'Check-in khách sạn, nghỉ ngơi', 2),
(1, 'Tham quan Văn Miếu Quốc Tử Giám', '16:30:00', '17:30:00', 'Tham quan ngôi trường đại học đầu tiên của Việt Nam', 3),
(1, 'Tham quan Hồ Hoàn Kiếm', '17:45:00', '18:30:00', 'Dạo quanh hồ, chụp ảnh, tham quan Đền Ngọc Sơn', 4),
(1, 'Ăn tối bún chả Hà Nội', '18:45:00', '19:45:00', 'Thưởng thức món ăn đặc sản Hà Nội', 5),
(1, 'Xem múa rối nước', '20:00:00', '21:00:00', 'Thưởng thức nghệ thuật truyền thống Việt Nam', 6);

-- Ngày 2 - Tour 1: Hạ Long
INSERT INTO hoat_dong_trong_ngay (lich_trinh_id, ten, gio_bat_dau, gio_ket_thuc, mo_ta, thu_tu)
VALUES
(2, 'Ăn sáng buffet', '07:00:00', '08:00:00', 'Bữa sáng tại khách sạn', 1),
(2, 'Di chuyển đến Hạ Long', '08:30:00', '12:00:00', 'Xe đưa đoàn đi Hạ Long, nghỉ dọc đường', 2),
(2, 'Lên tàu du lịch', '12:00:00', '12:30:00', 'Check-in cabin, giới thiệu an toàn', 3),
(2, 'Ăn trưa hải sản trên tàu', '12:30:00', '13:30:00', 'Thưởng thức hải sản tươi ngon', 4),
(2, 'Tham quan hang Sửng Sốt', '14:00:00', '15:00:00', 'Khám phá hang động kỳ vĩ nhất vịnh Hạ Long', 5),
(2, 'Chèo kayak và bơi lội', '15:30:00', '17:00:00', 'Tự do vui chơi trên biển', 6),
(2, 'Tham quan làng chài Cửa Vạn', '17:30:00', '18:30:00', 'Tìm hiểu đời sống ngư dân', 7),
(2, 'BBQ tối trên boong tàu', '19:00:00', '21:00:00', 'Tiệc BBQ hải sản trên boong tàu', 8),
(2, 'Câu mực đêm', '21:00:00', '22:00:00', 'Trải nghiệm câu mực đêm, tự do', 9);

-- Ngày 1 - Tour 2: Đà Nẵng - Bà Nà
INSERT INTO hoat_dong_trong_ngay (lich_trinh_id, ten, gio_bat_dau, gio_ket_thuc, mo_ta, thu_tu)
VALUES
(6, 'Đón khách tại sân bay Đà Nẵng', '10:00:00', '10:30:00', 'HDV đón khách tại sân bay', 1),
(6, 'Di chuyển lên Bà Nà Hills', '10:30:00', '11:30:00', 'Đi cáp treo lên Bà Nà (cáp treo dài nhất thế giới)', 2),
(6, 'Tham quan Cầu Vàng', '11:30:00', '12:15:00', 'Check-in Cầu Vàng nổi tiếng với đôi bàn tay khổng lồ', 3),
(6, 'Ăn trưa buffet', '12:30:00', '13:30:00', 'Buffet gần 100 món tại nhà hàng Bà Nà', 4),
(6, 'Tham quan Làng Pháp', '14:00:00', '15:00:00', 'Kiến trúc Pháp cổ kính, nhà thờ, quảng trường', 5),
(6, 'Vui chơi Fantasy Park', '15:00:00', '17:00:00', 'Công viên giải trí trong nhà lớn nhất Việt Nam', 6),
(6, 'Xuống núi', '17:00:00', '18:00:00', 'Di chuyển xuống núi bằng cáp treo', 7),
(6, 'Tắm biển Mỹ Khê', '18:00:00', '19:00:00', 'Tắm biển, chụp ảnh hoàng hôn', 8);

-- Ngày 2 - Tour 3: Safari Phú Quốc
INSERT INTO hoat_dong_trong_ngay (lich_trinh_id, ten, gio_bat_dau, gio_ket_thuc, mo_ta, thu_tu)
VALUES
(11, 'Ăn sáng buffet tại resort', '07:00:00', '08:00:00', 'Bữa sáng phong phú', 1),
(11, 'Di chuyển đến Safari Phú Quốc', '08:00:00', '08:30:00', 'Xe đưa đoàn đến Safari', 2),
(11, 'Tham quan khu Safari', '08:30:00', '11:00:00', 'Ngắm động vật hoang dã: hổ, sư tử, hươu cao cổ, voi...', 3),
(11, 'Xem show động vật biểu diễn', '11:00:00', '12:00:00', 'Show voi, chim, khỉ biểu diễn', 4),
(11, 'Ăn trưa tại nhà hàng', '12:00:00', '13:00:00', 'Buffet hải sản và món Việt', 5),
(11, 'Di chuyển đến cáp treo Hòn Thơm', '13:30:00', '14:00:00', 'Đi đến ga cáp treo', 6),
(11, 'Trải nghiệm cáp treo vượt biển', '14:00:00', '14:30:00', 'Cáp treo dài nhất thế giới, ngắm toàn cảnh biển đảo', 7),
(11, 'Tắm biển và chơi thể thao nước', '14:30:00', '17:00:00', 'Bơi lội, lặn ngắm san hô, chơi dù kéo, mô tô nước', 8),
(11, 'Trở về resort', '17:30:00', '18:00:00', 'Nghỉ ngơi tại resort', 9);

-- Ngày 2 - Tour 10: Bangkok City Tour
INSERT INTO hoat_dong_trong_ngay (lich_trinh_id, ten, gio_bat_dau, gio_ket_thuc, mo_ta, thu_tu)
VALUES
(16, 'Ăn sáng buffet', '07:00:00', '08:00:00', 'Bữa sáng tại khách sạn', 1),
(16, 'Tham quan Cung điện Hoàng gia', '08:30:00', '10:00:00', 'Grand Palace - cung điện lộng lẫy của hoàng gia Thái', 2),
(16, 'Tham quan Chùa Phật Ngọc', '10:00:00', '11:00:00', 'Wat Phra Kaew - ngôi chùa thiêng nhất Thái Lan', 3),
(16, 'Tham quan Chùa Vàng', '11:15:00', '12:00:00', 'Wat Traimit - tượng Phật vàng nguyên khối lớn nhất', 4),
(16, 'Ăn trưa buffet quốc tế', '12:15:00', '13:15:00', 'Buffet Thái - Âu - Á phong phú', 5),
(16, 'Tham quan Chùa Phật Nằm', '13:45:00', '14:45:00', 'Wat Pho - tượng Phật nằm dài 46m', 6),
(16, 'Đi thuyền trên sông Chao Phraya', '15:00:00', '16:00:00', 'Ngắm cảnh hai bờ sông, chụp ảnh', 7),
(16, 'Mua sắm King Power Duty Free', '16:30:00', '18:30:00', 'Shopping mỹ phẩm, thời trang, đồ lưu niệm', 8),
(16, 'Ăn tối và massage', '19:00:00', '21:00:00', 'Tự do ăn tối và massage Thái', 9);

