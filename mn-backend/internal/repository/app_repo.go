package repository

import (
	"moonick/internal/model"
)

type AppRepository interface {
	GetHelloInfo(param *model.GetHelloInfoRequest) (*model.GetHelloInfoResponse, error)
}
