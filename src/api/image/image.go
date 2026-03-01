package image_api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	api_common "github.com/kisbogdan-kolos/gallery/api/common_api"
	user_api "github.com/kisbogdan-kolos/gallery/api/user"
	"github.com/kisbogdan-kolos/gallery/db"
)

type ImageCreate struct {
	Name string `json:"name"`
}

type ImageReturn struct {
	ID       uint                 `json:"id"`
	Name     string               `json:"name"`
	Uploaded time.Time            `json:"uploaded"`
	Uploader *user_api.UserReturn `json:"uploader"`
	ImageID  *uuid.UUID           `json:"image"`
}

func Register(router *gin.RouterGroup) {
	router.POST("", handleCreate)
	router.GET("", handleAll)
	router.DELETE("/:id", handleDelete)
}

func handleCreate(c *gin.Context) {
	ok, claims := api_common.ValidateJWT(c)
	if !ok {
		return
	}

	var create ImageCreate
	if err := c.ShouldBindJSON(&create); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	image := db.Image{
		Name:      create.Name,
		CreatedBy: claims.UserID,
	}

	if len(image.Name) > 40 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name must not exceed 40 cahracters"})
		return
	}

	if err := db.DB.Create(&image).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": image.ID})
}

func handleAll(c *gin.Context) {
	var images []db.Image
	res := db.DB.Preload("User").Find(&images)
	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	var returns []ImageReturn
	for _, img := range images {
		returns = append(returns, ImageReturn{
			ID:       img.ID,
			Name:     img.Name,
			Uploaded: img.CreatedAt,
			Uploader: user_api.User2Return(&img.User),
			ImageID:  img.ImageID,
		})
	}

	if returns == nil {
		returns = []ImageReturn{}
	}

	c.JSON(http.StatusOK, returns)
}

func handleDelete(c *gin.Context) {
	ok, claims := api_common.ValidateJWT(c)
	if !ok {
		return
	}

	id := c.Param("id")
	var image db.Image
	if err := db.DB.First(&image, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "image not found"})
		return
	}

	if image.CreatedBy != claims.UserID && !claims.Admin {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	if err := db.DB.Delete(&image).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
