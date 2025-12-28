-- Insert data cho người dùng (cần thiết cho tour)
-- Password mặc định: "Password123!" (đã được hash với bcrypt)

INSERT INTO nguoi_dung (id, ho_ten, email, mat_khau_ma_hoa, so_dien_thoai, vai_tro, dang_hoat_dong, xac_thuc, ngay_tao)
VALUES
-- Admin/Quản trị viên
('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'Nguyễn Văn Admin', 'admin@travia.vn', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '0901234567', 'quan_tri', TRUE, TRUE, NOW()),

-- Nhà cung cấp
('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', 'Nguyễn Văn Minh', 'minh@saigontourist.net', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '0987654321', 'nha_cung_cap', TRUE, TRUE, NOW()),
('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a13', 'Trần Thị Thu', 'thu@vietravel.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '0976543210', 'nha_cung_cap', TRUE, TRUE, NOW()),
('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a14', 'Lê Văn Hải', 'hai@benthanhtourist.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '0965432109', 'nha_cung_cap', TRUE, TRUE, NOW()),
('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', 'Phạm Hồng Phúc', 'phuc@fiditour.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '0954321098', 'nha_cung_cap', TRUE, TRUE, NOW()),
('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a16', 'Nguyễn Thanh Tùng', 'tung@hanoitourist.com.vn', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '0943210987', 'nha_cung_cap', TRUE, TRUE, NOW()),

-- Khách hàng
('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a17', 'Võ Thị Lan', 'lan.vo@gmail.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '0932109876', 'khach_hang', TRUE, TRUE, NOW()),
('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a18', 'Hoàng Văn Nam', 'nam.hoang@gmail.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '0921098765', 'khach_hang', TRUE, TRUE, NOW()),
('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a19', 'Đỗ Thị Mai', 'mai.do@gmail.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '0910987654', 'khach_hang', TRUE, TRUE, NOW());

