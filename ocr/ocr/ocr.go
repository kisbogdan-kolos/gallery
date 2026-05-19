package ocr

import (
	"fmt"
	"io"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/kisbogdan-kolos/gallery/backend/db"
	"github.com/kisbogdan-kolos/gallery/backend/storage"
	"github.com/kisbogdan-kolos/gallery/ocr/email"
	"github.com/minio/minio-go/v7"
	"github.com/otiai10/gosseract/v2"
)

func Run(id uint) error {
	var image db.Image
	if err := db.DB.Preload("User").First(&image, id).Error; err != nil {
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

	err = process(boxes, &image)
	if err != nil {
		return err
	}

	err = sendMail(boxes, &image)

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

func process(boxes []gosseract.BoundingBox, image *db.Image) error {
	err := db.DB.Where("image_id = ?", image.ID).Delete(&db.ImageText{}).Error
	if err != nil {
		return err
	}

	imageText := []db.ImageText{}

	for _, box := range boxes {
		if box.Confidence < 50 {
			continue
		}

		imageText = append(imageText, db.ImageText{
			ImageID: image.ID,
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

	currentTime := time.Now()
	image.OCRTime = &currentTime
	err = db.DB.Save(image).Error

	return err
}

func sendMail(boxes []gosseract.BoundingBox, image *db.Image) error {
	admins := []db.User{}

	err := db.DB.Where(&db.User{Admin: true}).Find(&admins).Error
	if err != nil {
		return err
	}

	lines := []string{}
	sort.Slice(boxes, func(i, j int) bool {
		return boxes[i].Box.Min.Y < boxes[j].Box.Min.Y
	})

	yMax := 0
	rowIndex := 0
	line := []gosseract.BoundingBox{}
	for idx, box := range boxes {
		if idx == 0 {
			yMax = box.Box.Max.Y
		}

		if box.Box.Min.Y <= yMax && idx != len(boxes)-1 {
			line = append(line, box)
		} else {
			if idx == len(boxes)-1 {
				line = append(line, box)
			}

			sort.Slice(line, func(i, j int) bool {
				return line[i].Box.Min.X < line[j].Box.Min.X
			})

			lineStr := ""
			for _, lineBox := range line {
				lineStr = lineStr + " " + lineBox.Word
			}

			lines = append(lines, strings.TrimSpace(lineStr))
			rowIndex++

			line = []gosseract.BoundingBox{}
		}

		if yMax < box.Box.Max.Y {
			yMax = box.Box.Max.Y
		}
	}

	for _, admin := range admins {
		msg := fmt.Sprintf("Hello %s!\n\nOCR run was successful for image #%v:\n\tUploaded at: %s\n\tOCR finished at: %s\n\tUploaded by: %s\n\nOCR results:\n%s\n", admin.DisplayName, image.ID, image.CreatedAt.Local(), image.OCRTime.Local(), image.User.DisplayName, strings.Join(lines, "\n"))

		err := email.Send(fmt.Sprintf("%s <%s>", admin.DisplayName, admin.Email), fmt.Sprintf("Gallery OCR run #%v", image.ID), msg)
		if err != nil {
			log.Printf("Error while sending email: %v", err)
		} else {
			log.Printf("Email sent to %v", admin.Email)
		}

	}

	return nil
}
