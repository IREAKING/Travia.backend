-- Bảng lưu lịch sử chat với AI
CREATE TABLE IF NOT EXISTS lich_su_chat (
    id SERIAL PRIMARY KEY,
    nguoi_dung_id UUID REFERENCES nguoi_dung(id) ON DELETE CASCADE,
    ma_phien VARCHAR(100) NOT NULL, -- Session ID cho khách vãng lai
    cau_hoi TEXT NOT NULL,
    cau_tra_loi TEXT NOT NULL,
    ngay_tao TIMESTAMP DEFAULT NOW()
);

-- Indexes cho bảng lich_su_chat
CREATE INDEX idx_lich_su_chat_nguoi_dung_id ON lich_su_chat(nguoi_dung_id);
CREATE INDEX idx_lich_su_chat_session_id ON lich_su_chat(session_id);
CREATE INDEX idx_lich_su_chat_ngay_tao ON lich_su_chat(ngay_tao DESC);

