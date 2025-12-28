-- Insert data cho bảng tour
-- Lưu ý: Cần chạy các file sau trước khi chạy file này:
-- 1. nguoidung.sql
-- 2. nhacungcap.sql  
-- 3. danhmuctour.sql

INSERT INTO tour 
(tieu_de, mo_ta, danh_muc_id, so_ngay, so_dem, gia_nguoi_lon, gia_tre_em, don_vi_tien_te, trang_thai, noi_bat, nha_cung_cap_id, dang_hoat_dong)
VALUES
-- Tour 1: Hạ Long - Sapa (Tour nội địa)
('Hà Nội - Hạ Long - Sapa 5N4Đ: Khám Phá Miền Bắc Hùng Vĩ',
 'Trải nghiệm vẻ đẹp hùng vĩ của miền Bắc Việt Nam với Vịnh Hạ Long di sản thế giới và vùng cao Sapa thơ mộng. Chương trình bao gồm: Du thuyền ngủ đêm trên Vịnh Hạ Long, khám phá hang Sửng Sốt và làng chài, trekking ruộng bậc thang Sapa, tham quan bản Cát Cát, chinh phục đỉnh Fansipan bằng cáp treo. Thưởng thức ẩm thực đặc sản miền Bắc.',
 3, 5, 4, 12500000, 9500000, 'VND', 'cong_bo', TRUE, 
 '4d7c8912-3e14-46ef-a0c2-e4b3ed60ee47'::UUID, TRUE),

-- Tour 2: Đà Nẵng - Hội An - Huế (Tour nội địa)
('Đà Nẵng - Hội An - Huế 4N3Đ: Hành Trình Di Sản Miền Trung',
 'Khám phá ba thành phố di sản miền Trung: Đà Nẵng năng động, Hội An cổ kính, Huế cố đô. Điểm nhấn: Bà Nà Hills và cầu Vàng, phố cổ Hội An lung linh đèn lồng, Đại Nội Huế oai nghiêm, thưởng thức ẩm thực cung đình, tắm biển Mỹ Khê. Bao gồm vé cáp treo, vé tham quan các di tích.',
 6, 4, 3, 8900000, 6900000, 'VND', 'cong_bo', TRUE,
 '4d7c8912-3e14-46ef-a0c2-e4b3ed60ee47'::UUID, TRUE),

-- Tour 3: Phú Quốc (Du lịch nghỉ dưỡng)
('Phú Quốc 4N3Đ: Thiên Đường Biển Đảo - Resort 4 Sao',
 'Nghỉ dưỡng tại đảo Ngọc Phú Quốc với resort 4 sao view biển. Trải nghiệm: Safari Phú Quốc - vườn thú bán hoang dã lớn nhất Việt Nam, cáp treo Hòn Thơm vượt biển dài nhất thế giới, lặn ngắm san hô, câu cá, tham quan làng chài, chợ đêm Dinh Cậu, thưởng thức hải sản tươi sống.',
 4, 4, 3, 11500000, 8500000, 'VND', 'cong_bo', TRUE,
 'c91b086b-c8e6-49f5-91b2-ddf04b853a2d'::UUID, 3, TRUE),

-- Tour 4: Nha Trang (Du lịch nghỉ dưỡng)
('Nha Trang 3N2Đ: Biển Xanh - Cát Trắng - Nắng Vàng',
 'Tận hưởng kỳ nghỉ tại thành phố biển Nha Trang xinh đẹp. Hoạt động: Tour 4 đảo (Hòn Mun, Hòn Tằm, Hòn Một), lặn biển ngắm san hô, tắm bùn khoáng nóng I-Resort, tham quan VinWonders, tắm biển và thưởng thức hải sản. Phù hợp cho gia đình và nhóm bạn.',
 4, 3, 2, 6500000, 4500000, 'VND', 'cong_bo', FALSE,
 'c91b086b-c8e6-49f5-91b2-ddf04b853a2d'::UUID, 4, TRUE),

