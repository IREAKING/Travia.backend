-- Insert data cho lịch trình tour
-- Lưu ý: Cần chạy tour.sql trước khi chạy file này

-- Tour 1: Hạ Long - Sapa (5N4Đ)
INSERT INTO lich_trinh (tour_id, ngay_thu, tieu_de, mo_ta, gio_bat_dau, gio_ket_thuc, dia_diem, thong_tin_luu_tru)
VALUES
(1, 1, 'Hà Nội - Đón Khách - City Tour', 
 'Đón khách tại sân bay Nội Bài, di chuyển về khách sạn nhận phòng. Buổi chiều tham quan Văn Miếu Quốc Tử Giám, Hồ Hoàn Kiếm, Đền Ngọc Sơn, dạo phố cổ. Tối thưởng thức bún chả Hà Nội và xem múa rối nước.',
 '14:00:00', '21:00:00', 'Hà Nội', 'Khách sạn 3* trung tâm Hà Nội'),

(1, 2, 'Hà Nội - Vịnh Hạ Long', 
 'Sau bữa sáng, xe đưa đoàn đi Hạ Long (3.5 giờ). Lên tàu du lịch, thưởng thức hải sản trên tàu. Tham quan hang Sửng Sốt, làng chài Cửa Vạn, chèo kayak, bơi lội. Tối BBQ trên boong tàu, câu mực, ngắm bầu trời sao.',
 '07:00:00', '22:00:00', 'Vịnh Hạ Long', 'Ngủ đêm trên du thuyền 4*'),

(1, 3, 'Hạ Long - Hà Nội - Sapa', 
 'Xem bình minh trên vịnh, ăn sáng trên tàu. Tham quan thêm một số đảo đá và hang động. Trở về bến, xe đưa về Hà Nội. Chiều đi tàu hỏa hoặc xe limousine lên Sapa (5 giờ). Đến Sapa tối, nhận phòng nghỉ ngơi.',
 '06:00:00', '22:00:00', 'Hạ Long - Sapa', 'Khách sạn 3* Sapa view núi'),

(1, 4, 'Sapa - Chinh Phục Fansipan - Bản Cát Cát', 
 'Sáng đi cáp treo lên đỉnh Fansipan (3143m) - nóc nhà Đông Dương. Ngắm cảnh trên cao, check-in. Trưa về thị trấn ăn trưa. Chiều đi bộ xuống bản Cát Cát, tìm hiểu văn hóa dân tộc H Mông, thác nước, nhà máy thủy điện cổ.',
 '07:00:00', '18:00:00', 'Sapa', 'Khách sạn 3* Sapa'),

(1, 5, 'Sapa - Hà Nội - Tiễn Khách', 
 'Sau bữa sáng, tự do chụp ảnh tại Sapa. Trưa về Hà Nội bằng xe limousine. Mua sắm quà lưu niệm tại chợ Đồng Xuân. Tiễn khách ra sân bay về. Kết thúc chuyến đi.',
 '08:00:00', '20:00:00', 'Sapa - Hà Nội', NULL);

-- Tour 2: Đà Nẵng - Hội An - Huế (4N3Đ)
INSERT INTO lich_trinh (tour_id, ngay_thu, tieu_de, mo_ta, gio_bat_dau, gio_ket_thuc, dia_diem, thong_tin_luu_tru)
VALUES
(2, 1, 'Đà Nẵng - Đón Khách - Bà Nà Hills', 
 'Đón khách tại sân bay Đà Nẵng. Lên Bà Nà Hills bằng cáp treo, tham quan Cầu Vàng nổi tiếng, Làng Pháp, Vườn hoa Le Jardin, Fantasy Park. Trưa buffet tại Bà Nà. Chiều xuống núi, tắm biển Mỹ Khê. Check-in khách sạn.',
 '10:00:00', '19:00:00', 'Đà Nẵng', 'Khách sạn 4* view biển Đà Nẵng'),

(2, 2, 'Đà Nẵng - Hội An', 
 'Sáng tham quan Ngũ Hành Sơn, chùa Linh Ứng Bán đảo Sơn Trà. Trưa đi Hội An, check-in khách sạn. Chiều dạo phố cổ Hội An: chùa Cầu, nhà cổ, hội quán. Tối thả đèn hoa đăng trên sông Hoài, thưởng thức cao lầu, bánh mì Phượng.',
 '08:00:00', '21:00:00', 'Đà Nẵng - Hội An', 'Khách sạn 4* Hội An'),

(2, 3, 'Hội An - Huế', 
 'Sáng tự do khám phá Hội An hoặc tắm biển An Bàng. Trưa đi Huế qua đèo Hải Vân tuyệt đẹp. Chiều tham quan Đại Nội, lăng Khải Định, chùa Thiên Mụ. Tối thưởng thức ẩm thực cung đình Huế.',
 '10:00:00', '21:00:00', 'Hội An - Huế', 'Khách sạn 4* trung tâm Huế'),

