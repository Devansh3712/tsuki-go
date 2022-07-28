package routes

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Devansh3712/tsuki/database"
	"github.com/Devansh3712/tsuki/middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jordan-wright/email"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

const verificationMail = `<html>
<head></head>
<body
	style="
	font-family: 'Courier New', Courier, monospace;
	padding-left: 15px;
	padding-top: 10px;
	"
>
	<h1>Tsuki</h1>
	<h2>Account Verification</h2>
	<p>
	Hi %s, please confirm that %s is your e-mail address by clicking this link %s within 48 hours.
	</p>
</body>
</html>`

func createVerificationToken(id string) (string, error) {
	claims := middleware.JWTClaims{
		UserId: id,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 2).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    issuer,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

func oauth2Client() (*http.Client, error) {
	credentials := oauth2.Config{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		Scopes:       []string{"https://mail.google.com/"},
		Endpoint: oauth2.Endpoint{
			TokenURL: os.Getenv("TOKEN_URI"),
		},
	}
	expiry, _ := time.Parse(time.RFC3339, os.Getenv("EXPIRY"))
	token := oauth2.Token{
		AccessToken:  os.Getenv("TOKEN"),
		RefreshToken: os.Getenv("REFRESH_TOKEN"),
		Expiry:       expiry,
	}
	ctx := context.Background()
	if !token.Valid() {
		// Refresh the token
		refreshedToken, err := credentials.TokenSource(ctx, &token).Token()
		if err != nil {
			log.Println(err)
			return nil, err
		}
		os.Setenv("TOKEN", refreshedToken.AccessToken)
		os.Setenv("REFRESH_TOKEN", refreshedToken.RefreshToken)
		os.Setenv("EXPIRY", refreshedToken.Expiry.String())
		return credentials.Client(ctx, refreshedToken), nil
	}
	return credentials.Client(ctx, &token), nil
}

func SendVerificationMail(c *gin.Context) {
	session := sessions.Default(c)
	id := session.Get("userId")
	if id == nil {
		c.HTML(http.StatusUnauthorized, "error.tmpl.html", gin.H{
			"error":   "401 Unauthorized",
			"message": "User not logged in.",
		})
		return
	}
	client, err := oauth2Client()
	if err != nil {
		log.Println(err)
		c.HTML(http.StatusBadRequest, "error.tmpl.html", gin.H{
			"error":   "400 Bad Request",
			"message": "Unable to send verification mail, try again later.",
		})
		return
	}
	service, err := gmail.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		log.Println(err)
		c.HTML(http.StatusBadRequest, "error.tmpl.html", gin.H{
			"error":   "400 Bad Request",
			"message": "Unable to send verification mail, try again later.",
		})
		return
	}

	user := database.ReadUserById(id.(string))
	verificationToken, _ := createVerificationToken(user.Id)
	verificationId := uuid.NewString()
	database.CreateVerificationId(verificationToken, verificationId)
	message := &email.Email{
		To:      []string{*user.Email},
		From:    os.Getenv("EMAIL"),
		Subject: "Verify your Tsuki account",
		HTML: []byte(fmt.Sprintf(
			verificationMail,
			user.Username,
			*user.Email,
			fmt.Sprintf("%s/auth/verify/%s", c.Request.Host, verificationId),
		)),
	}
	byteMessage, _ := message.Bytes()
	mailContent := gmail.Message{
		Raw: base64.RawURLEncoding.EncodeToString(byteMessage),
	}
	if _, err := service.Users.Messages.Send("me", &mailContent).Do(); err != nil {
		log.Println(err)
		c.HTML(http.StatusBadRequest, "error.tmpl.html", gin.H{
			"error":   "400 Bad Request",
			"message": "Unable to send verification mail, try again later.",
		})
		return
	}
	response := fmt.Sprintf("Verification mail sent to %s", *user.Email)
	// Check if the request is redirected from signup
	if c.Query("signup") == "true" {
		response = "Account created succesfully. " + response
	}
	c.HTML(http.StatusOK, "response.tmpl.html", gin.H{
		"message": response,
	})
}

func Verify(c *gin.Context) {
	verificationId := c.Param("id")
	verificationToken := database.ReadVerificationId(verificationId)
	if verificationToken == "" {
		c.HTML(http.StatusNotFound, "error.tmpl.html", gin.H{
			"error":   "404 Not Found",
			"message": "Verification token not found in database.",
		})
	}
	parsedToken, err := middleware.ParseToken(verificationToken)
	if err != nil {
		log.Println(err)
		c.HTML(http.StatusUnauthorized, "error.tmpl.html", gin.H{
			"error":   "401 Unauthorized",
			"message": "Invalid verification token, request a verification mail again.",
		})
		return
	}
	userId := parsedToken.UserId
	user := database.ReadUserById(userId)
	if user.Verified {
		c.HTML(http.StatusOK, "response.tmpl.html", gin.H{
			"message": "Account already verified.",
		})
		return
	}
	if result := database.UpdateUser(userId, map[string]any{"verified": true}); !result {
		log.Println(err)
		c.HTML(http.StatusBadRequest, "error.tmpl.html", gin.H{
			"error":   "400 Bad Request",
			"message": "Unable to verify account, try again later.",
		})
		return
	}
	c.HTML(http.StatusOK, "response.tmpl.html", gin.H{
		"message": "Account verified successfully.",
	})
}
