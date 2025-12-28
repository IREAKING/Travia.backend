# Hướng dẫn tạo Tour với nhiều bảng liên quan

## Tổng quan vấn đề

Khi tạo tour, cần INSERT dữ liệu vào nhiều bảng:
- `tour` - Thông tin tour chính
- `anh_tour` - Ảnh của tour
- `tour_diem_den` - Điểm đến của tour
- `lich_trinh_tour` - Lịch trình theo ngày
- `hoat_dong_lich_trinh` - Hoạt động trong mỗi ngày
- `cau_hinh_nhom_tour` - Cấu hình số lượng khách
- `giam_gia_tour` - Giảm giá (nếu có)

## ❌ KHÔNG THỂ làm như thế này:

```sql
-- SAI - SQL không hỗ trợ INSERT vào nhiều bảng cùng lúc
INSERT INTO tour (...) VALUES (...)
AND INSERT INTO anh_tour (...) VALUES (...);  -- ❌ KHÔNG HỢP LỆ
```

## ✅ CÁC GIẢI PHÁP ĐÚNG

### Giải pháp 1: Transaction trong Go Code (RECOMMENDED)

Đây là cách **TỐT NHẤT** cho dự án của bạn vì:
- ✅ Tách biệt logic rõ ràng
- ✅ Dễ debug và maintain
- ✅ Linh hoạt xử lý business logic
- ✅ Tận dụng được sqlc đã có sẵn

#### Cấu trúc thực thi:

```
BEGIN TRANSACTION
  ↓
INSERT INTO tour → Lấy tour_id
  ↓
INSERT INTO anh_tour (sử dụng tour_id)
  ↓
INSERT INTO tour_diem_den (sử dụng tour_id)
  ↓
INSERT INTO lich_trinh_tour → Lấy lich_trinh_id
  ↓
INSERT INTO hoat_dong_lich_trinh (sử dụng lich_trinh_id)
  ↓
COMMIT (nếu thành công) hoặc ROLLBACK (nếu lỗi)
```

#### Ví dụ code thực tế:

