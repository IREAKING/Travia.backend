
-- name: GetReviewByTourId :many
WITH tour_reviews_data AS (
    -- 1. Lấy dữ liệu đánh giá và người dùng, chỉ tập trung vào tour cần tìm
    SELECT
        dg.tour_id,
        dg.diem_danh_gia, -- Giữ lại cột điểm đánh giá để tính toán
        json_build_object(
            'id', dg.id,
            'tieu_de', dg.tieu_de,
            'diem_danh_gia', dg.diem_danh_gia,
            'noi_dung', dg.noi_dung,
            'hinh_anh_dinh_kem', dg.hinh_anh_dinh_kem,
            'ngay_tao', dg.ngay_tao,
            'ho_ten', nd.ho_ten -- Tên người dùng
        ) AS review_detail
    FROM danh_gia dg
    LEFT JOIN nguoi_dung nd ON dg.nguoi_dung_id = nd.id
    WHERE dg.tour_id = $1 
    AND dg.dang_hoat_dong = TRUE
),
tour_reviews AS (
    -- 2. Tổng hợp dữ liệu, tính toán thống kê và điểm sao
    SELECT
        trd.tour_id,
        -- Tổng hợp mảng JSON (Giữ nguyên)
        json_agg(trd.review_detail ORDER BY (trd.review_detail->>'ngay_tao') DESC) AS thong_tin_danh_gia,
        
        -- Thống kê tổng quan
        COUNT(trd.tour_id) AS tong_so_danh_gia,
        COALESCE(AVG(trd.diem_danh_gia), 0) AS diem_trung_binh,
        
        -- THỐNG KÊ ĐIỂM SAO CHI TIẾT
        COUNT(trd.diem_danh_gia) FILTER (WHERE trd.diem_danh_gia = 5) AS so_luong_5_sao,
        COUNT(trd.diem_danh_gia) FILTER (WHERE trd.diem_danh_gia = 4) AS so_luong_4_sao,
        COUNT(trd.diem_danh_gia) FILTER (WHERE trd.diem_danh_gia = 3) AS so_luong_3_sao,
        COUNT(trd.diem_danh_gia) FILTER (WHERE trd.diem_danh_gia = 2) AS so_luong_2_sao,
        COUNT(trd.diem_danh_gia) FILTER (WHERE trd.diem_danh_gia = 1) AS so_luong_1_sao
        
    FROM tour_reviews_data trd
    GROUP BY trd.tour_id
)
-- SỬ DỤNG TRONG CÂU SELECT CUỐI CÙNG (CỦA TOÀN BỘ ENDPOINT)
SELECT 
    tr.thong_tin_danh_gia,
    tr.tong_so_danh_gia,
    tr.diem_trung_binh,
    tr.so_luong_5_sao,
    tr.so_luong_4_sao,
    tr.so_luong_3_sao,
    tr.so_luong_2_sao,
    tr.so_luong_1_sao
FROM tour_reviews tr;