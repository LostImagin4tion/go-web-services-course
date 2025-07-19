package services

import (
	"context"
	"stepikGoWebServices/generated/service"
)

var (
	BusinessLogicMethods = []string{
		service.BusinessLogic_Check_FullMethodName,
		service.BusinessLogic_Add_FullMethodName,
		service.BusinessLogic_Test_FullMethodName,
	}
)

type BusinessLogicService struct {
	service.BusinessLogicServer
}

func NewBusinessLogicService() *BusinessLogicService {
	return &BusinessLogicService{}
}

func (bs *BusinessLogicService) Check(
	context.Context,
	*service.Nothing,
) (*service.Nothing, error) {
	return &service.Nothing{}, nil
}

func (bs *BusinessLogicService) Add(
	context.Context,
	*service.Nothing,
) (*service.Nothing, error) {
	return &service.Nothing{}, nil
}

func (bs *BusinessLogicService) Test(
	context.Context,
	*service.Nothing,
) (*service.Nothing, error) {
	return &service.Nothing{}, nil
}
