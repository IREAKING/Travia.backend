-- Insert data cho ảnh tour
-- Lưu ý: Cần chạy tour.sql trước khi chạy file này

INSERT INTO anh_tour (tour_id, link, mo_ta_alt, la_anh_chinh, thu_tu_hien_thi)
VALUES
-- Tour 1: Hạ Long - Sapa
(1, 'https://images.pexels.com/photos/2422461/pexels-photo-2422461.jpeg', 'Vịnh Hạ Long tuyệt đẹp', TRUE, 1),
(1, 'https://images.pexels.com/photos/2412610/pexels-photo-2412610.jpeg', 'Ruộng bậc thang Sapa', FALSE, 2),
(1, 'https://images.pexels.com/photos/3998365/pexels-photo-3998365.jpeg', 'Phố cổ Hà Nội', FALSE, 3),
(1, 'https://images.pexels.com/photos/2422461/pexels-photo-2422461.jpeg', 'Du thuyền Hạ Long', FALSE, 4),

-- Tour 2: Đà Nẵng - Hội An - Huế
(2, 'https://images.pexels.com/photos/2412606/pexels-photo-2412606.jpeg', 'Cầu Vàng Bà Nà Hills', TRUE, 1),
(2, 'https://images.pexels.com/photos/1549326/pexels-photo-1549326.jpeg', 'Phố cổ Hội An về đêm', FALSE, 2),
(2, 'https://images.pexels.com/photos/2412609/pexels-photo-2412609.jpeg', 'Đại Nội Huế', FALSE, 3),
(2, 'https://images.pexels.com/photos/2412606/pexels-photo-2412606.jpeg', 'Bãi biển Mỹ Khê', FALSE, 4),

-- Tour 3: Phú Quốc
(3, 'https://images.pexels.com/photos/2412621/pexels-photo-2412621.jpeg', 'Bãi biển Phú Quốc', TRUE, 1),
(3, 'https://images.pexels.com/photos/2412621/pexels-photo-2412621.jpeg', 'Cáp treo Hòn Thơm', FALSE, 2),
(3, 'https://images.pexels.com/photos/2412621/pexels-photo-2412621.jpeg', 'Safari Phú Quốc', FALSE, 3),
(3, 'https://images.pexels.com/photos/2412621/pexels-photo-2412621.jpeg', 'Hoàng hôn Phú Quốc', FALSE, 4),

-- Tour 4: Nha Trang
(4, 'https://images.pexels.com/photos/2412614/pexels-photo-2412614.jpeg', 'Biển Nha Trang xanh ngắt', TRUE, 1),
(4, 'https://images.pexels.com/photos/2412614/pexels-photo-2412614.jpeg', 'Tour 4 đảo', FALSE, 2),
(4, 'https://images.pexels.com/photos/2412614/pexels-photo-2412614.jpeg', 'VinWonders Nha Trang', FALSE, 3),

-- Tour 5: Đà Lạt
(5, 'https://images.pexels.com/photos/2412608/pexels-photo-2412608.jpeg', 'Đà Lạt thành phố ngàn hoa', TRUE, 1),
(5, 'https://images.pexels.com/photos/2412608/pexels-photo-2412608.jpeg', 'Thác Datanla', FALSE, 2),
(5, 'https://images.pexels.com/photos/2412608/pexels-photo-2412608.jpeg', 'Đồi chè Cầu Đất', FALSE, 3),

-- Tour 6: Mỹ Tho - Cần Thơ
(6, 'https://images.pexels.com/photos/2412604/pexels-photo-2412604.jpeg', 'Chợ nổi Cái Răng', TRUE, 1),
(6, 'https://images.pexels.com/photos/2412604/pexels-photo-2412604.jpeg', 'Vườn trái cây miền Tây', FALSE, 2),
(6, 'https://images.pexels.com/photos/2412604/pexels-photo-2412604.jpeg', 'Du thuyền sông Tiền', FALSE, 3),

