package api

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-crypt/crypt"
	"github.com/go-crypt/crypt/algorithm/argon2"
	"github.com/kisbogdan-kolos/gallery/db"
	"gorm.io/gorm"
)

var hasher *argon2.Hasher
var decoder *crypt.Decoder

type UserLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserRegister struct {
	UserLogin

	DisplayName string `json:"displayname"`
}

func registerUserEndpoints(router *gin.RouterGroup) {
	h, err := argon2.New(argon2.WithProfileRFC9106LowMemory())
	if err != nil {
		log.Fatal(err)
	}
	hasher = h

	d, err := crypt.NewDefaultDecoder()
	if err != nil {
		log.Fatal(err)
	}
	decoder = d

	router.POST("/register", handleRegister)
	router.POST("/login", handleLogin)
	router.GET("/me", handleMe)
}

func handleRegister(c *gin.Context) {
	var register UserRegister

	err := c.ShouldBindBodyWithJSON(&register)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	digest, err := hasher.Hash(register.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	user := db.User{
		UserName:    register.Username,
		DisplayName: register.DisplayName,
		Password:    digest.String(),
	}

	res := db.DB.Create(&user)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrDuplicatedKey) {
			c.JSON(http.StatusConflict, gin.H{"error": "username already used"})
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": user.ID})
}

func handleLogin(c *gin.Context) {
	var login UserLogin

	err := c.ShouldBindBodyWithJSON(&login)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user db.User

	res := db.DB.Find(&user, "user_name = ?", login.Username)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	digest, err := decoder.Decode(user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if !digest.Match(login.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}

	token, err := generateJWT(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func handleMe(c *gin.Context) {
	ok, claims := validateJWT(c)
	if !ok {
		return
	}

	var user db.User

	res := db.DB.Find(&user, claims.ID)
	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, user)
}