```go
// db/sqlc/tour_tx.go
package db

import (
    "context"
    "fmt"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgtype"
)

// CreateTourWithDetailsParams chứa tất cả dữ liệu cần thiết
type CreateTourWithDetailsParams struct {
    // Tour info
    Tour CreateTourParams
    
    // Images
    Images []AddTourImageParams
    
    // Destinations
    Destinations []AddTourDestinationParams
    
    // Itinerary
    Itineraries []CreateItineraryWithActivitiesParams
    
    // Group config (optional)
    GroupConfig *CreateGroupConfigParams
}

type CreateItineraryWithActivitiesParams struct {
    Itinerary CreateItineraryParams
    Activities []CreateActivityParams
}

// CreateItineraryParams cho lịch trình
type CreateItineraryParams struct {
    NgayThu      int32
    TieuDe       string
    MoTa         *string
    GioBatDau    *string  // TIME format
    GioKetThuc   *string
    DiaDiem      *string
    ThongTinLuuTru *string
}

// CreateActivityParams cho hoạt động
type CreateActivityParams struct {
    Ten        string
    GioBatDau  *string
    GioKetThuc *string
    MoTa       *string
    ThuTu      *int32
}

// CreateGroupConfigParams cho cấu hình nhóm
type CreateGroupConfigParams struct {
    SoNhoNhat *int32
    SoLonNhat *int32
}

// CreateTourWithDetailsResult trả về kết quả
type CreateTourWithDetailsResult struct {
    Tour         Tour
    Images       []AnhTour
    Destinations []TourDiemDen
    Itineraries  []LichTrinhWithActivities
}

type LichTrinhWithActivities struct {
    LichTrinh  LichTrinhTour
    Activities []HoatDongLichTrinh
}

// CreateTourWithDetails tạo tour với tất cả dữ liệu liên quan trong 1 transaction
func (t *Travia) CreateTourWithDetails(
    ctx context.Context, 
    params CreateTourWithDetailsParams,
) (*CreateTourWithDetailsResult, error) {
    
    // Bắt đầu transaction
    tx, err := t.db.Begin(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to begin transaction: %w", err)
    }
    
    // Đảm bảo rollback nếu có lỗi
    defer func() {
        if err != nil {
            tx.Rollback(ctx)
        }
    }()
    
    // Tạo Queries với transaction
    qtx := t.Queries.WithTx(tx)
    
    // Kết quả
    result := &CreateTourWithDetailsResult{}
    
    // 1. Tạo tour chính
    tour, err := qtx.CreateTour(ctx, params.Tour)
    if err != nil {
        return nil, fmt.Errorf("failed to create tour: %w", err)
    }
    result.Tour = tour
    
    // 2. Thêm ảnh tour
    for _, imgParam := range params.Images {
        imgParam.TourID = tour.ID
        img, err := qtx.AddTourImage(ctx, imgParam)
        if err != nil {
            return nil, fmt.Errorf("failed to add tour image: %w", err)
        }
        result.Images = append(result.Images, img)
    }
    
    // 3. Thêm điểm đến
    for _, destParam := range params.Destinations {
        destParam.TourID = tour.ID
        err := qtx.AddTourDestination(ctx, destParam)
        if err != nil {
            return nil, fmt.Errorf("failed to add destination: %w", err)
        }
        // Có thể query lại để lấy kết quả nếu cần
    }
    
    // 4. Tạo lịch trình và hoạt động
    for _, itinParam := range params.Itineraries {
        // Tạo lịch trình
        lichTrinh, err := qtx.CreateItinerary(ctx, CreateItineraryDBParams{
            TourID:         tour.ID,
            NgayThu:        itinParam.Itinerary.NgayThu,
            TieuDe:         itinParam.Itinerary.TieuDe,
            MoTa:           itinParam.Itinerary.MoTa,
            // ... các field khác
        })
        if err != nil {
            return nil, fmt.Errorf("failed to create itinerary day %d: %w", 
                itinParam.Itinerary.NgayThu, err)
        }
        
        ltWithAct := LichTrinhWithActivities{
            LichTrinh: lichTrinh,
        }
        
        // Thêm hoạt động cho lịch trình này
        for _, actParam := range itinParam.Activities {
            activity, err := qtx.CreateActivity(ctx, CreateActivityDBParams{
                LichTrinhID: lichTrinh.ID,
                Ten:         actParam.Ten,
                GioBatDau:   actParam.GioBatDau,
                GioKetThuc:  actParam.GioKetThuc,
                MoTa:        actParam.MoTa,
                ThuTu:       actParam.ThuTu,
            })
            if err != nil {
                return nil, fmt.Errorf("failed to create activity for day %d: %w", 
                    itinParam.Itinerary.NgayThu, err)
            }
            ltWithAct.Activities = append(ltWithAct.Activities, activity)
        }
        
        result.Itineraries = append(result.Itineraries, ltWithAct)
    }
    
    // 5. Tạo cấu hình nhóm (nếu có)
    if params.GroupConfig != nil {
        _, err := qtx.CreateGroupConfig(ctx, CreateGroupConfigDBParams{
            TourID:    tour.ID,
            SoNhoNhat: params.GroupConfig.SoNhoNhat,
            SoLonNhat: params.GroupConfig.SoLonNhat,
        })
        if err != nil {
            return nil, fmt.Errorf("failed to create group config: %w", err)
        }
    }
    
    // Commit transaction
    if err = tx.Commit(ctx); err != nil {
        return nil, fmt.Errorf("failed to commit transaction: %w", err)
    }
    
    return result, nil
}
```

#### Cập nhật interface Z:

```go
// db/sqlc/travia.go
type Z interface {
    Querier
    
    // Transaction methods
    CreateTourWithDetails(ctx context.Context, params CreateTourWithDetailsParams) (*CreateTourWithDetailsResult, error)
}
```

#### Thêm method WithTx cho Queries:

```go
// db/sqlc/db.go
func (q *Queries) WithTx(tx pgx.Tx) *Queries {
    return &Queries{
        db: tx,
    }
}
```

#### Sử dụng trong handler:

