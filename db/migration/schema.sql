-- Tạo các ENUM types
CREATE TYPE vai_tro_nguoi_dung AS ENUM ('khach_hang', 'quan_tri', 'nha_cung_cap');

-- Cấu hình tìm kiếm toàn văn cho tiếng Việt (dựa trên unaccent + simple)
CREATE EXTENSION IF NOT EXISTS unaccent;
CREATE EXTENSION IF NOT EXISTS pgcrypto;
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_ts_config WHERE cfgname = 'vietnamese') THEN
        CREATE TEXT SEARCH CONFIGURATION vietnamese ( COPY = pg_catalog.simple );
        ALTER TEXT SEARCH CONFIGURATION vietnamese
            ALTER MAPPING FOR hword, hword_part, word WITH unaccent, simple;
    END IF;
END
$$;

-- Người dùng
CREATE TABLE nguoi_dung (
    id UUID PRIMARY KEY default gen_random_uuid(),
    ho_ten VARCHAR(255) not NULL,
    email VARCHAR(255) UNIQUE not null,               -- Có thể NULL nếu chỉ login OAuth
    mat_khau_ma_hoa TEXT not null,                    -- Mật khẩu đã mã hoá
    so_dien_thoai VARCHAR(50),
    vai_tro vai_tro_nguoi_dung DEFAULT 'khach_hang',-- khach_hang, quan_tri, nha_cung_cap
    dang_hoat_dong BOOLEAN DEFAULT TRUE,          -- Soft delete
    xac_thuc BOOLEAN default FALSE,
    ngay_tao TIMESTAMP DEFAULT NOW(),
    ngay_cap_nhat TIMESTAMP DEFAULT NOW()
);

-- Nhà cung cấp OAuth (Google, Facebook, ...)
CREATE TABLE nha_cung_cap_oauth (
    id UUID PRIMARY KEY default gen_random_uuid(),
    ten VARCHAR(50) UNIQUE NOT NULL,         -- google, facebook, github, apple
    client_id TEXT NOT NULL,
    client_secret TEXT NOT NULL,
    redirect_uri TEXT NOT NULL,
    ngay_tao TIMESTAMP DEFAULT NOW(),
    ngay_cap_nhat TIMESTAMP DEFAULT NOW()
);

-- Tài khoản OAuth của người dùng
CREATE TABLE tai_khoan_oauth (
    id UUID PRIMARY KEY default gen_random_uuid(),
    nguoi_dung_id UUID NOT NULL REFERENCES nguoi_dung(id) ON DELETE CASCADE,
    nha_cung_cap_id UUID NOT NULL REFERENCES nha_cung_cap_oauth(id) ON DELETE CASCADE,
    ma_tai_khoan VARCHAR(255) NOT NULL,  -- sub/id từ Google, FB...
    email VARCHAR(255),
    anh_dai_dien TEXT,
    access_token TEXT,              --lấy thông tin user, lấy ảnh profile, đọc Google Calendar
    refresh_token TEXT,
    het_han_token TIMESTAMP,
    ngay_tao TIMESTAMP DEFAULT NOW(),
    ngay_cap_nhat TIMESTAMP DEFAULT NOW(),
    UNIQUE (nha_cung_cap_id, ma_tai_khoan)
);
CREATE TABLE phien_dang_nhap (
    id SERIAL PRIMARY KEY,                               -- Khóa chính
    nguoi_dung_id UUID NOT NULL REFERENCES nguoi_dung(id) ON DELETE CASCADE, -- Liên kết với bảng người dùng
    access_token_hash TEXT NOT NULL,                     -- Hash của access token (bảo mật)
    refresh_token_hash TEXT NOT NULL,                    -- Hash của refresh token (bảo mật)
    thoi_han_token TIMESTAMP NOT NULL,                   -- Thời gian hết hạn token
    thong_tin_thiet_bi TEXT,                             -- Thiết bị, IP, trình duyệt
    dang_hoat_dong BOOLEAN DEFAULT TRUE,                      -- Soft delete
    ngay_tao TIMESTAMP DEFAULT CURRENT_TIMESTAMP,        -- Ngày tạo phiên
    ngay_cap_nhat TIMESTAMP DEFAULT CURRENT_TIMESTAMP    -- Ngày cập nhật
);


