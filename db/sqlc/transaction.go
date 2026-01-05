package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"travia.backend/api/utils"
)

// CreateTourWithDetailsParams chứa tất cả dữ liệu cần thiết để tạo tour
type CreateTourWithDetailsParams struct {
	// Tour info - Bảng: tour
	Tour CreateTourParams

	// Images - Bảng: anh_tour
	HinhAnhTours []TourImageInput

	// Destinations - Bảng: tour_diem_den
	DiaDiemTours []TourDestinationInput

	// Itineraries with activities - Bảng: lich_trinh & hoat_dong_trong_ngay
	LichTrinhTours []ItineraryWithActivitiesInput

	// Group config (optional) - Bảng: cau_hinh_nhom_tour
	CauHinhNhomTours *GroupConfigInput

	// Departures (optional) - Bảng: khoi_hanh_tour
	LichKhoiHanhTours []DepartureInput
}

// TourImageInput cho ảnh tour
// Tương ứng với bảng: anh_tour
// JSON field từ API: hinh_anh_tours
type TourImageInput struct {
	Link         string  // Column: duong_dan
	MoTaAlt      *string // Column: mo_ta
	LaAnhChinh   *bool   // Column: la_anh_chinh
	ThuTuHienThi *int32  // Column: thu_tu_hien_thi
}

// TourDestinationInput cho điểm đến
// Tương ứng với bảng: tour_diem_den
// JSON field từ API: dia_diem_tours
type TourDestinationInput struct {
	DiemDenID     int32  // Column: diem_den_id
	ThuTuThamQuan *int32 // Column: thu_tu_tham_quan
}

// ItineraryWithActivitiesInput cho lịch trình với hoạt động
// Tương ứng với bảng: lich_trinh
// JSON field từ API: lich_trinh_tours
type ItineraryWithActivitiesInput struct {
	NgayThu        int32           // Column: ngay_thu
	TieuDe         string          // Column: tieu_de
	MoTa           *string         // Column: mo_ta
	GioBatDau      *string         // Column: gio_bat_dau - TIME format "HH:MM:SS"
	GioKetThuc     *string         // Column: gio_ket_thuc - TIME format "HH:MM:SS"
	DiaDiem        *string         // Column: dia_diem
	ThongTinLuuTru *string         // Column: thong_tin_luu_tru
	Activities     []ActivityInput // Nested: hoat_dong_trong_ngay
}

// ActivityInput cho hoạt động
// Tương ứng với bảng: hoat_dong_trong_ngay
// JSON field từ API: hoat_dong_lich_trinh_tours (nested trong lich_trinh_tours)
type ActivityInput struct {
	Ten        string  // Column: ten
	GioBatDau  *string // Column: gio_bat_dau - TIME format "HH:MM:SS"
	GioKetThuc *string // Column: gio_ket_thuc - TIME format "HH:MM:SS"
	MoTa       *string // Column: mo_ta
	ThuTu      *int32  // Column: thu_tu
}

// GroupConfigInput cho cấu hình nhóm
// Tương ứng với bảng: cau_hinh_nhom_tour
// JSON field từ API: cau_hinh_nhom_tours
type GroupConfigInput struct {
	SoNhoNhat *int32 // Column: so_nho_nhat
	SoLonNhat *int32 // Column: so_lon_nhat
}

// DepartureInput cho lịch khởi hành
// Tương ứng với bảng: khoi_hanh_tour
// JSON field từ API: lich_khoi_hanh_tours
type DepartureInput struct {
	NgayKhoiHanh string  // Column: ngay_khoi_hanh - DATE format "YYYY-MM-DD"
	NgayKetThuc  string  // Column: ngay_ket_thuc - DATE format "YYYY-MM-DD"
	SucChua      int32   // Column: suc_chua
	TrangThai    *string // Column: trang_thai - Optional: len_lich, xac_nhan, huy, hoan_thanh
	GhiChu       *string // Column: ghi_chu
}