```go
// api/handler/tour.go

type CreateTourRequest struct {
    // Thông tin tour
    TieuDe       string  `json:"tieu_de" binding:"required"`
    MoTa         string  `json:"mo_ta"`
    DanhMucID    int32   `json:"danh_muc_id"`
    SoNgay       int32   `json:"so_ngay" binding:"required,min=1"`
    SoDem        int32   `json:"so_dem" binding:"required,min=0"`
    GiaMoiNguoi  float64 `json:"gia_moi_nguoi" binding:"required,gt=0"`
    DonViTienTe  string  `json:"don_vi_tien_te"`
    NhaCungCapID int32   `json:"nha_cung_cap_id"`
    
    // Ảnh
    Images []struct {
        Link         string `json:"link" binding:"required"`
        MoTaAlt      string `json:"mo_ta_alt"`
        LaAnhChinh   bool   `json:"la_anh_chinh"`
        ThuTuHienThi int32  `json:"thu_tu_hien_thi"`
    } `json:"images"`
    
    // Điểm đến
    Destinations []struct {
        DiemDenID       int32  `json:"diem_den_id" binding:"required"`
        ThuTuThamQuan   int32  `json:"thu_tu_tham_quan"`
        ThoiGianLuuTru  int32  `json:"thoi_gian_luu_tru_gio"`
    } `json:"destinations"`
    
    // Lịch trình
    Itineraries []struct {
        NgayThu    int32  `json:"ngay_thu" binding:"required"`
        TieuDe     string `json:"tieu_de" binding:"required"`
        MoTa       string `json:"mo_ta"`
        GioBatDau  string `json:"gio_bat_dau"`
        GioKetThuc string `json:"gio_ket_thuc"`
        DiaDiem    string `json:"dia_diem"`
        
        // Hoạt động trong ngày
        Activities []struct {
            Ten        string `json:"ten" binding:"required"`
            GioBatDau  string `json:"gio_bat_dau"`
            GioKetThuc string `json:"gio_ket_thuc"`
            MoTa       string `json:"mo_ta"`
            ThuTu      int32  `json:"thu_tu"`
        } `json:"activities"`
    } `json:"itineraries"`
}

// CreateTour godoc
// @Summary      Tạo tour mới với đầy đủ thông tin
// @Description  Tạo tour bao gồm ảnh, điểm đến, lịch trình và hoạt động
// @Tags         tour
// @Accept       json
// @Produce      json
// @Param        request body CreateTourRequest true "Tour data"
// @Success      201 {object} map[string]interface{}
// @Failure      400 {object} map[string]interface{}
// @Security     BearerAuth
// @Router       /tour/create [post]
func (s *Server) CreateTourFull(c *gin.Context) {
    var req CreateTourRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // Get user ID from JWT
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }
    
    var userUUID pgtype.UUID
    userUUID.Scan(userID.(string))
    
    // Convert request to params
    var giaMoiNguoi pgtype.Numeric
    // TODO: Convert float64 to pgtype.Numeric properly
    
    params := db.CreateTourWithDetailsParams{
        Tour: db.CreateTourParams{
            TieuDe:       req.TieuDe,
            MoTa:         &req.MoTa,
            DanhMucID:    &req.DanhMucID,
            SoNgay:       req.SoNgay,
            SoDem:        req.SoDem,
            GiaMoiNguoi:  giaMoiNguoi,
            DonViTienTe:  &req.DonViTienTe,
            TrangThai:    stringPtr("nhap"),
            NoiBat:       boolPtr(false),
            NguoiTaoID:   userUUID,
            NhaCungCapID: &req.NhaCungCapID,
            DangHoatDong: boolPtr(true),
        },
    }
    
    // Convert images
    for _, img := range req.Images {
        params.Images = append(params.Images, db.AddTourImageParams{
            Link:           img.Link,
            MoTaAlt:        &img.MoTaAlt,
            LaAnhChinh:     &img.LaAnhChinh,
            ThuTuHienThi:   &img.ThuTuHienThi,
        })
    }
    
    // Convert destinations
    for _, dest := range req.Destinations {
        params.Destinations = append(params.Destinations, db.AddTourDestinationParams{
            DiemDenID:         dest.DiemDenID,
            ThuTuThamQuan:     &dest.ThuTuThamQuan,
            ThoiGianLuuTruGio: &dest.ThoiGianLuuTru,
        })
    }
    
    // Convert itineraries with activities
    for _, itin := range req.Itineraries {
        itinParam := db.CreateItineraryWithActivitiesParams{
            Itinerary: db.CreateItineraryParams{
                NgayThu:    itin.NgayThu,
                TieuDe:     itin.TieuDe,
                MoTa:       &itin.MoTa,
                GioBatDau:  &itin.GioBatDau,
                GioKetThuc: &itin.GioKetThuc,
                DiaDiem:    &itin.DiaDiem,
            },
        }
        
        for _, act := range itin.Activities {
            itinParam.Activities = append(itinParam.Activities, db.CreateActivityParams{
                Ten:        act.Ten,
                GioBatDau:  &act.GioBatDau,
                GioKetThuc: &act.GioKetThuc,
                MoTa:       &act.MoTa,
                ThuTu:      &act.ThuTu,
            })
        }
        
        params.Itineraries = append(params.Itineraries, itinParam)
    }
    
    // Execute transaction
    result, err := s.z.CreateTourWithDetails(c.Request.Context(), params)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Không thể tạo tour",
            "details": err.Error(),
        })
        return
    }
    
    c.JSON(http.StatusCreated, gin.H{
        "message": "Tạo tour thành công",
        "data":    result,
    })
}

// Helper functions
func stringPtr(s string) *string {
    return &s
}

func boolPtr(b bool) *bool {
    return &b
}
```

