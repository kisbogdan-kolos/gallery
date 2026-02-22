package api

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/kisbogdan-kolos/gallery/db"
	"github.com/kisbogdan-kolos/gallery/helper"
)

type JWTClaims struct {
	jwt.RegisteredClaims
	UserID uint `json:"userid"`
	Admin  bool `json:"admin"`
}

var jwtSecret []byte
var jwtExpirationSeconds uint

func Register(router *gin.RouterGroup) {
	jwtSecret = []byte(helper.EnvGet("JWT_SECRET", "Almafa12"))
	exp, err := strconv.Atoi(helper.EnvGet("JWT_EXPIRATION", "1800"))
	if err != nil {
		log.Fatalf("Failed to parse JWT expiration: %v", err)
	}
	jwtExpirationSeconds = uint(exp)

	registerUserEndpoints(router.Group("/user"))
	registerImage(router.Group("/image"))
}

func generateJWT(user *db.User) (string, error) {
	expTime := time.Now().Add(time.Second * time.Duration(jwtExpirationSeconds))
	claims := &JWTClaims{
		UserID: user.ID,
		Admin:  user.Admin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(jwtSecret)
}

func validateJWT(c *gin.Context) (bool, *JWTClaims) {
	bearerToken := c.Request.Header.Get("Authorization")
	parts := strings.Split(bearerToken, " ")
	if len(parts) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid authorization header"})
		return false, nil
	}

	tokenRequest := parts[1]
	var claims JWTClaims

	token, err := jwt.ParseWithClaims(tokenRequest, &claims, func(_ *jwt.Token) (any, error) { return jwtSecret, nil })

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JWT"})
		return false, nil
	}

	if !token.Valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JWT"})
		return false, nil
	}

	return true, &claims
}
