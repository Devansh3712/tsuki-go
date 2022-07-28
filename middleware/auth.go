package middleware

import (
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type JWTClaims struct {
	UserId string
	jwt.StandardClaims
}

var (
	issuer          string
	secretKey       []byte
	errInvalidToken = errors.New("invalid token")
)

func init() {
	godotenv.Load(".env")
	issuer = os.Getenv("ISSUER")
	secretKey = []byte(os.Getenv("SECRET_KEY"))
}

func CreateToken(id string) (string, error) {
	claims := JWTClaims{
		id,
		jwt.StandardClaims{
			Issuer:   issuer,
			IssuedAt: time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

func ParseToken(token string) (*JWTClaims, error) {
	parsedToken, err := jwt.ParseWithClaims(token, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := parsedToken.Claims.(*JWTClaims); ok && parsedToken.Valid {
		return claims, nil
	}
	return nil, errInvalidToken
}

func AuthMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		token := session.Get("Authorization")
		if token == nil {
			c.HTML(http.StatusUnauthorized, "error.tmpl.html", gin.H{
				"error":   "401 Unauthorized",
				"message": "User not logged in.",
			})
			c.Abort()
			return
		}
		parsedToken, err := ParseToken(token.(string))
		if err != nil {
			c.HTML(http.StatusUnauthorized, "error.tmpl.html", gin.H{
				"error":   "401 Unauthorized",
				"message": "Invalid authorization token, try logging in again.",
			})
			c.Abort()
			return
		}
		session.Set("userId", parsedToken.UserId)
		session.Save()
		c.Next()
	}
}