// CreateTourWithDetailsResult trả về kết quả
type CreateTourWithDetailsResult struct {
	Tour          Tour                      // Bảng: tour
	Images        []AnhTour                 // Bảng: anh_tour
	Destinations  []int32                   // IDs từ bảng: tour_diem_den
	Itineraries   []ItineraryWithActivities // Bảng: lich_trinh & hoat_dong_trong_ngay
	GroupConfigID *int32                    // ID từ bảng: cau_hinh_nhom_tour
	Departures    []KhoiHanhTour            // Bảng: khoi_hanh_tour
}

// ItineraryWithActivities kết hợp lịch trình và hoạt động
// Kết hợp dữ liệu từ 2 bảng: lich_trinh & hoat_dong_trong_ngay
type ItineraryWithActivities struct {
	Itinerary  LichTrinh           // Bảng: lich_trinh
	Activities []HoatDongTrongNgay // Bảng: hoat_dong_trong_ngay
}

// CreateTourWithDetails tạo tour với tất cả dữ liệu liên quan trong 1 transaction
// Đảm bảo tính toàn vẹn dữ liệu: nếu có bất kỳ lỗi nào, tất cả sẽ được rollback
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
	result := &CreateTourWithDetailsResult{
		Images:       make([]AnhTour, 0),
		Destinations: make([]int32, 0),
		Itineraries:  make([]ItineraryWithActivities, 0),
		Departures:   make([]KhoiHanhTour, 0),
	}

	// ============================================================
	// BƯỚC 1: TẠO TOUR CHÍNH
	// ============================================================
	tour, err := qtx.CreateTour(ctx, params.Tour)
	if err != nil {
		return nil, fmt.Errorf("failed to create tour: %w", err)
	}
	result.Tour = tour

	// ============================================================
	// BƯỚC 2: THÊM ẢNH TOUR (Bảng: anh_tour)
	// ============================================================
	for i, imgInput := range params.HinhAnhTours {
		imgParam := AddTourImageParams{
			TourID:       tour.ID,
			DuongDan:     imgInput.Link,         // duong_dan column
			MoTa:         imgInput.MoTaAlt,      // mo_ta column
			LaAnhChinh:   imgInput.LaAnhChinh,   // la_anh_chinh column
			ThuTuHienThi: imgInput.ThuTuHienThi, // thu_tu_hien_thi column
		}

		img, err := qtx.AddTourImage(ctx, imgParam)
		if err != nil {
			return nil, fmt.Errorf("failed to add tour image #%d: %w", i+1, err)
		}
		result.Images = append(result.Images, img)
	}

	// ============================================================
	// BƯỚC 3: THÊM ĐIỂM ĐẾN
	// ============================================================
	for i, destInput := range params.DiaDiemTours {
		destParam := AddTourDestinationParams{
			TourID:        tour.ID,
			DiemDenID:     destInput.DiemDenID,
			ThuTuThamQuan: destInput.ThuTuThamQuan,
		}

		err := qtx.AddTourDestination(ctx, destParam)
		if err != nil {
			return nil, fmt.Errorf("failed to add destination #%d (ID: %d): %w",
				i+1, destInput.DiemDenID, err)
		}
		result.Destinations = append(result.Destinations, destInput.DiemDenID)
	}

	// ============================================================
	// BƯỚC 4: TẠO LỊCH TRÌNH VÀ HOẠT ĐỘNG
	// ============================================================
	for _, itinInput := range params.LichTrinhTours {
		// Parse TIME values if provided
		var gioBatDau, gioKetThuc pgtype.Time
		if itinInput.GioBatDau != nil {
			if err := gioBatDau.Scan(utils.OnlyTine(*itinInput.GioBatDau)); err != nil {
				return nil, fmt.Errorf("invalid gio_bat_dau format for day %d: %w",
					itinInput.NgayThu, err)
			}
		}
		if itinInput.GioKetThuc != nil {
			if err := gioKetThuc.Scan(utils.OnlyTine(*itinInput.GioKetThuc)); err != nil {
				return nil, fmt.Errorf("invalid gio_ket_thuc format for day %d: %w",
					itinInput.NgayThu, err)
			}
		}

		// Tạo lịch trình
		itinParam := CreateItineraryParams{
			TourID:         tour.ID,
			NgayThu:        itinInput.NgayThu,
			TieuDe:         itinInput.TieuDe,
			MoTa:           itinInput.MoTa,
			GioBatDau:      gioBatDau,
			GioKetThuc:     gioKetThuc,
			DiaDiem:        itinInput.DiaDiem,
			ThongTinLuuTru: itinInput.ThongTinLuuTru,
		}

		lichTrinh, err := qtx.CreateItinerary(ctx, itinParam)
		if err != nil {
			return nil, fmt.Errorf("failed to create itinerary for day %d: %w",
				itinInput.NgayThu, err)
		}

		// Cấu trúc lưu lịch trình + activities
		ltWithAct := ItineraryWithActivities{
			Itinerary:  lichTrinh,
			Activities: make([]HoatDongTrongNgay, 0),
		}

		// Thêm hoạt động cho lịch trình này
		for j, actInput := range itinInput.Activities {
			// Parse TIME values for activity
			var actGioBatDau, actGioKetThuc pgtype.Time
			if actInput.GioBatDau != nil {
				if err := actGioBatDau.Scan(utils.OnlyTine(*actInput.GioBatDau)); err != nil {
					return nil, fmt.Errorf("invalid gio_bat_dau format for activity #%d on day %d: %w",
						j+1, itinInput.NgayThu, err)
				}
			}
			if actInput.GioKetThuc != nil {
				if err := actGioKetThuc.Scan(utils.OnlyTine(*actInput.GioKetThuc)); err != nil {
					return nil, fmt.Errorf("invalid gio_ket_thuc format for activity #%d on day %d: %w",
						j+1, itinInput.NgayThu, err)
				}
			}

			actParam := CreateActivityParams{
				LichTrinhID: lichTrinh.ID,
				Ten:         actInput.Ten,
				GioBatDau:   actGioBatDau,
				GioKetThuc:  actGioKetThuc,
				MoTa:        actInput.MoTa,
				ThuTu:       actInput.ThuTu,
			}

			activity, err := qtx.CreateActivity(ctx, actParam)
			if err != nil {
				return nil, fmt.Errorf("failed to create activity #%d for day %d: %w",
					j+1, itinInput.NgayThu, err)
			}
			ltWithAct.Activities = append(ltWithAct.Activities, activity)
		}

		result.Itineraries = append(result.Itineraries, ltWithAct)
	}

	// ============================================================
	// BƯỚC 5: TẠO CẤU HÌNH NHÓM (NẾU CÓ)
	// ============================================================
	if params.CauHinhNhomTours != nil {
		gcParam := CreateGroupConfigParams{
			TourID:    tour.ID,
			SoNhoNhat: params.CauHinhNhomTours.SoNhoNhat,
			SoLonNhat: params.CauHinhNhomTours.SoLonNhat,
		}

		groupConfig, err := qtx.CreateGroupConfig(ctx, gcParam)
		if err != nil {
			return nil, fmt.Errorf("failed to create group config: %w", err)
		}
		result.GroupConfigID = &groupConfig.ID
	}

	// ============================================================
	// BƯỚC 6: TẠO LỊCH KHỞI HÀNH (NẾU CÓ)
	// ============================================================
	for i, depInput := range params.LichKhoiHanhTours {
		// Parse DATE values
		var ngayKhoiHanh, ngayKetThuc pgtype.Date
		if err := ngayKhoiHanh.Scan(depInput.NgayKhoiHanh); err != nil {
			return nil, fmt.Errorf("invalid ngay_khoi_hanh format for departure #%d: %w",
				i+1, err)
		}
		if err := ngayKetThuc.Scan(depInput.NgayKetThuc); err != nil {
			return nil, fmt.Errorf("invalid ngay_ket_thuc format for departure #%d: %w",
				i+1, err)
		}

		// Tạo trang_thai (default: len_lich)
		var trangThai NullTrangThaiKhoiHanh
		if depInput.TrangThai != nil && *depInput.TrangThai != "" {
			trangThai.Valid = true
			trangThai.TrangThaiKhoiHanh = TrangThaiKhoiHanh(*depInput.TrangThai)
		}

		depParam := CreateDepartureParams{
			TourID:       tour.ID,
			NgayKhoiHanh: ngayKhoiHanh,
			NgayKetThuc:  ngayKetThuc,
			SucChua:      depInput.SucChua,
			TrangThai:    trangThai,
			GhiChu:       depInput.GhiChu,
		}

		departure, err := qtx.CreateDeparture(ctx, depParam)
		if err != nil {
			return nil, fmt.Errorf("failed to create departure #%d: %w", i+1, err)
		}
		result.Departures = append(result.Departures, departure)
	}

	// ============================================================
	// COMMIT TRANSACTION
	// ============================================================
	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result, nil
}

