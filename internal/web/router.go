package web

import (
	"github.com/gin-gonic/gin"
)

var (
	MapOfRoute = map[string]string{
		"firstroute": "first",
	}
)

func (guard *APIGuard) HandleRequest1(ctx *gin.Context) {
	ctx.JSON(200, "everything fine")
}

func (guard *APIGuard) RegisterRoutes(e *gin.Engine) *gin.Engine {
	//TODO Implement this function
	e.GET("/firstroute", guard.HandleRequest1)
	return e
}