-- Tour 5: Đà Lạt (Du lịch khám phá)
('Đà Lạt 3N2Đ: Thành Phố Ngàn Hoa Lãng Mạn',
 'Khám phá thành phố ngàn hoa với khí hậu mát mẻ quanh năm. Điểm đến: Thác Datanla và trải nghiệm xe trượt, Đà Lạt Canyoning mạo hiểm, làng Cù Lần, đồi chè Cầu Đất, hồ Tuyền Lâm, check-in Quảng trường Lâm Viên, ga Đà Lạt, chợ đêm. Thích hợp cho cặp đôi và gia đình.',
 3, 3, 2, 5500000, 4000000, 'VND', 'cong_bo', FALSE,
 '9ad12bdd-4f4c-4868-885e-c9e41ed0e89c'::UUID, 5, TRUE),

-- Tour 6: Sài Gòn - Mỹ Tho - Cần Thơ (Du lịch sinh thái)
('Sài Gòn - Mỹ Tho - Cần Thơ 3N2Đ: Miền Tây Sông Nước',
 'Trải nghiệm miền Tây sông nước với chợ nổi Cái Răng, vườn trái cây, làng nghề, ẩm thực đặc sản. Hoạt động: Đi thuyền trên sông Tiền, nghe đờn ca tài tử, thăm vườn trái cây và thưởng thức, tham quan chợ nổi Cái Răng lúc bình minh, thăm nhà cổ, chùa Khmer.',
 1, 3, 2, 4900000, 3500000, 'VND', 'cong_bo', FALSE,
 '9ad12bdd-4f4c-4868-885e-c9e41ed0e89c'::UUID, 1, TRUE),

-- Tour 7: Ninh Bình - Hạ Long (Du lịch văn hóa)
('Ninh Bình - Hạ Long 4N3Đ: Hạ Long Trên Cạn & Trên Biển',
 'Khám phá cả hai vịnh di sản thế giới. Ninh Bình: Tràng An, Tam Cốc, Hang Múa, chùa Bái Đính lớn nhất Việt Nam. Hạ Long: Du thuyền qua đêm, hang Sửng Sốt, làng chài Cửa Vạn, chèo kayak, bơi lội. Trải nghiệm hoàn hảo cho người yêu thiên nhiên.',
 1, 4, 3, 9800000, 7500000, 'VND', 'cong_bo', TRUE,
 'faf013af-4377-445c-a235-0f60a88209fb'::UUID, 2, TRUE),

-- Tour 8: Côn Đảo (Du lịch tâm linh)
('Côn Đảo 3N2Đ: Đảo Thiêng - Biển Đẹp - Rừng Xanh',
 'Hành trình tâm linh và nghỉ dưỡng tại Côn Đảo. Tham quan: Nhà tù Côn Đảo, Nghĩa trang Hàng Dương, viếng mộ cô Sáu, chùa Núi Một. Hoạt động: Lặn biển ngắm san hô, trekking trong rừng quốc gia, tắm biển bãi Nhát, thưởng thức hải sản tươi ngon.',
 2, 3, 2, 7800000, 6000000, 'VND', 'cong_bo', FALSE,
 'faf013af-4377-445c-a235-0f60a88209fb'::UUID, 3, TRUE),

-- Tour 9: Quy Nhơn - Phú Yên (Du lịch nghỉ dưỡng)
('Quy Nhơn - Phú Yên 4N3Đ: Miền Duyên Hải Miền Trung',
 'Khám phá bờ biển hoang sơ của Quy Nhơn và Phú Yên. Điểm đến: Eo Gió, Kỳ Co, ghềnh Ráng, tháp Đôi Chăm, mũi Điện, đèo Cả, Vũng Rô. Tắm biển, chụp ảnh check-in, thưởng thức hải sản và các món ăn đặc sản miền Trung. Tour yên tĩnh, ít người.',
 1, 4, 3, 8200000, 6200000, 'VND', 'cong_bo', FALSE,
 '4baab3bf-b772-46e7-8db8-4a7db10215f4'::UUID, 4, TRUE),

-- Tour 10: Bangkok - Pattaya (Tour quốc tế)
('Bangkok - Pattaya 5N4Đ: Thái Lan Sôi Động',
 'Khám phá xứ sở chùa Vàng với Bangkok hiện đại và Pattaya sôi động. Tham quan: Cung điện Hoàng gia, chùa Vàng, chùa Phật Ngọc, chợ nổi Damnoen Saduak, Nong Nooch Garden, show Alcazar, đảo Coral, Safari World. Mua sắm tại Big C, King Power. Ăn uống hải sản buffet.',
 2, 5, 4, 9900000, 7900000, 'VND', 'cong_bo', TRUE,
 '4baab3bf-b772-46e7-8db8-4a7db10215f4'::UUID, 5, TRUE),

