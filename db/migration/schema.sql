-- ===========================================
-- EXTENSIONS & CONFIGURATIONS
-- ===========================================

CREATE EXTENSION IF NOT EXISTS unaccent;
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Cấu hình tìm kiếm toàn văn cho tiếng Việt (dựa trên unaccent + simple)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_ts_config WHERE cfgname = 'vietnamese') THEN
        CREATE TEXT SEARCH CONFIGURATION vietnamese ( COPY = pg_catalog.simple );
        ALTER TEXT SEARCH CONFIGURATION vietnamese
            ALTER MAPPING FOR hword, hword_part, word WITH unaccent, simple;
    END IF;
END
$$;

-- ===========================================
-- ENUM TYPES
-- ===========================================

CREATE TYPE vai_tro_nguoi_dung AS ENUM ('khach_hang', 'quan_tri', 'nha_cung_cap');
CREATE TYPE trang_thai_khoi_hanh AS ENUM ('len_lich', 'xac_nhan', 'huy', 'hoan_thanh', 'het_cho', 'con_cho');
CREATE TYPE trang_thai_dat_cho AS ENUM ('cho_xac_nhan', 'da_xac_nhan', 'da_thanh_toan', 'da_huy', 'hoan_thanh');
CREATE TYPE loai_thanh_toan AS ENUM (
    'stripe_card',      -- Credit/Debit card qua Stripe
    'paypal',           -- PayPal
    'vnpay',            -- VNPay (local Vietnam)
    'momo',             -- MoMo wallet
    'bank_transfer',    -- Chuyển khoản ngân hàng
    'cash'              -- Tiền mặt
);
CREATE TYPE trang_thai_thanh_toan AS ENUM (
    'dang_cho_thanh_toan',  -- Đang chờ thanh toán
    'dang_xuly',            -- Đang xử lý
    'thanh_cong',           -- Hoàn thành
    'that_bai',             -- Thất bại
    'da_huy',               -- Đã hủy
    'da_hoan_tien',         -- Đã hoàn tiền
    'hoan_mot_phan'         -- Hoàn một phần
);

-- Trạng thái chi trả cho nhà cung cấp
CREATE TYPE trang_thai_chi_tra AS ENUM (
    'cho_chi_tra',          -- Chờ chi trả (tour chưa hoàn thành)
    'san_sang',             -- Sẵn sàng chi trả (tour đã hoàn thành, đủ điều kiện)
    'dang_xu_ly',           -- Đang xử lý chi trả
    'da_chi_tra',           -- Đã chi trả thành công
    'that_bai',             -- Chi trả thất bại
    'da_giu',               -- Bị giữ lại (có tranh chấp)
    'da_huy'                -- Đã hủy (booking bị hủy)
);
-- ===========================================
-- USER & AUTHENTICATION TABLES
-- ===========================================

-- Người dùng
CREATE TABLE nguoi_dung (
    id UUID PRIMARY KEY default gen_random_uuid(),
    ho_ten VARCHAR(255) not NULL,
    email VARCHAR(255) UNIQUE not null,
    mat_khau_ma_hoa TEXT not null,
    so_dien_thoai VARCHAR(50),
    vai_tro vai_tro_nguoi_dung DEFAULT 'khach_hang',
    dang_hoat_dong BOOLEAN DEFAULT TRUE,
    xac_thuc BOOLEAN default FALSE,
    ngay_tao TIMESTAMP DEFAULT NOW(),
    ngay_cap_nhat TIMESTAMP DEFAULT NOW()
);

