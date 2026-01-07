-- Migration: Thêm các bảng và chức năng cho AI gợi ý tour
-- Tạo extension pgvector nếu chưa có (cho vector embeddings)
CREATE EXTENSION IF NOT EXISTS vector;

-- Lưu vết lịch sử xem tour để biết sở thích khách hàng
CREATE TABLE lich_su_xem_tour (
    id SERIAL PRIMARY KEY,
    nguoi_dung_id UUID REFERENCES nguoi_dung(id) ON DELETE CASCADE,
    tour_id INT REFERENCES tour(id) ON DELETE CASCADE,
    thoi_gian_xem TIMESTAMP DEFAULT NOW(),
    thoi_luong_xem_giay INT, -- Thời lượng xem tính bằng giây, AI sẽ ưu tiên tour khách dừng lại xem lâu
    ip_address VARCHAR(45), -- Lưu IP để track khách vãng lai
    user_agent TEXT -- Lưu user agent để phân tích
);

-- Indexes cho bảng lich_su_xem_tour
CREATE INDEX idx_lich_su_xem_tour_nguoi_dung_id ON lich_su_xem_tour(nguoi_dung_id);
CREATE INDEX idx_lich_su_xem_tour_tour_id ON lich_su_xem_tour(tour_id);
CREATE INDEX idx_lich_su_xem_tour_thoi_gian_xem ON lich_su_xem_tour(thoi_gian_xem DESC);
CREATE INDEX idx_lich_su_xem_tour_thoi_luong_xem ON lich_su_xem_tour(thoi_luong_xem_giay DESC);

-- Lưu Vector Embedding cho Tour (dùng cho AI tìm kiếm ngữ nghĩa/Semantic Search)
CREATE TABLE tour_embeddings (
    tour_id INT PRIMARY KEY REFERENCES tour(id) ON DELETE CASCADE,
    embedding vector(1536), -- 1536 là số chiều phổ biến của OpenAI
    ngay_tao TIMESTAMP DEFAULT NOW(),
    ngay_cap_nhat TIMESTAMP DEFAULT NOW()
);

-- Index cho vector search (sử dụng HNSW index cho hiệu suất tốt)
CREATE INDEX idx_tour_embeddings_vector ON tour_embeddings USING hnsw (embedding vector_cosine_ops);

-- Bảng lưu sở thích người dùng dựa trên hành vi
CREATE TABLE so_thich_nguoi_dung (
    nguoi_dung_id UUID REFERENCES nguoi_dung(id) ON DELETE CASCADE,
    loai_so_thich VARCHAR(50) NOT NULL, -- 'danh_muc' hoặc 'diem_den'
    gia_tri_id INT NOT NULL,            -- ID của danh mục hoặc điểm đến
    diem_so DECIMAL(10, 2) DEFAULT 0,  -- Điểm số sở thích (càng cao càng thích)
    ngay_cap_nhat TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (nguoi_dung_id, loai_so_thich, gia_tri_id)
);

-- Indexes cho bảng so_thich_nguoi_dung
CREATE INDEX idx_so_thich_nguoi_dung_nguoi_dung_id ON so_thich_nguoi_dung(nguoi_dung_id);
CREATE INDEX idx_so_thich_nguoi_dung_loai_so_thich ON so_thich_nguoi_dung(loai_so_thich);
CREATE INDEX idx_so_thich_nguoi_dung_diem_so ON so_thich_nguoi_dung(diem_so DESC);

-- Function tự động cập nhật điểm sở thích khi có hành động mới
CREATE OR REPLACE FUNCTION cap_nhat_diem_so_thich()
RETURNS TRIGGER AS $$
DECLARE
    v_tour_id INT;
    v_danh_muc_id INT;
    v_diem_cong INT;
    v_diem_den_ids INT[];
BEGIN
    -- Xác định điểm cộng dựa trên hành động
    IF TG_TABLE_NAME = 'lich_su_xem_tour' THEN 
        v_diem_cong := 1; 
        v_tour_id := NEW.tour_id;
        -- Nếu xem lâu (>30 giây) thì cộng thêm điểm
        IF NEW.thoi_luong_xem_giay > 30 THEN
            v_diem_cong := 2;
        END IF;
    ELSIF TG_TABLE_NAME = 'tour_yeu_thich' THEN 
        v_diem_cong := 3; 
        v_tour_id := NEW.tour_id;
    ELSIF TG_TABLE_NAME = 'dat_cho' THEN 
        v_diem_cong := 10; 
        SELECT khoi_hanh_tour.tour_id INTO v_tour_id FROM khoi_hanh_tour WHERE id = NEW.khoi_hanh_id;
    END IF;

    -- Lấy danh mục tour
    SELECT danh_muc_id INTO v_danh_muc_id FROM tour WHERE id = v_tour_id;

    -- Cập nhật hoặc thêm mới điểm cho Danh mục
    IF v_danh_muc_id IS NOT NULL THEN
        INSERT INTO so_thich_nguoi_dung (nguoi_dung_id, loai_so_thich, gia_tri_id, diem_so)
        VALUES (NEW.nguoi_dung_id, 'danh_muc', v_danh_muc_id, v_diem_cong)
        ON CONFLICT (nguoi_dung_id, loai_so_thich, gia_tri_id)
        DO UPDATE SET diem_so = so_thich_nguoi_dung.diem_so + v_diem_cong, ngay_cap_nhat = NOW();
    END IF;

    -- Cập nhật điểm cho các điểm đến của tour
    SELECT array_agg(diem_den_id) INTO v_diem_den_ids
    FROM tour_diem_den
    WHERE tour_id = v_tour_id;

    IF v_diem_den_ids IS NOT NULL THEN
        FOR i IN 1..array_length(v_diem_den_ids, 1) LOOP
            INSERT INTO so_thich_nguoi_dung (nguoi_dung_id, loai_so_thich, gia_tri_id, diem_so)
            VALUES (NEW.nguoi_dung_id, 'diem_den', v_diem_den_ids[i], v_diem_cong)
            ON CONFLICT (nguoi_dung_id, loai_so_thich, gia_tri_id)
            DO UPDATE SET diem_so = so_thich_nguoi_dung.diem_so + v_diem_cong, ngay_cap_nhat = NOW();
        END LOOP;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger cho lich_su_xem_tour
CREATE TRIGGER trigger_cap_nhat_so_thich_xem_tour
AFTER INSERT ON lich_su_xem_tour
FOR EACH ROW
WHEN (NEW.nguoi_dung_id IS NOT NULL)
EXECUTE FUNCTION cap_nhat_diem_so_thich();

-- Trigger cho tour_yeu_thich
CREATE TRIGGER trigger_cap_nhat_so_thich_yeu_thich
AFTER INSERT ON tour_yeu_thich
FOR EACH ROW
EXECUTE FUNCTION cap_nhat_diem_so_thich();

-- Trigger cho dat_cho
CREATE TRIGGER trigger_cap_nhat_so_thich_dat_cho
AFTER INSERT ON dat_cho
FOR EACH ROW
EXECUTE FUNCTION cap_nhat_diem_so_thich();

