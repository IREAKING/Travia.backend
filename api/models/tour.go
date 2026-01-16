package models

// CreateTourFullRequest chứa tất cả dữ liệu để tạo tour
type CreateTourFullRequest struct {
	// Thông tin tour cơ bản
	TieuDe       string  `json:"tieu_de" binding:"required"`
	MoTa         string  `json:"mo_ta"`
	DanhMucID    int32   `json:"danh_muc_id"`
	SoNgay       int32   `json:"so_ngay" binding:"required,min=1"`
	SoDem        int32   `json:"so_dem" binding:"min=0"`
	GiaNguoiLon  float64 `json:"gia_nguoi_lon" binding:"required,gt=0"`
	GiaTreEm     float64 `json:"gia_tre_em" binding:"required,gt=0"`
	DonViTienTe  string  `json:"don_vi_tien_te"`
	TrangThai    string  `json:"trang_thai"`
	NoiBat       bool    `json:"noi_bat"`
	NhaCungCapID int32   `json:"nha_cung_cap_id" binding:"required"`

	// Ảnh tour
	Images []struct {
		Link         string `json:"link" binding:"required"`
		MoTaAlt      string `json:"mo_ta_alt"`
		LaAnhChinh   bool   `json:"la_anh_chinh"`
		ThuTuHienThi int32  `json:"thu_tu_hien_thi"`
	} `json:"hinh_anh_tours"`

	// Điểm đến
	Destinations []struct {
		DiemDenID     int32 `json:"diem_den_id" binding:"required"`
		ThuTuThamQuan int32 `json:"thu_tu_tham_quan"`
	} `json:"dia_diem_tours"`

	// Lịch trình với hoạt động
	LichTrinhTours []struct {
		NgayThu        int32  `json:"ngay_thu" binding:"required"`
		TieuDe         string `json:"tieu_de" binding:"required"`
		MoTa           string `json:"mo_ta"`
		GioBatDau      string `json:"gio_bat_dau"`  // Format: "HH:MM:SS"
		GioKetThuc     string `json:"gio_ket_thuc"` // Format: "HH:MM:SS"
		DiaDiem        string `json:"dia_diem"`
		ThongTinLuuTru string `json:"thong_tin_luu_tru"`

		// Hoạt động trong ngày
		Activities []struct {
			Ten        string `json:"ten" binding:"required"`
			GioBatDau  string `json:"gio_bat_dau"`  // Format: "HH:MM:SS"
			GioKetThuc string `json:"gio_ket_thuc"` // Format: "HH:MM:SS"
			MoTa       string `json:"mo_ta"`
			ThuTu      int32  `json:"thu_tu"`
		} `json:"hoat_dong_lich_trinh_tours"`
	} `json:"lich_trinh_tours"`

	// Cấu hình nhóm (optional)
	GroupConfig *struct {
		SoNhoNhat int32 `json:"so_nho_nhat"`
		SoLonNhat int32 `json:"so_lon_nhat"`
	} `json:"cau_hinh_nhom_tours"`

	// Lịch khởi hành (optional)
	LichKhoiHanhTours []struct {
		NgayKhoiHanh string `json:"ngay_khoi_hanh" binding:"required"` // Format: "YYYY-MM-DD"
		NgayKetThuc  string `json:"ngay_ket_thuc" binding:"required"`  // Format: "YYYY-MM-DD"
		SucChua      int32  `json:"suc_chua" binding:"required,min=1"` // Capacity
		TrangThai    string `json:"trang_thai"`                        // Optional: len_lich, xac_nhan, huy, hoan_thanh
		GhiChu       string `json:"ghi_chu"`                           // Optional notes
	} `json:"lich_khoi_hanh_tours"`
}

