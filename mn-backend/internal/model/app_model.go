package model

type GetHelloInfoRequest struct {
	Name string `json:"name" binding:"required"`
}

type GetHelloInfoResponse struct {
	Name string `json:"name"`
}
