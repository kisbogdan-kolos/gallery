package api

import (
	"github.com/gin-gonic/gin"
	common_api "github.com/kisbogdan-kolos/gallery/backend/api/common"
	image_api "github.com/kisbogdan-kolos/gallery/backend/api/image"
	user_api "github.com/kisbogdan-kolos/gallery/backend/api/user"
)

func Register(router *gin.RouterGroup) {
	common_api.Init()
	user_api.Register(router.Group("/user"))
	image_api.Register(router.Group("/image"))
	image_api.RegisterStorage(router.Group("/storage"))
}
