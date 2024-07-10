package router

import (
	"common/config"
	"common/rpc"
	"gate/api"
	"gate/auth"
	"github.com/gin-gonic/gin"
	"net/http"
)

func RegisterRouter() *gin.Engine {
	if config.Conf.Log.Level == "DEBUG" {
		gin.SetMode(gin.DebugMode)

	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	rpc.Init()
	r := gin.Default()
	r.Use(auth.Cors())
	userHandler := api.NewUserHandler()
	r.POST("/register", userHandler.Register)
	r.GET("/123", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World")
	})
	return r
}
