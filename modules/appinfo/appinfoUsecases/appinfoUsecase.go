package appinfoUsecases

import "github.com/NatthawutSK/ri-shop/modules/appinfo/appinfoRepositories"

type IAppinfoUsecase interface{}

type appinfoUsecase struct {
	appinfoRepository appinfoRepositories.IAppinfoRepository
}

func AppinfoUsecase(appinfoRepository appinfoRepositories.IAppinfoRepository) IAppinfoUsecase {
	return &appinfoUsecase{
		appinfoRepository: appinfoRepository,
	}
}