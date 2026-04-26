package response

type TripSummary struct {
	ID                int64  `json:"id"`
	UserID            int64  `json:"userId"`
	TripType          string `json:"tripType"`
	FromText          string `json:"fromText"`
	ToText            string `json:"toText"`
	DepartureDate     string `json:"departureDate"`
	DepartureTime     string `json:"departureTime"`
	SeatCount         int    `json:"seatCount"`
	IsPriceNegotiable bool   `json:"isPriceNegotiable"`
	Status            string `json:"status"`
	Favorited         bool   `json:"favorited"`
	Unavailable       bool   `json:"unavailable"`
}

type TripDetail struct {
	ID                int64  `json:"id"`
	UserID            int64  `json:"userId"`
	TripType          string `json:"tripType"`
	FromText          string `json:"fromText"`
	ToText            string `json:"toText"`
	DepartureDate     string `json:"departureDate"`
	DepartureTime     string `json:"departureTime"`
	SeatCount         int    `json:"seatCount"`
	IsPriceNegotiable bool   `json:"isPriceNegotiable"`
	ContactWechat     string `json:"contactWechat"`
	ContactPhone      string `json:"contactPhone"`
	Status            string `json:"status"`
	Favorited         bool   `json:"favorited"`
	CreatedAt         string `json:"createdAt"`
	UpdatedAt         string `json:"updatedAt"`
}

type TripListResponse struct {
	Items    []*TripSummary `json:"items"`
	Total    int            `json:"total"`
	PageNum  int            `json:"pageNum"`
	PageSize int            `json:"pageSize"`
}

type ToggleFavoriteResponse struct {
	Favorited bool `json:"favorited"`
}

type AdminDashboardSummary struct {
	TotalUsers     int `json:"totalUsers"`
	TotalTrips     int `json:"totalTrips"`
	ActiveTrips    int `json:"activeTrips"`
	ExpiredTrips   int `json:"expiredTrips"`
	TotalFavorites int `json:"totalFavorites"`
}

type AdminUserSummary struct {
	ID       int64  `json:"id"`
	Phone    string `json:"phone"`
	Nickname string `json:"nickname"`
	Status   string `json:"status"`
}

type AdminUserDetail struct {
	ID                 int64  `json:"id"`
	Phone              string `json:"phone"`
	Nickname           string `json:"nickname"`
	AvatarURL          string `json:"avatarUrl"`
	Status             string `json:"status"`
	DefaultWechat      string `json:"defaultWechat"`
	DefaultPhone       string `json:"defaultPhone"`
	PublishedTripCount int    `json:"publishedTripCount"`
	FavoriteCount      int    `json:"favoriteCount"`
}

type AdminUserListResponse struct {
	Items    []*AdminUserSummary `json:"items"`
	Total    int                 `json:"total"`
	PageNum  int                 `json:"pageNum"`
	PageSize int                 `json:"pageSize"`
}
