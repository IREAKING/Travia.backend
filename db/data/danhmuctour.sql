-- insert data cho danh mục tour
INSERT INTO danh_muc_tour (ten, mo_ta, anh, dang_hoat_dong, ngay_tao)
VALUES
-- 1. Theo phạm vi địa lý
('Tour nội địa', 'Các chuyến du lịch trong nước như Hà Nội – Đà Nẵng – TP.HCM.', 'noidia.jpg', TRUE, NOW()),
('Tour quốc tế', 'Các tour ra nước ngoài như Nhật Bản, Hàn Quốc, Thái Lan.', 'quocte.jpg', TRUE, NOW()),
('Tour inbound', 'Tour đón khách quốc tế đến tham quan Việt Nam.', 'inbound.jpg', TRUE, NOW()),

-- 2. Theo loại hình du lịch
('Du lịch nghỉ dưỡng', 'Thư giãn tại resort, bãi biển, suối khoáng.', 'nghiduong.jpg', TRUE, NOW()),
('Du lịch khám phá - mạo hiểm', 'Leo núi, trekking, hang động, lặn biển.', 'khampha.jpg', TRUE, NOW()),
('Du lịch sinh thái', 'Trải nghiệm thiên nhiên, vườn quốc gia, làng nghề.', 'sinhthai.jpg', TRUE, NOW()),
('Du lịch văn hóa - lịch sử', 'Tham quan di tích, lễ hội, bảo tàng.', 'vanhoa.jpg', TRUE, NOW()),
('Du lịch tâm linh', 'Hành hương, viếng chùa, đền thờ.', 'tamlinh.jpg', TRUE, NOW()),
('Du lịch ẩm thực', 'Khám phá và thưởng thức đặc sản vùng miền.', 'amthuc.jpg', TRUE, NOW()),
('Du lịch MICE', 'Tour kết hợp hội nghị, hội thảo, triển lãm.', 'mice.jpg', TRUE, NOW()),
('Du lịch học tập - trải nghiệm', 'Dành cho học sinh, sinh viên học hỏi và khám phá.', 'hoctap.jpg', TRUE, NOW()),
('Du lịch chăm sóc sức khỏe', 'Tour spa, yoga, tắm khoáng, thiền.', 'suckhoe.jpg', TRUE, NOW()),
('Du lịch cộng đồng', 'Sống và trải nghiệm cùng người dân địa phương.', 'congdong.jpg', TRUE, NOW()),

-- 5. Theo đối tượng khách hàng
('Tour gia đình', 'Dành cho các thành viên trong gia đình, phù hợp trẻ em.', 'giadinh.jpg', TRUE, NOW()),
('Tour trăng mật', 'Dành cho cặp đôi, không gian riêng tư, lãng mạn.', 'trangmat.jpg', TRUE, NOW()),
('Tour đoàn thể - team building', 'Dành cho công ty, tổ chức, kết nối tập thể.', 'teambuilding.jpg', TRUE, NOW()),
('Tour học sinh - sinh viên', 'Giá hợp lý, nội dung học tập kết hợp vui chơi.', 'hocsinh.jpg', TRUE, NOW()),
('Tour cao cấp', 'Dịch vụ 5 sao, resort sang trọng, xe riêng.', 'caocap.jpg', TRUE, NOW());