// CreateTourWithImagesRequest chứa dữ liệu để tạo tour với URLs ảnh đã upload
type CreateTourRequest struct {
	// Thông tin tour cơ bản
	TieuDe      string  `json:"tieu_de" binding:"required"`
	MoTa        string  `json:"mo_ta"`
	DanhMucID   int32   `json:"danh_muc_id"`
	SoNgay      int32   `json:"so_ngay" binding:"required,min=1"`
	SoDem       int32   `json:"so_dem" binding:"min=0"`
	GiaNguoiLon float64 `json:"gia_nguoi_lon" binding:"required,gt=0"`
	GiaTreEm    float64 `json:"gia_tre_em" binding:"required,gt=0"`
	DonViTienTe string  `json:"don_vi_tien_te"`
	TrangThai   string  `json:"trang_thai"`
	NoiBat      bool    `json:"noi_bat"`

	// Ảnh tour với URLs đã upload
	HinhAnhTours []struct {
		Link         string `json:"link" binding:"required"`
		MoTaAlt      string `json:"mo_ta_alt"`
		LaAnhChinh   bool   `json:"la_anh_chinh"`
		ThuTuHienThi int32  `json:"thu_tu_hien_thi"`
	} `json:"hinh_anh_tours"`

	// Điểm đến
	DiaDiemTours []struct {
		DiemDenID     int32 `json:"diem_den_id" binding:"required"`
		ThuTuThamQuan int32 `json:"thu_tu_tham_quan"`
	} `json:"dia_diem_tours"`

	// Lịch trình với hoạt động
	LichTrinhTours []struct {
		NgayThu        int32  `json:"ngay_thu" binding:"required"`
		TieuDe         string `json:"tieu_de" binding:"required"`
		MoTa           string `json:"mo_ta"`
		GioBatDau      string `json:"gio_bat_dau"`  // Format: "HH:MM:SS"
		GioKetThuc     string `json:"gio_ket_thuc"` // Format: "HH:MM:SS"
		DiaDiem        string `json:"dia_diem"`
		ThongTinLuuTru string `json:"thong_tin_luu_tru"`

		// Hoạt động trong ngày
		HoatDongTours []struct {
			Ten        string `json:"ten" binding:"required"`
			GioBatDau  string `json:"gio_bat_dau"`  // Format: "HH:MM:SS"
			GioKetThuc string `json:"gio_ket_thuc"` // Format: "HH:MM:SS"
			MoTa       string `json:"mo_ta"`
			ThuTu      int32  `json:"thu_tu"`
		} `json:"hoat_dong_lich_trinh_tours"`
	} `json:"lich_trinh_tours"`

	// Cấu hình nhóm (optional)
	CauHinhNhomTours *struct {
		SoNhoNhat int32 `json:"so_nho_nhat"`
		SoLonNhat int32 `json:"so_lon_nhat"`
	} `json:"cau_hinh_nhom_tours"`

	// Lịch khởi hành (optional)
	LichKhoiHanhTours []struct {
		NgayKhoiHanh string `json:"ngay_khoi_hanh" binding:"required"` // Format: "YYYY-MM-DD"
		NgayKetThuc  string `json:"ngay_ket_thuc" binding:"required"`  // Format: "YYYY-MM-DD"
		SucChua      int32  `json:"suc_chua" binding:"required,min=1"` // Capacity
		TrangThai    string `json:"trang_thai"`                        // Optional: len_lich, xac_nhan, huy, hoan_thanh
		GhiChu       string `json:"ghi_chu"`                           // Optional notes
	} `json:"lich_khoi_hanh_tours"`
}

type FilterToursRequest struct {
	Limit     int32    `json:"limit"`
	Offset    int32    `json:"offset"`
	DanhMucID *int32   `json:"danh_muc_id"`
	GiaMin    float64  `json:"gia_min"`
	GiaMax    float64  `json:"gia_max"`
	SoNgayMin *int32   `json:"so_ngay_min"`
	SoNgayMax *int32   `json:"so_ngay_max"`
	RatingMin *float64 `json:"rating_min"`
	SortBy    *string  `json:"sort_by"`
}

type UpdateLichTrinhRequest struct {
	NgayThu        int32  `json:"ngay_thu" binding:"required"`
	TieuDe         string `json:"tieu_de" binding:"required"`
	MoTa           string `json:"mo_ta"`
	GioBatDau      string `json:"gio_bat_dau"`  // Format: "HH:MM:SS"
	GioKetThuc     string `json:"gio_ket_thuc"` // Format: "HH:MM:SS"
	DiaDiem        string `json:"dia_diem"`
	ThongTinLuuTru string `json:"thong_tin_luu_tru"`
}

type UpdateHoatDongTrongNgayRequest struct {
	Ten        string `json:"ten" binding:"required"`
	GioBatDau  string `json:"gio_bat_dau"`  // Format: "HH:MM:SS"
	GioKetThuc string `json:"gio_ket_thuc"` // Format: "HH:MM:SS"
	MoTa       string `json:"mo_ta"`
	ThuTu      int32  `json:"thu_tu"`
}

type UpdateKhoiHanhTourRequest struct {
	NgayKhoiHanh string `json:"ngay_khoi_hanh" binding:"required"` // Format: "YYYY-MM-DD"
	NgayKetThuc  string `json:"ngay_ket_thuc" binding:"required"`  // Format: "YYYY-MM-DD"
	SucChua      int32  `json:"suc_chua" binding:"required,min=1"` // Capacity
	TrangThai    string `json:"trang_thai"`                        // Optional: len_lich, xac_nhan, huy, hoan_thanh
	GhiChu       string `json:"ghi_chu"`                           // Optional notes
}
type UpdateTourRequest struct {
	TieuDe       string  `json:"tieu_de" binding:"required"`
	MoTa         string  `json:"mo_ta"`
	DanhMucID    int32   `json:"danh_muc_id"`
	SoNgay       int32   `json:"so_ngay" binding:"required,min=1"`
	SoDem        int32   `json:"so_dem" binding:"min=0"`
	GiaNguoiLon  float64 `json:"gia_nguoi_lon" binding:"required,gt=0"`
	GiaTreEm     float64 `json:"gia_tre_em" binding:"required,gt=0"`
	DonViTienTe  string  `json:"don_vi_tien_te"`
	TrangThai    string  `json:"trang_thai"`
	NoiBat       bool    `json:"noi_bat"`
	NhaCungCapID string  `json:"nha_cung_cap_id" binding:"required"`
}
type AddHinhAnhTourRequest struct {
	DuongDan string `json:"duong_dan" binding:"required"`
	MoTa string `json:"mo_ta"`
	LaAnhChinh bool `json:"la_anh_chinh"`
	ThuTuHienThi int32 `json:"thu_tu_hien_thi"`
}
type AddTourDestinationRequest struct {
	DiemDenID     int32 `json:"diem_den_id" binding:"required"`
	ThuTuThamQuan int32 `json:"thu_tu_tham_quan"`
}