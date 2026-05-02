package request

type UpsertTripRequest struct {
	TripType          string `json:"tripType" binding:"required"`
	FromText          string `json:"fromText" binding:"required"`
	ToText            string `json:"toText" binding:"required"`
	DepartureDate     string `json:"departureDate" binding:"required"`
	DepartureTime     string `json:"departureTime" binding:"required"`
	SeatCount         int    `json:"seatCount"`
	IsPriceNegotiable bool   `json:"isPriceNegotiable"`
	ContactWechat     string `json:"contactWechat"`
	ContactPhone      string `json:"contactPhone"`
	Remark            string `json:"remark"`
}

type ListTripRequest struct {
	PageNum  int    `form:"pageNum"`
	PageSize int    `form:"pageSize"`
	TripType string `form:"tripType"`
	Status   string `form:"status"`
	Keyword  string `form:"keyword"`
}

type AdminUpdateTripRequest struct {
	Status string `json:"status" binding:"required"`
}

type AdminUpdateTripDetailRequest struct {
	TripType          string   `json:"tripType" binding:"required"`
	FromText          string   `json:"fromText" binding:"required"`
	ToText            string   `json:"toText" binding:"required"`
	DepartureDate     string   `json:"departureDate" binding:"required"`
	DepartureTime     string   `json:"departureTime" binding:"required"`
	SeatCount         int      `json:"seatCount" binding:"required"`
	PriceAmount       *float64 `json:"priceAmount"`
	IsPriceNegotiable *bool    `json:"isPriceNegotiable"`
	ContactWechat     string   `json:"contactWechat"`
	ContactPhone      string   `json:"contactPhone"`
	Remark            *string  `json:"remark"`
	Status            string   `json:"status" binding:"required"`
}

type ListUserRequest struct {
	PageNum  int    `form:"pageNum"`
	PageSize int    `form:"pageSize"`
	Keyword  string `form:"keyword"`
}