-- Tour 7: Ninh Bình - Hạ Long
(7, 'https://images.pexels.com/photos/2412623/pexels-photo-2412623.jpeg', 'Tràng An Ninh Bình', TRUE, 1),
(7, 'https://images.pexels.com/photos/2412623/pexels-photo-2412623.jpeg', 'Hang Múa view đẹp', FALSE, 2),
(7, 'https://images.pexels.com/photos/2422461/pexels-photo-2422461.jpeg', 'Vịnh Hạ Long', FALSE, 3),

-- Tour 10: Bangkok - Pattaya
(10, 'https://images.pexels.com/photos/1031659/pexels-photo-1031659.jpeg', 'Chùa Vàng Bangkok', TRUE, 1),
(10, 'https://images.pexels.com/photos/1031659/pexels-photo-1031659.jpeg', 'Pattaya về đêm', FALSE, 2),
(10, 'https://images.pexels.com/photos/1031659/pexels-photo-1031659.jpeg', 'Show Alcazar', FALSE, 3),

-- Tour 11: Singapore - Malaysia
(11, 'https://images.pexels.com/photos/1031659/pexels-photo-1031659.jpeg', 'Marina Bay Sands Singapore', TRUE, 1),
(11, 'https://images.pexels.com/photos/1031659/pexels-photo-1031659.jpeg', 'Gardens by the Bay', FALSE, 2),
(11, 'https://images.pexels.com/photos/1031659/pexels-photo-1031659.jpeg', 'Petronas Twin Towers', FALSE, 3),

-- Tour 12: Bali
(12, 'https://images.pexels.com/photos/2166559/pexels-photo-2166559.jpeg', 'Đền Tanah Lot Bali', TRUE, 1),
(12, 'https://images.pexels.com/photos/2166559/pexels-photo-2166559.jpeg', 'Ruộng bậc thang Tegallalang', FALSE, 2),
(12, 'https://images.pexels.com/photos/2166559/pexels-photo-2166559.jpeg', 'Bãi biển Seminyak', FALSE, 3),

-- Tour 13: Seoul - Nami
(13, 'https://images.pexels.com/photos/237211/pexels-photo-237211.jpeg', 'Cung điện Gyeongbokgung', TRUE, 1),
(13, 'https://images.pexels.com/photos/237211/pexels-photo-237211.jpeg', 'Đảo Nami lãng mạn', FALSE, 2),
(13, 'https://images.pexels.com/photos/237211/pexels-photo-237211.jpeg', 'Tháp N Seoul', FALSE, 3),

-- Tour 14: Tokyo - Osaka
(14, 'https://images.pexels.com/photos/2614818/pexels-photo-2614818.jpeg', 'Núi Phú Sĩ Nhật Bản', TRUE, 1),
(14, 'https://images.pexels.com/photos/2614818/pexels-photo-2614818.jpeg', 'Shibuya Crossing Tokyo', FALSE, 2),
(14, 'https://images.pexels.com/photos/2614818/pexels-photo-2614818.jpeg', 'Osaka Castle', FALSE, 3),
(14, 'https://images.pexels.com/photos/2614818/pexels-photo-2614818.jpeg', 'Fushimi Inari Kyoto', FALSE, 4),

-- Tour 16: Trăng mật Đà Nẵng
(16, 'https://images.pexels.com/photos/2412606/pexels-photo-2412606.jpeg', 'Sunset Đà Nẵng lãng mạn', TRUE, 1),
(16, 'https://images.pexels.com/photos/1549326/pexels-photo-1549326.jpeg', 'Hội An đèn lồng', FALSE, 2),
(16, 'https://images.pexels.com/photos/2412606/pexels-photo-2412606.jpeg', 'Bà Nà Hills', FALSE, 3),

-- Tour 17: Gia đình Phú Quốc
(17, 'https://images.pexels.com/photos/2412621/pexels-photo-2412621.jpeg', 'VinWonders Phú Quốc', TRUE, 1),
(17, 'https://images.pexels.com/photos/2412621/pexels-photo-2412621.jpeg', 'Safari Phú Quốc', FALSE, 2),
(17, 'https://images.pexels.com/photos/2412621/pexels-photo-2412621.jpeg', 'Grand World', FALSE, 3);