---

### Giải pháp 2: Stored Procedure trong PostgreSQL

Nếu muốn logic ở database layer:

```sql
-- db/migration/add_create_tour_procedure.sql

CREATE OR REPLACE FUNCTION create_tour_with_details(
    -- Tour params
    p_tieu_de VARCHAR(200),
    p_mo_ta TEXT,
    p_danh_muc_id INTEGER,
    p_so_ngay INTEGER,
    p_so_dem INTEGER,
    p_gia_moi_nguoi DECIMAL(10,2),
    p_don_vi_tien_te VARCHAR(3),
    p_nguoi_tao_id UUID,
    p_nha_cung_cap_id INTEGER,
    
    -- Images (JSON array)
    p_images JSONB,
    
    -- Destinations (JSON array)
    p_destinations JSONB,
    
    -- Itineraries with activities (JSON array)
    p_itineraries JSONB
)
RETURNS JSONB
LANGUAGE plpgsql
AS $$
DECLARE
    v_tour_id INTEGER;
    v_lich_trinh_id INTEGER;
    v_image JSONB;
    v_destination JSONB;
    v_itinerary JSONB;
    v_activity JSONB;
    v_result JSONB;
BEGIN
    -- 1. Insert tour
    INSERT INTO tour (
        tieu_de, mo_ta, danh_muc_id, so_ngay, so_dem,
        gia_moi_nguoi, don_vi_tien_te, trang_thai,
        nguoi_tao_id, nha_cung_cap_id, dang_hoat_dong
    ) VALUES (
        p_tieu_de, p_mo_ta, p_danh_muc_id, p_so_ngay, p_so_dem,
        p_gia_moi_nguoi, p_don_vi_tien_te, 'nhap',
        p_nguoi_tao_id, p_nha_cung_cap_id, TRUE
    )
    RETURNING id INTO v_tour_id;
    
    -- 2. Insert images
    IF p_images IS NOT NULL THEN
        FOR v_image IN SELECT * FROM jsonb_array_elements(p_images)
        LOOP
            INSERT INTO anh_tour (tour_id, link, mo_ta_alt, la_anh_chinh, thu_tu_hien_thi)
            VALUES (
                v_tour_id,
                v_image->>'link',
                v_image->>'mo_ta_alt',
                (v_image->>'la_anh_chinh')::BOOLEAN,
                (v_image->>'thu_tu_hien_thi')::INTEGER
            );
        END LOOP;
    END IF;
    
    -- 3. Insert destinations
    IF p_destinations IS NOT NULL THEN
        FOR v_destination IN SELECT * FROM jsonb_array_elements(p_destinations)
        LOOP
            INSERT INTO tour_diem_den (tour_id, diem_den_id, thu_tu_tham_quan, thoi_gian_luu_tru_gio)
            VALUES (
                v_tour_id,
                (v_destination->>'diem_den_id')::INTEGER,
                (v_destination->>'thu_tu_tham_quan')::INTEGER,
                (v_destination->>'thoi_gian_luu_tru_gio')::INTEGER
            );
        END LOOP;
    END IF;
    
    -- 4. Insert itineraries with activities
    IF p_itineraries IS NOT NULL THEN
        FOR v_itinerary IN SELECT * FROM jsonb_array_elements(p_itineraries)
        LOOP
            -- Insert itinerary
            INSERT INTO lich_trinh_tour (
                tour_id, ngay_thu, tieu_de, mo_ta, gio_bat_dau, gio_ket_thuc, dia_diem
            ) VALUES (
                v_tour_id,
                (v_itinerary->>'ngay_thu')::INTEGER,
                v_itinerary->>'tieu_de',
                v_itinerary->>'mo_ta',
                (v_itinerary->>'gio_bat_dau')::TIME,
                (v_itinerary->>'gio_ket_thuc')::TIME,
                v_itinerary->>'dia_diem'
            )
            RETURNING id INTO v_lich_trinh_id;
            
            -- Insert activities for this itinerary
            IF v_itinerary->'activities' IS NOT NULL THEN
                FOR v_activity IN SELECT * FROM jsonb_array_elements(v_itinerary->'activities')
                LOOP
                    INSERT INTO hoat_dong_lich_trinh (
                        lich_trinh_id, ten, gio_bat_dau, gio_ket_thuc, mo_ta, thu_tu
                    ) VALUES (
                        v_lich_trinh_id,
                        v_activity->>'ten',
                        (v_activity->>'gio_bat_dau')::TIME,
                        (v_activity->>'gio_ket_thuc')::TIME,
                        v_activity->>'mo_ta',
                        (v_activity->>'thu_tu')::INTEGER
                    );
                END LOOP;
            END IF;
        END LOOP;
    END IF;
    
    -- Return result
    v_result = jsonb_build_object(
        'success', TRUE,
        'tour_id', v_tour_id,
        'message', 'Tour created successfully'
    );
    
    RETURN v_result;
    
EXCEPTION
    WHEN OTHERS THEN
        -- Rollback happens automatically
        RETURN jsonb_build_object(
            'success', FALSE,
            'error', SQLERRM
        );
END;
$$;
```

