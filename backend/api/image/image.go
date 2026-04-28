package image_api

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	common_api "github.com/kisbogdan-kolos/gallery/backend/api/common"
	user_api "github.com/kisbogdan-kolos/gallery/backend/api/user"
	"github.com/kisbogdan-kolos/gallery/backend/db"
	"github.com/kisbogdan-kolos/gallery/backend/queue"
	"github.com/kisbogdan-kolos/gallery/backend/storage"
	"github.com/minio/minio-go/v7"
)

type ImageCreate struct {
	Name string `json:"name"`
}

type ImageTextReturn struct {
	Text string `json:"text"`
	XMin int    `json:"xmin"`
	YMin int    `json:"ymin"`
	XMax int    `json:"xmax"`
	YMax int    `json:"ymax"`
}

type ImageReturn struct {
	ID       uint                 `json:"id"`
	Name     string               `json:"name"`
	Uploaded time.Time            `json:"uploaded"`
	Uploader *user_api.UserReturn `json:"uploader"`
	ImageID  *uuid.UUID           `json:"image"`
	Texts    []ImageTextReturn    `json:"text"`
	OCRTime  *time.Time           `json:"ocrtime"`
}

func Image2Return(image *db.Image) *ImageReturn {
	return &ImageReturn{
		ID:       image.ID,
		Name:     image.Name,
		Uploaded: image.CreatedAt,
		Uploader: user_api.User2Return(&image.User),
		ImageID:  image.ImageID,
		Texts:    ImageTexts2Return(image.OCRTexts),
		OCRTime:  image.OCRTime,
	}
}

func ImageTexts2Return(texts []db.ImageText) []ImageTextReturn {
	ret := make([]ImageTextReturn, len(texts))
	for i, text := range texts {
		ret[i].Text = text.Text
		ret[i].XMax = text.XMax
		ret[i].XMin = text.XMin
		ret[i].YMax = text.YMax
		ret[i].YMin = text.YMin
	}

	return ret
}

func Register(router *gin.RouterGroup) {
	router.POST("", handleCreate)
	router.GET("", handleAll)
	router.DELETE("/:id", handleDelete)
	router.POST("/:id/upload", handleUpload)
	router.POST("/:id/ocr", handleOCR)
}

func RegisterStorage(router *gin.RouterGroup) {
	router.GET("/:id", handleImageGet)
}

func handleCreate(c *gin.Context) {
	ok, claims := common_api.ValidateJWT(c)
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
		log.Printf("Failed to create image")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if err := db.DB.First(&image.User, image.CreatedBy).Error; err != nil {
		log.Printf("Failed to find user for image: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, Image2Return(&image))
}

func handleAll(c *gin.Context) {
	var images []db.Image
	res := db.DB.Preload("User").Preload("OCRTexts").Find(&images)
	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	var returns []*ImageReturn
	for _, img := range images {
		returns = append(returns, Image2Return(&img))
	}

	if returns == nil {
		returns = []*ImageReturn{}
	}

	c.JSON(http.StatusOK, returns)
}

func handleDelete(c *gin.Context) {
	ok, claims := common_api.ValidateJWT(c)
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

	if image.ImageID != nil {
		err := storage.Delete(*image.ImageID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
	}

	if err := db.DB.Delete(&image).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func handleUpload(c *gin.Context) {
	ok, claims := common_api.ValidateJWT(c)
	if !ok {
		return
	}

	id := c.Param("id")
	var image db.Image
	err := db.DB.Preload("User").Preload("OCRTexts").First(&image, id).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "image not found"})
		return
	}

	if image.CreatedBy != claims.UserID && !claims.Admin {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	if image.ImageID != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image already updated"})
		return
	}

	imageId := uuid.New()
	contentType := c.Request.Header.Get("Content-Type")
	contentLength := c.Request.ContentLength
	err = storage.Set(imageId, contentType, c.Request.Body, uint(contentLength))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	image.ImageID = &imageId
	err = db.DB.Save(&image).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	err = queue.Send(fmt.Sprint(image.ID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, Image2Return(&image))
}

func handleOCR(c *gin.Context) {
	ok, claims := common_api.ValidateJWT(c)
	if !ok {
		return
	}

	id := c.Param("id")
	var image db.Image
	err := db.DB.Preload("User").Preload("OCRTexts").First(&image, id).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "image not found"})
		return
	}

	if image.CreatedBy != claims.UserID && !claims.Admin {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	err = queue.Send(fmt.Sprint(image.ID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, Image2Return(&image))
}

func handleImageGet(c *gin.Context) {
	id := c.Param("id")
	uuid, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid uuid"})
		return
	}

	reader, size, contentType, err := storage.Get(uuid)
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			c.JSON(http.StatusNotFound, gin.H{"error": "image not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.DataFromReader(http.StatusOK, int64(size), contentType, reader, map[string]string{})
}