(2, 4, 'Huế - Đà Nẵng - Tiễn Khách', 
 'Sáng tham quan lăng Tự Đức, chợ Đông Ba. Trưa về Đà Nẵng, mua sắm tại chợ Hàn, Lotte Mart. Chiều tiễn khách ra sân bay. Kết thúc chuyến đi.',
 '08:00:00', '17:00:00', 'Huế - Đà Nẵng', NULL);

-- Tour 3: Phú Quốc (4N3Đ)
INSERT INTO lich_trinh (tour_id, ngay_thu, tieu_de, mo_ta, gio_bat_dau, gio_ket_thuc, dia_diem, thong_tin_luu_tru)
VALUES
(3, 1, 'Phú Quốc - Đón Khách - Nghỉ Dưỡng', 
 'Đón khách tại sân bay Phú Quốc, đưa về resort check-in. Tự do nghỉ ngơi, tắm biển, tắm hồ bơi, tận hưởng các tiện ích của resort. Tối BBQ hải sản tại bãi biển.',
 '14:00:00', '22:00:00', 'Phú Quốc', 'Resort 4* Phú Quốc'),

(3, 2, 'Tour Khám Phá Bắc Đảo - Safari', 
 'Sáng tham quan Safari Phú Quốc - vườn thú bán hoang dã lớn nhất VN, xem show thú biểu diễn. Trưa ăn tại nhà hàng. Chiều đi cáp treo Hòn Thơm vượt biển dài nhất thế giới, tắm biển, chơi các trò chơi nước.',
 '08:00:00', '18:00:00', 'Phú Quốc', 'Resort 4* Phú Quốc'),

(3, 3, 'Tour Nam Đảo - Lặn Biển', 
 'Sáng đi tour Nam đảo, lặn biển ngắm san hô tại Hòn Móng Tay, Hòn Gầm Ghì. Câu cá, bơi lội. Trưa picnic trên đảo. Chiều về resort nghỉ ngơi. Tối tự do dạo chợ đêm Dinh Cậu.',
 '08:00:00', '20:00:00', 'Phú Quốc', 'Resort 4* Phú Quốc'),

(3, 4, 'Tự Do - Tiễn Khách', 
 'Sáng tự do tắm biển, nghỉ dưỡng. Trưa trả phòng, mua sắm tại chợ Dương Đông. Tiễn khách ra sân bay về. Kết thúc chuyến đi.',
 '10:00:00', '16:00:00', 'Phú Quốc', NULL);

-- Tour 10: Bangkok - Pattaya (5N4Đ)
INSERT INTO lich_trinh (tour_id, ngay_thu, tieu_de, mo_ta, gio_bat_dau, gio_ket_thuc, dia_diem, thong_tin_luu_tru)
VALUES
(10, 1, 'Bangkok - Đón Khách - Chợ Đêm', 
 'Đón khách tại sân bay Suvarnabhumi Bangkok. Đưa về khách sạn nhận phòng. Nghỉ ngơi. Tối đi chợ đêm Asiatique, mua sắm, ăn uống. Massage Thái cổ truyền.',
 '18:00:00', '23:00:00', 'Bangkok', 'Khách sạn 4* Bangkok'),

(10, 2, 'Bangkok City Tour', 
 'Sáng tham quan Cung điện Hoàng gia Grand Palace, chùa Phật Ngọc Wat Phra Kaew, chùa Vàng Wat Traimit. Trưa ăn buffet quốc tế. Chiều tham quan chùa Phật Nằm Wat Pho, đi thuyền trên sông Chao Phraya. Mua sắm tại King Power Duty Free.',
 '08:00:00', '20:00:00', 'Bangkok', 'Khách sạn 4* Bangkok'),

(10, 3, 'Bangkok - Pattaya - Show Alcazar', 
 'Sáng đi Pattaya (2 giờ), tham quan chợ nổi 4 miền, Nong Nooch Garden xem show voi và văn hóa Thái. Trưa ăn buffet hải sản. Chiều đi đảo Coral, lặn biển, chơi sport nước. Tối xem show Alcazar nổi tiếng.',
 '08:00:00', '22:00:00', 'Bangkok - Pattaya', 'Khách sạn 4* Pattaya'),

(10, 4, 'Pattaya - Bangkok - Safari World', 
 'Sáng về Bangkok, tham quan Safari World và Marine Park, xem show cướp biển, show cá heo. Trưa ăn trong khu. Chiều mua sắm tại Big C, Central World. Tối ăn tối buffet lẩu.',
 '08:00:00', '21:00:00', 'Pattaya - Bangkok', 'Khách sạn 4* Bangkok'),

(10, 5, 'Bangkok - Mua Sắm - Tiễn Khách', 
 'Sáng tự do nghỉ ngơi hoặc mua sắm. Trưa check-out, mua sắm tại MBK Center, Platinum. Tiễn khách ra sân bay về. Kết thúc chuyến đi.',
 '10:00:00', '17:00:00', 'Bangkok', NULL);

