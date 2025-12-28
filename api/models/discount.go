package models

type CreateDiscountTourRequest struct {
	TourID      int32   `json:"tour_id"`
	PhanTram    float64 `json:"phan_tram"`
	NgayBatDau  string  `json:"ngay_bat_dau"`
	NgayKetThuc string  `json:"ngay_ket_thuc"`
}

type UpdateDiscountTourRequest struct {
	ID          int32   `json:"id"`
	TourID      int32   `json:"tour_id"`
	PhanTram    float64 `json:"phan_tram"`
	NgayBatDau  string  `json:"ngay_bat_dau"`
	NgayKetThuc string  `json:"ngay_ket_thuc"`
}
