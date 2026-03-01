package api

import (
	"github.com/gin-gonic/gin"
	api_common "github.com/kisbogdan-kolos/gallery/api/common_api"
	image_api "github.com/kisbogdan-kolos/gallery/api/image"
	user_api "github.com/kisbogdan-kolos/gallery/api/user"
)

func Register(router *gin.RouterGroup) {
	api_common.Init()
	user_api.Register(router.Group("/user"))
	image_api.Register(router.Group("/image"))
}
