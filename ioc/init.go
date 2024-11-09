package ioc

import (
	"github.com/gin-gonic/gin"
	"github.com/wureny/FluxGo/internal/service"
	"github.com/wureny/FluxGo/internal/web"
)

func Init() (*gin.Engine, error) {
	serv := initService()
	guard := initAPIGuard(serv)
	e := initWebServer(guard)
	return e, nil
}

func initService() *service.Serv {
	return &service.Serv{}
}

func initAPIGuard(serv *service.Serv) *web.APIGuard {
	guard := &web.APIGuard{
		Serv: nil,
	}
	//TODO: implement the params
	return guard
}

func initWebServer(guard *web.APIGuard) *gin.Engine {
	//TODO: Implement this function
	e := gin.New()
	e = guard.RegisterRoutes(e)
	return e
}
