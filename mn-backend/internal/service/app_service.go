package service

import (
	"moonick/internal/model"
	"moonick/internal/repository"
)

type AppService struct {
	// 依赖接口，而不是具体结构体
	repo repository.AppRepository
}

// 构造函数注入依赖
func NewAppService(repo repository.AppRepository) *AppService {
	return &AppService{repo: repo}
}

func (s *AppService) GetHelloInfo(param *model.GetHelloInfoRequest) (*model.GetHelloInfoResponse, error) {
	return s.repo.GetHelloInfo(param)
}