-- Điểm đến (Destination)
CREATE TABLE diem_den (
    id SERIAL PRIMARY KEY,
    ten VARCHAR(255) NOT NULL,       -- Ví dụ: Hà Nội, Paris, Tokyo
    quoc_gia VARCHAR(100),
    khu_vuc VARCHAR(100),
    mo_ta TEXT,
    anh TEXT,
    vi_do DECIMAL(9,6),
    kinh_do DECIMAL(9,6),
    ngay_tao TIMESTAMP DEFAULT NOW(),
    ngay_cap_nhat TIMESTAMP DEFAULT NOW()
);

-- Danh mục tour
CREATE TABLE danh_muc_tour (
    id SERIAL PRIMARY KEY,
    ten VARCHAR(50) NOT NULL UNIQUE,
    mo_ta TEXT,
    anh VARCHAR(255),
    dang_hoat_dong BOOLEAN DEFAULT TRUE,
    ngay_tao TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE nha_cung_cap (
    id SERIAL PRIMARY KEY,
    ten VARCHAR(255) NOT NULL,             -- Tên công ty/nhà cung cấp
    email VARCHAR(255) UNIQUE,
    so_dien_thoai VARCHAR(50),
    dia_chi TEXT,
    website VARCHAR(255),
    mo_ta TEXT,
    logo_url TEXT,
    trang_thai VARCHAR(20) DEFAULT 'hoat_dong' CHECK (trang_thai IN ('hoat_dong', 'tam_dung', 'ngung')),
    nguoi_dai_dien VARCHAR(255),           -- Người liên hệ chính
    dang_hoat_dong BOOLEAN DEFAULT TRUE,        -- Soft delete
    ngay_tao TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ngay_cap_nhat TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tour
CREATE TABLE tour (
    id SERIAL PRIMARY KEY,
    tieu_de VARCHAR(200) NOT NULL,
    mo_ta TEXT,
    danh_muc_id INTEGER REFERENCES danh_muc_tour(id),
    so_ngay INTEGER NOT NULL CHECK (so_ngay > 0),
    so_dem INTEGER NOT NULL CHECK (so_dem >= 0),
    gia_moi_nguoi DECIMAL(10, 2) NOT NULL CHECK (gia_moi_nguoi > 0),
    don_vi_tien_te VARCHAR(3) DEFAULT 'VND',
    trang_thai VARCHAR(20) DEFAULT 'nhap' CHECK (trang_thai IN ('nhap', 'cong_bo', 'luu_tru')),
    noi_bat BOOLEAN DEFAULT FALSE,
    nguoi_tao_id UUID REFERENCES nguoi_dung(id),
    nha_cung_cap_id INT REFERENCES nha_cung_cap(id),
    dang_hoat_dong BOOLEAN DEFAULT TRUE,          -- Soft delete
    ngay_tao TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ngay_cap_nhat TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_tour_ngay_cap_nhat CHECK (ngay_cap_nhat >= ngay_tao)
);

-- Cài đặt số lượng khách
CREATE TABLE cau_hinh_nhom_tour (
    id SERIAL PRIMARY KEY,
    tour_id INT NOT NULL REFERENCES tour(id) ON DELETE CASCADE,
    so_nho_nhat INT DEFAULT 1 CHECK (so_nho_nhat > 0),
    so_lon_nhat INT DEFAULT 20 CHECK (so_lon_nhat > 0),
    ngay_tao TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ngay_cap_nhat TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Giảm giá tour
CREATE TABLE giam_gia_tour (
    id SERIAL PRIMARY KEY,
    tour_id INT NOT NULL REFERENCES tour(id) ON DELETE CASCADE,
    phan_tram DECIMAL(5,2) CHECK (phan_tram >= 0 AND phan_tram <= 100),
    ngay_bat_dau DATE NOT NULL,
    ngay_ket_thuc DATE NOT NULL,
    ngay_tao TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ngay_cap_nhat TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CHECK (ngay_bat_dau <= ngay_ket_thuc)
);

-- Ảnh tour
CREATE TABLE anh_tour (
    id SERIAL PRIMARY KEY,
    tour_id INTEGER NOT NULL REFERENCES tour(id) ON DELETE CASCADE,
    link VARCHAR(255) NOT NULL,
    mo_ta_alt VARCHAR(100),
    la_anh_chinh BOOLEAN DEFAULT FALSE,
    thu_tu_hien_thi INTEGER DEFAULT 0,
    ngay_tao TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Liên kết tour với điểm đến
CREATE TABLE tour_diem_den (
    tour_id INTEGER NOT NULL REFERENCES tour(id) ON DELETE CASCADE,
    diem_den_id INTEGER NOT NULL REFERENCES diem_den(id) ON DELETE CASCADE,
    thu_tu_tham_quan INTEGER DEFAULT 0,
    thoi_gian_luu_tru_gio INTEGER DEFAULT 0,
    PRIMARY KEY (tour_id, diem_den_id)
);

-- Lịch trình tour (theo ngày)
CREATE TABLE lich_trinh_tour (
    id SERIAL PRIMARY KEY,
    tour_id INTEGER NOT NULL REFERENCES tour(id) ON DELETE CASCADE,
    ngay_thu INTEGER NOT NULL CHECK (ngay_thu > 0),
    tieu_de VARCHAR(200) NOT NULL,
    mo_ta TEXT,
    gio_bat_dau TIME,
    gio_ket_thuc TIME,
    dia_diem TEXT,
    thong_tin_luu_tru TEXT,
    ngay_tao TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ngay_cap_nhat TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tour_id, ngay_thu)
);

-- Hoạt động trong lịch trình
CREATE TABLE hoat_dong_lich_trinh (
    id SERIAL PRIMARY KEY,
    lich_trinh_id INTEGER NOT NULL REFERENCES lich_trinh_tour(id) ON DELETE CASCADE,
    ten VARCHAR(200) NOT NULL,
    gio_bat_dau TIME,
    gio_ket_thuc TIME,
    mo_ta TEXT,
    thu_tu INTEGER DEFAULT 0,
    ngay_tao TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Lịch khởi hành của tour
CREATE TYPE trang_thai_khoi_hanh AS ENUM ('len_lich', 'xac_nhan', 'huy', 'hoan_thanh');

CREATE TABLE khoi_hanh_tour (
    id SERIAL PRIMARY KEY,
    tour_id INT NOT NULL REFERENCES tour(id) ON DELETE CASCADE,
    ngay_khoi_hanh DATE NOT NULL,
    ngay_ket_thuc DATE NOT NULL CHECK (ngay_ket_thuc >= ngay_khoi_hanh),
    suc_chua INT NOT NULL CHECK (suc_chua > 0),
    trang_thai trang_thai_khoi_hanh DEFAULT 'len_lich',
    huong_dan_vien_id UUID REFERENCES nguoi_dung(id) ON DELETE SET NULL,
    gia_dac_biet DECIMAL(10,2) CHECK (gia_dac_biet >= 0),
    ghi_chu TEXT,
    ngay_tao TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ngay_cap_nhat TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- Đặt chỗ tour
CREATE TYPE trang_thai_dat_cho AS ENUM ('cho_xac_nhan', 'da_xac_nhan', 'da_thanh_toan', 'da_huy', 'hoan_thanh');

CREATE TABLE dat_cho_tour (
    id SERIAL PRIMARY KEY,
    nguoi_dung_id UUID NOT NULL REFERENCES nguoi_dung(id) ON DELETE CASCADE,   -- Khách đặt tour
    khoi_hanh_id INT NOT NULL REFERENCES khoi_hanh_tour(id) ON DELETE CASCADE, -- Chuyến khởi hành cụ thể
    so_nguoi_lon INT DEFAULT 1 CHECK (so_nguoi_lon >= 0),
    so_tre_em INT DEFAULT 0 CHECK (so_tre_em >= 0),
    so_em_be INT DEFAULT 0 CHECK (so_em_be >= 0),
    tong_tien DECIMAL(12,2) NOT NULL CHECK (tong_tien >= 0),
    don_vi_tien_te VARCHAR(3) DEFAULT 'USD',
    trang_thai trang_thai_dat_cho DEFAULT 'cho_xac_nhan',
    phuong_thuc_thanh_toan VARCHAR(50),      -- ví dụ: "the_tin_dung", "chuyen_khoan", "paypal"
    ngay_dat TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ngay_cap_nhat TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- Hành khách trong một đặt chỗ
CREATE TABLE hanh_khach_dat_cho (
    id SERIAL PRIMARY KEY,
    dat_cho_id INT NOT NULL REFERENCES dat_cho_tour(id) ON DELETE CASCADE,
    ho_ten VARCHAR(200) NOT NULL,
    ngay_sinh DATE,
    loai_khach VARCHAR(20) CHECK (loai_khach IN ('nguoi_lon', 'tre_em', 'em_be')),
    gioi_tinh VARCHAR(10), -- Nam, Nữ, Khác
    so_ho_chieu VARCHAR(50),
    quoc_tich VARCHAR(100),
    ghi_chu TEXT
);
-- Phương thức thanh toán: thẻ, chuyển khoản, ví điện tử...
CREATE TYPE phuong_thuc_tt AS ENUM ('the_tin_dung', 'chuyen_khoan', 'paypal', 'momo', 'zalo_pay', 'khac');

-- Trạng thái thanh toán
CREATE TYPE trang_thai_tt AS ENUM ('cho_xu_ly', 'thanh_cong', 'that_bai', 'hoan_tien');

-- Bảng thanh toán
CREATE TABLE thanh_toan (
    id SERIAL PRIMARY KEY,
    dat_cho_id INT NOT NULL REFERENCES dat_cho_tour(id) ON DELETE CASCADE, -- Gắn với đặt chỗ
    so_tien DECIMAL(12,2) NOT NULL CHECK (so_tien > 0),                     -- Số tiền thanh toán
    don_vi_tien_te VARCHAR(3) DEFAULT 'VND',                               -- USD, VND...
    phuong_thuc phuong_thuc_tt NOT NULL,                                   -- phương thức
    trang_thai trang_thai_tt DEFAULT 'cho_xu_ly',
    ma_giao_dich VARCHAR(100),                                             -- Mã giao dịch từ cổng thanh toán
    ngay_thanh_toan TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ngay_cap_nhat TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ghi_chu TEXT
);
CREATE TABLE danh_gia (
    id SERIAL PRIMARY KEY,
    tour_id INT NOT NULL REFERENCES tour(id) ON DELETE CASCADE, -- Gắn với tour
    nguoi_dung_id UUID NOT NULL REFERENCES nguoi_dung(id) ON DELETE CASCADE, -- Người đánh giá
    dat_cho_id INT NOT NULL REFERENCES dat_cho_tour(id) ON DELETE CASCADE, -- Phải đặt tour trước khi đánh giá
    rating INT NOT NULL CHECK (rating BETWEEN 1 AND 5),               -- Điểm đánh giá (1–5 sao)
    tieu_de VARCHAR(255),                                             -- Tiêu đề review
    noi_dung TEXT,                                                    -- Nội dung chi tiết
    hinh_anh_url TEXT[],                                              -- Mảng ảnh minh họa (Postgres hỗ trợ array)
    dang_hoat_dong BOOLEAN DEFAULT TRUE,                              -- Soft delete
    ngay_tao TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ngay_cap_nhat TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (tour_id, nguoi_dung_id) -- Mỗi user chỉ review 1 lần cho 1 tour
);


-- ===========================================
-- INDEXES CHO PERFORMANCE
-- ===========================================

-- Indexes cho bảng nguoi_dung
CREATE INDEX idx_nguoi_dung_email ON nguoi_dung(email);
CREATE INDEX idx_nguoi_dung_vai_tro ON nguoi_dung(vai_tro);
CREATE INDEX idx_nguoi_dung_dang_hoat_dong ON nguoi_dung(dang_hoat_dong);

-- Indexes cho bảng phien_dang_nhap
CREATE INDEX idx_phien_dang_nhap_nguoi_dung_id ON phien_dang_nhap(nguoi_dung_id);
CREATE INDEX idx_phien_dang_nhap_dang_hoat_dong ON phien_dang_nhap(dang_hoat_dong);
CREATE INDEX idx_phien_dang_nhap_thoi_han ON phien_dang_nhap(thoi_han_token);

-- Indexes cho bảng tour
CREATE INDEX idx_tour_danh_muc_id ON tour(danh_muc_id);
CREATE INDEX idx_tour_trang_thai ON tour(trang_thai);
CREATE INDEX idx_tour_nha_cung_cap_id ON tour(nha_cung_cap_id);
CREATE INDEX idx_tour_dang_hoat_dong ON tour(dang_hoat_dong);
CREATE INDEX idx_tour_noi_bat ON tour(noi_bat);

-- Indexes cho bảng khoi_hanh_tour
CREATE INDEX idx_khoi_hanh_tour_tour_id ON khoi_hanh_tour(tour_id);
CREATE INDEX idx_khoi_hanh_tour_ngay_khoi_hanh ON khoi_hanh_tour(ngay_khoi_hanh);
CREATE INDEX idx_khoi_hanh_tour_trang_thai ON khoi_hanh_tour(trang_thai);

-- Indexes cho bảng dat_cho_tour
CREATE INDEX idx_dat_cho_tour_nguoi_dung_id ON dat_cho_tour(nguoi_dung_id);
CREATE INDEX idx_dat_cho_tour_khoi_hanh_id ON dat_cho_tour(khoi_hanh_id);
CREATE INDEX idx_dat_cho_tour_trang_thai ON dat_cho_tour(trang_thai);
CREATE INDEX idx_dat_cho_tour_ngay_dat ON dat_cho_tour(ngay_dat);

-- Indexes cho bảng thanh_toan
CREATE INDEX idx_thanh_toan_dat_cho_id ON thanh_toan(dat_cho_id);
CREATE INDEX idx_thanh_toan_trang_thai ON thanh_toan(trang_thai);
CREATE INDEX idx_thanh_toan_ma_giao_dich ON thanh_toan(ma_giao_dich);

-- Indexes cho bảng danh_gia
CREATE INDEX idx_danh_gia_tour_id ON danh_gia(tour_id);
CREATE INDEX idx_danh_gia_nguoi_dung_id ON danh_gia(nguoi_dung_id);
CREATE INDEX idx_danh_gia_dat_cho_id ON danh_gia(dat_cho_id);
CREATE INDEX idx_danh_gia_rating ON danh_gia(rating);
CREATE INDEX idx_danh_gia_dang_hoat_dong ON danh_gia(dang_hoat_dong);

-- Indexes cho bảng diem_den
CREATE INDEX idx_diem_den_quoc_gia ON diem_den(quoc_gia);
CREATE INDEX idx_diem_den_khu_vuc ON diem_den(khu_vuc);

-- Indexes cho bảng lich_trinh_tour
CREATE INDEX idx_lich_trinh_tour_tour_id ON lich_trinh_tour(tour_id);
CREATE INDEX idx_lich_trinh_tour_ngay_thu ON lich_trinh_tour(ngay_thu);

-- ===========================================
-- COMPOSITE INDEXES CHO PERFORMANCE TỐI ƯU
-- ===========================================

-- Composite indexes cho bảng nguoi_dung
CREATE INDEX idx_nguoi_dung_email_dang_hoat_dong ON nguoi_dung(email, dang_hoat_dong);
CREATE INDEX idx_nguoi_dung_vai_tro_dang_hoat_dong ON nguoi_dung(vai_tro, dang_hoat_dong);

-- Composite indexes cho bảng tour
CREATE INDEX idx_tour_trang_thai_dang_hoat_dong ON tour(trang_thai, dang_hoat_dong);
CREATE INDEX idx_tour_danh_muc_trang_thai ON tour(danh_muc_id, trang_thai);
CREATE INDEX idx_tour_nha_cung_cap_trang_thai ON tour(nha_cung_cap_id, trang_thai);
CREATE INDEX idx_tour_noi_bat_trang_thai ON tour(noi_bat, trang_thai) WHERE noi_bat = TRUE;

-- Composite indexes cho bảng khoi_hanh_tour
CREATE INDEX idx_khoi_hanh_tour_tour_trang_thai ON khoi_hanh_tour(tour_id, trang_thai);
CREATE INDEX idx_khoi_hanh_tour_ngay_trang_thai ON khoi_hanh_tour(ngay_khoi_hanh, trang_thai);

-- Composite indexes cho bảng dat_cho_tour
CREATE INDEX idx_dat_cho_tour_nguoi_dung_trang_thai ON dat_cho_tour(nguoi_dung_id, trang_thai);
CREATE INDEX idx_dat_cho_tour_khoi_hanh_trang_thai ON dat_cho_tour(khoi_hanh_id, trang_thai);
CREATE INDEX idx_dat_cho_tour_ngay_dat_trang_thai ON dat_cho_tour(ngay_dat, trang_thai);

-- Composite indexes cho bảng thanh_toan
CREATE INDEX idx_thanh_toan_dat_cho_trang_thai ON thanh_toan(dat_cho_id, trang_thai);
CREATE INDEX idx_thanh_toan_phuong_thuc_trang_thai ON thanh_toan(phuong_thuc, trang_thai);

-- Composite indexes cho bảng danh_gia
CREATE INDEX idx_danh_gia_tour_dang_hoat_dong ON danh_gia(tour_id, dang_hoat_dong);
CREATE INDEX idx_danh_gia_nguoi_dung_dang_hoat_dong ON danh_gia(nguoi_dung_id, dang_hoat_dong);
CREATE INDEX idx_danh_gia_rating_dang_hoat_dong ON danh_gia(rating, dang_hoat_dong);

-- ===========================================
-- PARTIAL INDEXES CHO ACTIVE RECORDS
-- ===========================================

-- Partial indexes cho active users
CREATE INDEX idx_nguoi_dung_active_email ON nguoi_dung(email) WHERE dang_hoat_dong = TRUE;
CREATE INDEX idx_nguoi_dung_active_vai_tro ON nguoi_dung(vai_tro) WHERE dang_hoat_dong = TRUE;

-- Partial indexes cho active tours
CREATE INDEX idx_tour_active_trang_thai ON tour(trang_thai) WHERE dang_hoat_dong = TRUE;
CREATE INDEX idx_tour_active_danh_muc ON tour(danh_muc_id) WHERE dang_hoat_dong = TRUE;
CREATE INDEX idx_tour_active_nha_cung_cap ON tour(nha_cung_cap_id) WHERE dang_hoat_dong = TRUE;
CREATE INDEX idx_tour_active_noi_bat ON tour(id) WHERE noi_bat = TRUE AND dang_hoat_dong = TRUE;

-- Partial indexes cho active sessions
CREATE INDEX idx_phien_dang_nhap_active ON phien_dang_nhap(nguoi_dung_id, thoi_han_token) WHERE dang_hoat_dong = TRUE;

-- Partial indexes cho active bookings
CREATE INDEX idx_dat_cho_tour_active ON dat_cho_tour(nguoi_dung_id, ngay_dat) WHERE trang_thai IN ('cho_xac_nhan', 'da_xac_nhan', 'da_thanh_toan');

-- Partial indexes cho active reviews
CREATE INDEX idx_danh_gia_active ON danh_gia(tour_id, rating) WHERE dang_hoat_dong = TRUE;

-- ===========================================
-- FOREIGN KEY INDEXES BỔ SUNG
-- ===========================================

-- Indexes cho foreign keys còn thiếu
CREATE INDEX idx_tai_khoan_oauth_nguoi_dung_id ON tai_khoan_oauth(nguoi_dung_id);
CREATE INDEX idx_tai_khoan_oauth_nha_cung_cap_id ON tai_khoan_oauth(nha_cung_cap_id);
CREATE INDEX idx_cau_hinh_nhom_tour_tour_id ON cau_hinh_nhom_tour(tour_id);
CREATE INDEX idx_giam_gia_tour_tour_id ON giam_gia_tour(tour_id);
CREATE INDEX idx_anh_tour_tour_id ON anh_tour(tour_id);
CREATE INDEX idx_tour_diem_den_tour_id ON tour_diem_den(tour_id);
CREATE INDEX idx_tour_diem_den_diem_den_id ON tour_diem_den(diem_den_id);
CREATE INDEX idx_hoat_dong_lich_trinh_lich_trinh_id ON hoat_dong_lich_trinh(lich_trinh_id);
CREATE INDEX idx_hanh_khach_dat_cho_dat_cho_id ON hanh_khach_dat_cho(dat_cho_id);

-- ===========================================
-- SPECIALIZED INDEXES CHO BUSINESS LOGIC
-- ===========================================

-- Index cho tìm kiếm tours theo giá và ngày
CREATE INDEX idx_khoi_hanh_tour_ngay_gia ON khoi_hanh_tour(ngay_khoi_hanh, gia_dac_biet) WHERE trang_thai = 'xac_nhan';

-- Index cho tìm kiếm tours theo địa điểm
CREATE INDEX idx_diem_den_ten_quoc_gia ON diem_den(ten, quoc_gia);

-- Index cho lịch trình tour theo thứ tự
CREATE INDEX idx_lich_trinh_tour_tour_ngay ON lich_trinh_tour(tour_id, ngay_thu);

-- Index cho hoạt động theo thứ tự trong lịch trình
CREATE INDEX idx_hoat_dong_lich_trinh_thu_tu ON hoat_dong_lich_trinh(lich_trinh_id, thu_tu);

-- ===========================================
-- TEXT SEARCH INDEXES
-- ===========================================

-- Full-text search indexes
CREATE INDEX idx_tour_tieu_de_gin ON tour USING gin(to_tsvector('vietnamese', tieu_de));
CREATE INDEX idx_tour_mo_ta_gin ON tour USING gin(to_tsvector('vietnamese', mo_ta));
CREATE INDEX idx_diem_den_ten_gin ON diem_den USING gin(to_tsvector('vietnamese', ten));
CREATE INDEX idx_danh_gia_noi_dung_gin ON danh_gia USING gin(to_tsvector('vietnamese', noi_dung)) WHERE dang_hoat_dong = TRUE;