-- Tour 11: Singapore - Malaysia (Tour quốc tế)
('Singapore - Malaysia 6N5Đ: Sư Tử Biển Và Cao Nguyên',
 'Tour kết hợp hai quốc gia Đông Nam Á phát triển. Singapore: Marina Bay Sands, Gardens by the Bay, Sentosa, Universal Studios, Merlion. Malaysia: Genting Highlands, Batu Caves, tháp đôi Petronas, Putrajaya. Trải nghiệm văn hóa đa sắc tộc, ẩm thực phong phú.',
 1, 6, 5, 19500000, 15500000, 'VND', 'cong_bo', TRUE,
 '4baab3bf-b772-46e7-8db8-4a7db10215f4'::UUID, 1, TRUE),

-- Tour 12: Bali - Indonesia (Tour quốc tế)
('Bali 5N4Đ: Thiên Đường Nghỉ Dưỡng Nhiệt Đới',
 'Trải nghiệm hòn đảo thiên đường với văn hóa Hindu độc đáo. Tham quan: Đền Tanah Lot, Uluwatu, ruộng bậc thang Tegallalang, làng Ubud, tắm tại Tegenungan waterfall, xem múa Barong, thả diều ở bãi biển Seminyak. Massage truyền thống Bali, yoga.',
 3, 5, 4, 15900000, 12900000, 'VND', 'cong_bo', FALSE,
 '4baab3bf-b772-46e7-8db8-4a7db10215f4'::UUID, 2, TRUE),

-- Tour 13: Hàn Quốc Seoul - Nami (Tour quốc tế)
('Seoul - Nami - Everland 5N4Đ: Xứ Sở Kim Chi',
 'Khám phá văn hóa Hàn Quốc hiện đại. Tham quan: Cung điện Gyeongbokgung, làng Bukchon Hanok, tháp N Seoul, thử hanbok, đảo Nami lãng mạn, Everland, shopping tại Myeongdong, Dongdaemun. Thưởng thức BBQ Hàn Quốc, món Ginseng chicken. Tặng mặt nạ Hàn Quốc.',
 3, 5, 4, 18900000, 15900000, 'VND', 'cong_bo', TRUE,
 '9ad12bdd-4f4c-4868-885e-c9e41ed0e89c'::UUID, 3, TRUE),

-- Tour 14: Nhật Bản Tokyo - Osaka (Tour quốc tế)
('Tokyo - Fuji - Osaka 7N6Đ: Hoa Anh Đào & Geisha',
 'Tour Nhật Bản toàn diện từ Tokyo đến Osaka. Điểm đến: Núi Ph富士, đền Senso-ji, Shibuya crossing, Harajuku, Osaka Castle, Dotonbori, Nara công viên hươu, Kyoto Fushimi Inari. Tàu bullet Shinkansen, thưởng thức sushi, ramen, wagyu beef. Mua sắm tại Ginza.',
 3, 7, 6, 39900000, 35900000, 'VND', 'cong_bo', TRUE,
 '9ad12bdd-4f4c-4868-885e-c9e41ed0e89c'::UUID, 4, TRUE),

-- Tour 15: Team Building Vũng Tàu (Tour đoàn thể)
('Team Building Vũng Tàu 2N1Đ: Gắn Kết Đội Nhóm',
 'Chương trình team building chuyên nghiệp tại Vũng Tàu. Hoạt động: Trò chơi ngoài trời, thử thách nhóm, gala dinner bãi biển, BBQ hải sản, trò chơi cồng chiêng, thi nấu ăn, mini game. Phù hợp cho công ty, tổ chức 30-100 người. Có MC và đội ngũ hỗ trợ chuyên nghiệp.',
 3, 2, 1, 2500000, 1800000, 'VND', 'cong_bo', FALSE,
 'c91b086b-c8e6-49f5-91b2-ddf04b853a2d'::UUID, 5, TRUE),