Sử dụng stored procedure:

```sql
-- name: CreateTourWithDetailsProc :one
SELECT create_tour_with_details(
    $1, $2, $3, $4, $5, $6, $7, $8, $9,
    $10::jsonb, $11::jsonb, $12::jsonb
) as result;
```

---

### Giải pháp 3: CTE với RETURNING (cho case đơn giản)

Chỉ dùng khi không có nhiều logic phức tạp:

```sql
-- Ví dụ: Tạo tour + ảnh + điểm đến trong 1 query
WITH new_tour AS (
    INSERT INTO tour (tieu_de, mo_ta, so_ngay, so_dem, gia_moi_nguoi)
    VALUES ('Tour Hà Nội', 'Tham quan Hà Nội', 3, 2, 1500000)
    RETURNING id
),
new_images AS (
    INSERT INTO anh_tour (tour_id, link, la_anh_chinh)
    SELECT 
        id,
        unnest(ARRAY['img1.jpg', 'img2.jpg', 'img3.jpg']),
        unnest(ARRAY[true, false, false])
    FROM new_tour
    RETURNING *
),
new_destinations AS (
    INSERT INTO tour_diem_den (tour_id, diem_den_id, thu_tu_tham_quan)
    SELECT 
        id,
        unnest(ARRAY[1, 2, 3]::INTEGER[]),
        unnest(ARRAY[1, 2, 3]::INTEGER[])
    FROM new_tour
    RETURNING *
)
SELECT * FROM new_tour;
```

**Nhược điểm:** 
- ❌ Không linh hoạt
- ❌ Khó xử lý nested data (lịch trình → hoạt động)
- ❌ Khó debug khi có lỗi

---

## 🎯 Khuyến nghị cho dự án của bạn

**Sử dụng Giải pháp 1 (Transaction trong Go)** vì:

✅ **Ưu điểm:**
1. **Dễ maintain và debug** - Code rõ ràng, dễ theo dõi
2. **Linh hoạt** - Có thể thêm business logic, validation
3. **Tận dụng sqlc** - Sử dụng các queries đã generate
4. **Type-safe** - Go compiler check types
5. **Dễ test** - Có thể mock từng bước
6. **Rollback tự động** - Defer rollback on error

❌ **KHÔNG nên dùng:**
- Stored Procedure - Khó maintain, khó test, khó version control
- CTE - Không đủ linh hoạt cho case phức tạp như tour

---

## 📝 Các bước thực hiện

1. Thêm các SQL queries còn thiếu vào `db/query/tour.sql`
2. Tạo file `db/sqlc/tour_tx.go` với transaction logic
3. Cập nhật interface `Z` trong `db/sqlc/travia.go`
4. Thêm method `WithTx` vào `db/sqlc/db.go`
5. Tạo handler mới trong `api/handler/tour.go`
6. Thêm route trong `api/handler/router.go`
7. Test kỹ lưỡng với data thật

---

## 🔒 Tại sao phải dùng TRANSACTION?

```
Scenario không có transaction:
❌ INSERT tour → SUCCESS ✅
❌ INSERT anh_tour → SUCCESS ✅
❌ INSERT lich_trinh_tour → ERROR ❌
→ Kết quả: Tour có ảnh nhưng KHÔNG có lịch trình = DATA INCONSISTENT

Scenario có transaction:
✅ BEGIN
✅ INSERT tour → SUCCESS
✅ INSERT anh_tour → SUCCESS
✅ INSERT lich_trinh_tour → ERROR
✅ ROLLBACK → Tất cả bị hủy, database vẫn CONSISTENT
```

---

## 📚 Tài liệu tham khảo

- [PostgreSQL Transactions](https://www.postgresql.org/docs/current/tutorial-transactions.html)
- [pgx Transactions](https://github.com/jackc/pgx/wiki/Transactions)
- [sqlc Documentation](https://docs.sqlc.dev/en/stable/)

