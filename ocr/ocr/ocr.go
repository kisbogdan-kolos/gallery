package ocr

import (
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/kisbogdan-kolos/gallery/backend/db"
	"github.com/kisbogdan-kolos/gallery/backend/storage"
	"github.com/minio/minio-go/v7"
	"github.com/otiai10/gosseract/v2"
)

func Run(id uint) error {
	var image db.Image
	if err := db.DB.First(&image, id).Error; err != nil {
		return err
	}

	if image.ImageID == nil {
		return fmt.Errorf("no image uploaded for %v", image.ID)
	}

	reader, size, _, err := storage.Get(*image.ImageID)
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return fmt.Errorf("no image with ID %v", image.ImageID)
		}
		return err
	}

	log.Printf("Processing image ID: %v", image.ID)

	input := make([]byte, size)
	n, err := reader.Read(input)
	if n != int(size) {
		return fmt.Errorf("failed to read image: read %v bytes, expected %v", n, size)
	}
	if err != nil && err != io.EOF {
		return err
	}

	boxes, err := ocr(input)
	if err != nil {
		log.Print(err)
	}

	err = process(boxes, id)

	log.Printf("Processing done.")

	return err
}

func ocr(in []byte) ([]gosseract.BoundingBox, error) {
	client := gosseract.NewClient()
	defer client.Close()

	client.SetImageFromBytes(in)

	boxes, err := client.GetBoundingBoxes(gosseract.RIL_WORD)
	if err != nil {
		return nil, err
	}

	return boxes, nil
}

func process(boxes []gosseract.BoundingBox, id uint) error {
	err := db.DB.Where("image_id = ?", id).Delete(&db.ImageText{}).Error
	if err != nil {
		return err
	}

	imageText := []db.ImageText{}

	for _, box := range boxes {
		if box.Confidence < 50 {
			continue
		}

		imageText = append(imageText, db.ImageText{
			ImageID: id,
			Text:    strings.Trim(box.Word, "\n\t "),
			XMin:    box.Box.Min.X,
			YMin:    box.Box.Min.Y,
			XMax:    box.Box.Max.X,
			YMax:    box.Box.Max.Y,
		})
	}

	if len(imageText) > 0 {
		err = db.DB.Create(imageText).Error
		if err != nil {
			return err
		}
	}

	img := db.Image{}
	err = db.DB.First(&img, id).Error
	if err != nil {
		return err
	}

	currentTime := time.Now()
	img.OCRTime = &currentTime
	err = db.DB.Save(&img).Error

	return err
}