// UpdateTourWithDetails cập nhật tour và các dữ liệu liên quan
// Note: Đây là một implementation đơn giản, có thể cần logic phức tạp hơn
// để xử lý việc thêm/sửa/xóa images, destinations, itineraries
func (t *Travia) UpdateTourWithDetails(
	ctx context.Context,
	tourID int32,
	params CreateTourWithDetailsParams,
) (*CreateTourWithDetailsResult, error) {

	tx, err := t.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	_ = t.Queries.WithTx(tx) // qtx will be used when implemented

	// TODO: Implement update logic
	// 1. Update tour basic info
	// 2. Delete old images, add new ones (or implement smart diff)
	// 3. Delete old destinations, add new ones
	// 4. Delete old itineraries & activities, add new ones
	// 5. Update group config

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil, fmt.Errorf("UpdateTourWithDetails not yet implemented")
}


type CreateSupplierWithUserParams struct {
	CreateUserParams
	CreateSupplierParams
}
type CreateSupplierWithUserResult struct {
	NguoiDung  NguoiDung
	NhaCungCap NhaCungCap
}

func (t *Travia) CreateSupplierWithUser(ctx context.Context, req CreateSupplierWithUserParams) (*CreateSupplierWithUserResult, error) {
	tx, err := t.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()
	qtx := t.Queries.WithTx(tx)

	user, err := qtx.CreateUser(ctx, req.CreateUserParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	// Tạo nhà cung cấp
	supplierParams := req.CreateSupplierParams
	supplierParams.ID = user.ID
	fmt.Println("supplierParams", supplierParams.ID)
	// Tạo nhà cung cấp
	supplier, err := qtx.CreateSupplier(ctx, supplierParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create supplier: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return &CreateSupplierWithUserResult{
		NguoiDung:  user,
		NhaCungCap: supplier,
	}, nil
}

type UpdateSupplierWithUserParams struct {
	UpdateSupplierParams
	UpdateUserParams
}
type UpdateSupplierWithUserResult struct {
	NguoiDung  NguoiDung
	NhaCungCap NhaCungCap
}

func (t *Travia) UpdateSupplierWithUser(ctx context.Context, req UpdateSupplierWithUserParams) (*UpdateSupplierWithUserResult, error) {
	tx, err := t.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()
	qtx := t.Queries.WithTx(tx)
	supplier, err := qtx.UpdateSupplier(ctx, req.UpdateSupplierParams)
	if err != nil {
		return nil, fmt.Errorf("failed to update supplier: %w", err)
	}
	req.UpdateUserParams.ID = supplier.ID
	user, err := qtx.UpdateUser(ctx, req.UpdateUserParams)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return &UpdateSupplierWithUserResult{
		NguoiDung:  user,
		NhaCungCap: supplier,
	}, nil

}