-- Tour 16: Trăng mật Đà Nẵng (Tour trăng mật)
('Đà Nẵng - Hội An 4N3Đ: Trăng Mật Lãng Mạn',
 'Gói trăng mật đặc biệt cho cặp đôi mới cưới. Nghỉ resort 5 sao view biển, phòng honeymoon trang trí hoa, champagne. Hoạt động: Chụp ảnh cưới tại Bà Nà Hills, dạo phố cổ Hội An lãng mạn, dinner trên bãi biển, spa couple, tham quan Bán đảo Sơn Trà. Nhiều ưu đãi đặc biệt.',
 4, 4, 3, 18500000, 0, 'VND', 'cong_bo', TRUE,
 'c91b086b-c8e6-49f5-91b2-ddf04b853a2d'::UUID, 1, TRUE),

-- Tour 17: Tour gia đình Phú Quốc (Tour gia đình)
('Phú Quốc 5N4Đ: Kỳ Nghỉ Gia Đình Trọn Vẹn',
 'Tour thiết kế dành riêng cho gia đình có trẻ em. VinWonders Phú Quốc vui chơi cả ngày, Grand World check-in, Safari ngắm thú, tắm biển an toàn, câu cá, làm gốm tại làng nghề. Resort có kid club, bể bơi trẻ em. Bữa ăn phù hợp cho trẻ. Hướng dẫn viên kinh nghiệm với gia đình.',
 4, 5, 4, 16500000, 11500000, 'VND', 'cong_bo', TRUE,
 'c91b086b-c8e6-49f5-91b2-ddf04b853a2d'::UUID, 2, TRUE),

-- Tour 18: Du lịch học tập Hà Nội (Tour học sinh)
('Hà Nội 3N2Đ: Học Tập - Trải Nghiệm Thủ Đô',
 'Chương trình giáo dục dành cho học sinh, sinh viên. Tham quan: Lăng Bác, Văn Miếu Quốc Tử Giám, Bảo tàng Lịch sử, Bảo tàng Dân tộc học, nhà tù Hỏa Lò, Quốc hội. Học về lịch sử, văn hóa Việt Nam. Hoạt động ngoại khóa tại các di tích, giao lưu văn nghệ.',
 6, 3, 2, 3500000, 3000000, 'VND', 'cong_bo', FALSE,
 '4d7c8912-3e14-46ef-a0c2-e4b3ed60ee47'::UUID, 3, TRUE),

-- Tour 19: Phong Nha - Kẻ Bàng (Du lịch khám phá)
('Phong Nha - Kẻ Bàng 4N3Đ: Khám Phá Hang Động Kỳ Vĩ',
 'Thám hiểm vườn quốc gia di sản thiên nhiên thế giới. Khám phá: Hang Phong Nha, động Thiên Đường dài nhất Châu Á, hang Tám Cô, suối Nước Moọc, vườn thực vật, zipline qua sông. Tour dành cho người yêu thích khám phá và mạo hiểm. Cảnh đẹp hoang sơ, độc đáo.',
 3, 4, 3, 7500000, 5500000, 'VND', 'cong_bo', FALSE,
 '4d7c8912-3e14-46ef-a0c2-e4b3ed60ee47'::UUID, 4, TRUE),

-- Tour 20: Tour nháp - Châu Âu (đang soạn thảo)
('Châu Âu 10N9Đ: Pháp - Thụy Sĩ - Ý (Tour Cao Cấp)',
 'Tour cao cấp khám phá ba quốc gia Châu Âu hấp dẫn nhất. Paris lãng mạn, Thụy Sĩ thiên đường, Ý cổ kính. (Đang cập nhật chi tiết hành trình)',
 3, 10, 9, 89000000, 85000000, 'VND', 'nhap', FALSE,
 '4d7c8912-3e14-46ef-a0c2-e4b3ed60ee47'::UUID, 5, TRUE),

-- Tour 21: Tour lưu trữ - Mùa Covid
('Tour Tết Nguyên Đán 2020: Đà Nẵng 4N3Đ',
 'Tour đã kết thúc do dịch Covid-19. Lưu trữ để tham khảo.',
 3, 4, 3, 8500000, 6500000, 'VND', 'luu_tru', FALSE,
 '4d7c8912-3e14-46ef-a0c2-e4b3ed60ee47'::UUID, 1, FALSE);

