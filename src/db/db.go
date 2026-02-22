package db

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/kisbogdan-kolos/gallery/helper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

type User struct {
	gorm.Model

	UserName string `gorm:"unique"`
	Password string `gorm:"not null"`

	DisplayName string `gorm:"not null"`

	Admin bool `gorm:"default:false"`

	Images []Image `gorm:"foreignKey:CreatedBy"`
}

type Image struct {
	gorm.Model

	ImageID uuid.UUID `gorm:"unique;not null"`

	CreatedBy uint `gorm:"not null"`
}

func DbConnect() error {
	host := helper.EnvGet("DB_HOST", "localhost")
	user := helper.EnvGet("DB_USER", "gorm")
	pass := helper.EnvGet("DB_PASS", "gorm")
	name := helper.EnvGet("DB_NAME", "gorm")
	port := helper.EnvGet("DB_PORT", "5432")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Europe/Budapest", host, user, pass, name, port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	DB = db

	models := []any{&User{}, &Image{}}

	for _, model := range models {
		err = DB.AutoMigrate(model)
		if err != nil {
			return err
		}
	}

	return err
}