-- Indexes cho bảng nguoi_dung
CREATE INDEX idx_nguoi_dung_email ON nguoi_dung(email);
CREATE INDEX idx_nguoi_dung_vai_tro ON nguoi_dung(vai_tro);
CREATE INDEX idx_nguoi_dung_dang_hoat_dong ON nguoi_dung(dang_hoat_dong);
CREATE INDEX idx_nguoi_dung_email_dang_hoat_dong ON nguoi_dung(email, dang_hoat_dong);
CREATE INDEX idx_nguoi_dung_vai_tro_dang_hoat_dong ON nguoi_dung(vai_tro, dang_hoat_dong);
CREATE INDEX idx_nguoi_dung_active_email ON nguoi_dung(email) WHERE dang_hoat_dong = TRUE;
CREATE INDEX idx_nguoi_dung_active_vai_tro ON nguoi_dung(vai_tro) WHERE dang_hoat_dong = TRUE;


-- Indexes cho bảng tai_khoan_oauth
CREATE INDEX idx_tai_khoan_oauth_nguoi_dung_id ON tai_khoan_oauth(nguoi_dung_id);
CREATE INDEX idx_tai_khoan_oauth_nha_cung_cap_id ON tai_khoan_oauth(nha_cung_cap_id);

-- Phiên đăng nhập
CREATE TABLE phien_dang_nhap (
    id SERIAL PRIMARY KEY,
    nguoi_dung_id UUID NOT NULL REFERENCES nguoi_dung(id) ON DELETE CASCADE,
    ma_token_truy_cap TEXT NOT NULL,
    ma_token_lam_moi TEXT NOT NULL,
    thoi_han_token TIMESTAMP NOT NULL,
    thong_tin_thiet_bi TEXT,
    dang_hoat_dong BOOLEAN DEFAULT TRUE,
    ngay_tao TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ngay_cap_nhat TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes cho bảng phien_dang_nhap
CREATE INDEX idx_phien_dang_nhap_nguoi_dung_id ON phien_dang_nhap(nguoi_dung_id);
CREATE INDEX idx_phien_dang_nhap_dang_hoat_dong ON phien_dang_nhap(dang_hoat_dong);
CREATE INDEX idx_phien_dang_nhap_thoi_han ON phien_dang_nhap(thoi_han_token);
CREATE INDEX idx_phien_dang_nhap_active ON phien_dang_nhap(nguoi_dung_id, thoi_han_token) WHERE dang_hoat_dong = TRUE;

-- ===========================================
-- DESTINATION & CATEGORY TABLES
-- ===========================================

-- Điểm đến (Destination)
CREATE TABLE diem_den (
    id SERIAL PRIMARY KEY,
    ten VARCHAR(255) NOT NULL,
    tinh VARCHAR(255),
    quoc_gia VARCHAR(100),
    khu_vuc VARCHAR(100),
    iso2 CHAR(2),
    iso3 CHAR(3),
    mo_ta TEXT,
    anh TEXT,
    vi_do DECIMAL(9,6),
    kinh_do DECIMAL(9,6),
    ngay_tao TIMESTAMP DEFAULT NOW(),
    ngay_cap_nhat TIMESTAMP DEFAULT NOW()
);

-- Indexes cho bảng diem_den
CREATE INDEX idx_diem_den_quoc_gia ON diem_den(quoc_gia);
CREATE INDEX idx_diem_den_khu_vuc ON diem_den(khu_vuc);
CREATE INDEX idx_diem_den_ten_quoc_gia ON diem_den(ten, quoc_gia);
CREATE INDEX idx_diem_den_ten_gin ON diem_den USING gin(to_tsvector('vietnamese', ten));

-- Danh mục tour
CREATE TABLE danh_muc_tour (
    id SERIAL PRIMARY KEY,
    ten VARCHAR(50) NOT NULL UNIQUE,
    mo_ta TEXT,
    anh VARCHAR(255),
    dang_hoat_dong BOOLEAN DEFAULT TRUE,
    ngay_tao TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ===========================================
-- SUPPLIER & TOUR TABLES
-- ===========================================

-- Nhà cung cấp
CREATE TABLE nha_cung_cap (
    id UUID PRIMARY KEY REFERENCES nguoi_dung(id) ON DELETE CASCADE,
    ten VARCHAR(255) NOT NULL,
    dia_chi TEXT,
    website VARCHAR(255),
    mo_ta TEXT,
    logo TEXT,
    nam_thanh_lap DATE,
    thanh_pho VARCHAR(255),
    quoc_gia VARCHAR(255),
    ma_so_thue VARCHAR(255),
    so_nhan_vien VARCHAR(255),
    giay_to_kinh_doanh TEXT
);

-- Tài khoản ngân hàng của nhà cung cấp (Đơn giản hóa)
CREATE TABLE tai_khoan_ngan_hang (
    id SERIAL PRIMARY KEY,
    nha_cung_cap_id UUID NOT NULL REFERENCES nha_cung_cap(id) ON DELETE CASCADE,
    ten_ngan_hang VARCHAR(100) NOT NULL,
    so_tai_khoan VARCHAR(50) NOT NULL,
    ten_chu_tai_khoan VARCHAR(255) NOT NULL,
    chi_nhanh VARCHAR(255),
    la_mac_dinh BOOLEAN DEFAULT FALSE
);


-- Tour
CREATE TABLE tour (
    id SERIAL PRIMARY KEY,
    tieu_de VARCHAR(200) NOT NULL,
    mo_ta TEXT,
    danh_muc_id INTEGER REFERENCES danh_muc_tour(id),
    so_ngay INTEGER NOT NULL CHECK (so_ngay > 0),
    so_dem INTEGER NOT NULL CHECK (so_dem >= 0),
    gia_nguoi_lon DECIMAL(10, 2) NOT NULL CHECK (gia_nguoi_lon > 0),
    gia_tre_em DECIMAL(10, 2) NOT NULL CHECK (gia_tre_em > 0),
    don_vi_tien_te VARCHAR(3) DEFAULT 'VND',
    trang_thai VARCHAR(20) DEFAULT 'nhap' CHECK (trang_thai IN ('nhap', 'cong_bo', 'luu_tru')),
    noi_bat BOOLEAN DEFAULT FALSE,
    nha_cung_cap_id UUID REFERENCES nha_cung_cap(id) ON DELETE CASCADE,
    dang_hoat_dong BOOLEAN DEFAULT TRUE,
    ngay_tao TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ngay_cap_nhat TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_tour_ngay_cap_nhat CHECK (ngay_cap_nhat >= ngay_tao)
);

-- Indexes cho bảng tour
CREATE INDEX idx_tour_danh_muc_id ON tour(danh_muc_id);
CREATE INDEX idx_tour_trang_thai ON tour(trang_thai);
CREATE INDEX idx_tour_nha_cung_cap_id ON tour(nha_cung_cap_id);
CREATE INDEX idx_tour_dang_hoat_dong ON tour(dang_hoat_dong);
CREATE INDEX idx_tour_noi_bat ON tour(noi_bat);
CREATE INDEX idx_tour_trang_thai_dang_hoat_dong ON tour(trang_thai, dang_hoat_dong);
CREATE INDEX idx_tour_danh_muc_trang_thai ON tour(danh_muc_id, trang_thai);
CREATE INDEX idx_tour_nha_cung_cap_trang_thai ON tour(nha_cung_cap_id, trang_thai);
CREATE INDEX idx_tour_noi_bat_trang_thai ON tour(noi_bat, trang_thai) WHERE noi_bat = TRUE;
CREATE INDEX idx_tour_active_trang_thai ON tour(trang_thai) WHERE dang_hoat_dong = TRUE;
CREATE INDEX idx_tour_active_danh_muc ON tour(danh_muc_id) WHERE dang_hoat_dong = TRUE;
CREATE INDEX idx_tour_active_nha_cung_cap ON tour(nha_cung_cap_id) WHERE dang_hoat_dong = TRUE;
CREATE INDEX idx_tour_active_noi_bat ON tour(id) WHERE noi_bat = TRUE AND dang_hoat_dong = TRUE;
CREATE INDEX idx_tour_tieu_de_gin ON tour USING gin(to_tsvector('vietnamese', tieu_de));
CREATE INDEX idx_tour_mo_ta_gin ON tour USING gin(to_tsvector('vietnamese', mo_ta));

-- Cài đặt số lượng khách
CREATE TABLE cau_hinh_nhom_tour (
    id SERIAL PRIMARY KEY,
    tour_id INT NOT NULL REFERENCES tour(id) ON DELETE CASCADE,
    so_nho_nhat INT DEFAULT 1 CHECK (so_nho_nhat > 0),
    so_lon_nhat INT DEFAULT 20 CHECK (so_lon_nhat > 0)
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

-- Indexes cho bảng giam_gia_tour
CREATE INDEX idx_giam_gia_tour_tour_id ON giam_gia_tour(tour_id);

-- Ảnh tour
CREATE TABLE anh_tour (
    id SERIAL PRIMARY KEY,
    tour_id INTEGER NOT NULL REFERENCES tour(id) ON DELETE CASCADE,
    duong_dan VARCHAR(255) NOT NULL,
    mo_ta VARCHAR(100),
    la_anh_chinh BOOLEAN DEFAULT FALSE,
    thu_tu_hien_thi INTEGER DEFAULT 0,
    ngay_tao TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Liên kết tour với điểm đến
CREATE TABLE tour_diem_den (
    tour_id INTEGER NOT NULL REFERENCES tour(id) ON DELETE CASCADE,
    diem_den_id INTEGER NOT NULL REFERENCES diem_den(id) ON DELETE CASCADE,
    thu_tu_tham_quan INTEGER DEFAULT 0,
    PRIMARY KEY (tour_id, diem_den_id)
);

-- Lịch trình tour (theo ngày)
CREATE TABLE lich_trinh (
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
CREATE TABLE hoat_dong_trong_ngay (
    id SERIAL PRIMARY KEY,
    lich_trinh_id INTEGER NOT NULL REFERENCES lich_trinh(id) ON DELETE CASCADE,
    ten VARCHAR(200) NOT NULL,
    gio_bat_dau TIME,
    gio_ket_thuc TIME,
    mo_ta TEXT,
    thu_tu INTEGER DEFAULT 0,
    ngay_tao TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE khoi_hanh_tour (
    id SERIAL PRIMARY KEY,
    tour_id INT NOT NULL REFERENCES tour(id) ON DELETE CASCADE,
    ngay_khoi_hanh DATE NOT NULL,
    ngay_ket_thuc DATE NOT NULL CHECK (ngay_ket_thuc >= ngay_khoi_hanh),
    suc_chua INT NOT NULL CHECK (suc_chua > 0),
    so_cho_da_dat INT DEFAULT 0 CHECK (so_cho_da_dat >= 0),
    trang_thai trang_thai_khoi_hanh DEFAULT 'len_lich',
    ghi_chu TEXT,
    ngay_tao TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ngay_cap_nhat TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE dat_cho (
    id SERIAL PRIMARY KEY,
    nguoi_dung_id UUID NOT NULL REFERENCES nguoi_dung(id) ON DELETE CASCADE,
    khoi_hanh_id INT NOT NULL REFERENCES khoi_hanh_tour(id) ON DELETE CASCADE,
    so_nguoi_lon INT DEFAULT 1 CHECK (so_nguoi_lon >= 0),
    so_tre_em INT DEFAULT 0 CHECK (so_tre_em >= 0),
    tong_tien DECIMAL(12,2) NOT NULL CHECK (tong_tien >= 0),
    don_vi_tien_te VARCHAR(3) DEFAULT 'VND',
    trang_thai trang_thai_dat_cho DEFAULT 'cho_xac_nhan',
    phuong_thuc_thanh_toan VARCHAR(50),
    ngay_dat TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ngay_cap_nhat TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE hanh_khach (
    id SERIAL PRIMARY KEY,
    dat_cho_id INT NOT NULL REFERENCES dat_cho(id) ON DELETE CASCADE,
    ho_ten VARCHAR(200) NOT NULL,
    ngay_sinh DATE,
    loai_khach VARCHAR(20) CHECK (loai_khach IN ('nguoi_lon', 'tre_em')),
    gioi_tinh VARCHAR(10),
    so_giay_to_tuy_thanh VARCHAR(50), -- CMND, CCCD, Hộ chiếu, ...
    quoc_tich VARCHAR(100),
    ghi_chu TEXT
);    

CREATE TABLE cong_thanh_toan (
    id VARCHAR(50) PRIMARY KEY, -- 'vnpay', 'momo', 'stripe', 'tien_mat'
    ten_hien_thi VARCHAR(100) NOT NULL,
    hoat_dong BOOLEAN DEFAULT TRUE,
    phi_giao_dich_phan_tram DECIMAL(5,2) DEFAULT 0
);
CREATE TABLE lich_su_giao_dich (
    id SERIAL PRIMARY KEY,
    dat_cho_id INTEGER REFERENCES dat_cho(id) ON DELETE SET NULL,
    nguoi_dung_id UUID REFERENCES nguoi_dung(id),
    
    -- Thông tin giao dịch
    ma_giao_dich_noi_bo VARCHAR(50) UNIQUE NOT NULL, -- Hệ thống tự sinh
    ma_tham_chieu_cong_thanh_toan VARCHAR(255), -- Transaction ID của VNPay/Momo/Stripe
    
    cong_thanh_toan_id VARCHAR(50) REFERENCES cong_thanh_toan(id),
    so_tien DECIMAL(15, 2) NOT NULL,
    loai_giao_dich VARCHAR(50) DEFAULT 'thanh_toan', -- 'thanh_toan', 'hoan_tien'
    trang_thai trang_thai_thanh_toan DEFAULT 'cho_thanh_toan',
    
    noi_dung_chuyen_khoan TEXT,
    
    ngay_tao TIMESTAMP DEFAULT NOW(),
    ngay_hoan_thanh TIMESTAMP
);

CREATE TABLE danh_gia (
    id SERIAL PRIMARY KEY,
    tour_id INT NOT NULL REFERENCES tour(id) ON DELETE CASCADE,
    nguoi_dung_id UUID NOT NULL REFERENCES nguoi_dung(id) ON DELETE CASCADE,
    dat_cho_id INT NOT NULL REFERENCES dat_cho(id) ON DELETE CASCADE,
    diem_danh_gia INT NOT NULL CHECK (diem_danh_gia BETWEEN 1 AND 5),
    tieu_de VARCHAR(255),
    noi_dung TEXT,
    hinh_anh_dinh_kem TEXT[],
    dang_hoat_dong BOOLEAN DEFAULT TRUE,
    ngay_tao TIMESTAMP DEFAULT NOW(),
    ngay_cap_nhat TIMESTAMP DEFAULT NOW()
);
CREATE TABLE thong_bao (
    id SERIAL PRIMARY KEY,
    nguoi_dung_id UUID NOT NULL REFERENCES nguoi_dung(id) ON DELETE CASCADE,
    tieu_de VARCHAR(255),
    noi_dung TEXT,
    loai VARCHAR(50), -- booking, payment, system, promotion
    lien_ket TEXT,
    da_doc BOOLEAN DEFAULT FALSE,
    ngay_tao TIMESTAMP DEFAULT NOW()
);

CREATE TABLE tour_yeu_thich (
    id SERIAL PRIMARY KEY,
    nguoi_dung_id UUID NOT NULL REFERENCES nguoi_dung(id) ON DELETE CASCADE,
    tour_id INT NOT NULL REFERENCES tour(id) ON DELETE CASCADE,
    ngay_tao TIMESTAMP DEFAULT NOW(),
    UNIQUE(nguoi_dung_id, tour_id)
);
