package models

type CreateSupplierRequest struct {
	ThongTinDangNhap struct {
		HoTen       string  `json:"nguoi_dai_dien"`
		Email       string  `json:"email"`
		MatKhau     string  `json:"mat_khau"`
		SoDienThoai *string `json:"so_dien_thoai"`
	} `json:"thong_tin_dang_nhap"`
	ThongTinNhaCungCap struct {
		Ten             string  `json:"ten"`
		DiaChi          *string `json:"dia_chi"`
		Website         *string `json:"website"`
		MoTa            *string `json:"mo_ta"`
		LogoUrl         *string `json:"logo_url"`
		ThanhPho        *string `json:"thanh_pho"`
		QuocGia         *string `json:"quoc_gia"`
		MaSoThue        *string `json:"ma_so_thue"`
		SoNhanVien      *string `json:"so_nhan_vien"`
		GiayToKinhDoanh *string `json:"giay_to_kinh_doanh"`
		NamThanhLap     *string `json:"nam_thanh_lap"`
	} `json:"thong_tin_nha_cung_cap"`
}
