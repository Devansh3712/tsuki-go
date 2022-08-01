package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/Devansh3712/tsuki-go/database"
	"github.com/Devansh3712/tsuki-go/middleware"
	"github.com/Devansh3712/tsuki-go/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load(".env")
}

func DiscordSignUp(c *gin.Context) {
	c.Redirect(http.StatusFound, os.Getenv("DISCORD_SIGNUP_URL"))
}

func DiscordLogin(c *gin.Context) {
	c.Redirect(http.StatusFound, os.Getenv("DISCORD_LOGIN_URL"))
}

func DiscordAuth(c *gin.Context) {
	// Retrieve user access token
	api := "https://discord.com/api/v10"
	authCode := c.Query("code")
	data := url.Values{
		"client_id":     []string{os.Getenv("DISCORD_CLIENT_ID")},
		"client_secret": []string{os.Getenv("DISCORD_CLIENT_SECRET")},
		"grant_type":    []string{"authorization_code"},
		"code":          []string{authCode},
	}
	switch c.Query("login") {
	case "true":
		data.Add("redirect_uri", "https://tsukigo.herokuapp.com/auth/discord?login=true")
	default:
		data.Add("redirect_uri", "https://tsukigo.herokuapp.com/auth/discord")
	}
	response, err := http.PostForm(api+"/oauth2/token", data)
	if err != nil {
		log.Println(err)
		c.HTML(http.StatusBadRequest, "error.tmpl.html", gin.H{
			"error": "400 Bad Request",
		})
		return
	}
	defer response.Body.Close()
	var responseData map[string]interface{}
	if err := json.NewDecoder(response.Body).Decode(&responseData); err != nil {
		log.Println(err)
		c.HTML(http.StatusBadRequest, "error.tmpl.html", gin.H{
			"error": "400 Bad Request",
		})
		return
	}
	accessToken := responseData["access_token"].(string)
	// Fetch user data
	request, _ := http.NewRequest("GET", api+"/users/@me", nil)
	request.Header.Add("Authorization", "Bearer "+accessToken)
	client := &http.Client{}
	response, err = client.Do(request)
	if err != nil {
		log.Println(err)
		c.HTML(http.StatusBadRequest, "error.tmpl.html", gin.H{
			"error": "400 Bad Request",
		})
		return
	}
	defer response.Body.Close()
	var authUser models.DiscordUser
	if err := json.NewDecoder(response.Body).Decode(&authUser); err != nil {
		log.Println(err)
		c.HTML(http.StatusBadRequest, "error.tmpl.html", gin.H{
			"error": "400 Bad Request",
		})
		return
	}
	// Signup or login user
	exists := database.ReadUserByEmail(*authUser.Email)
	switch c.Query("login") {
	case "true":
		if exists == nil {
			c.HTML(http.StatusUnauthorized, "error.tmpl.html", gin.H{
				"error":   "401 Unauthorized",
				"message": "User does not exist.",
			})
			return
		}
		token, _ := middleware.CreateToken(exists.Id)
		session := sessions.Default(c)
		session.Set("Authorization", token)
		session.Save()
		c.Redirect(http.StatusFound, "/feed")
	default:
		if exists != nil {
			c.HTML(http.StatusForbidden, "error.tmpl.html", gin.H{
				"error":   "403 Forbidden",
				"message": "Account already exists with the given email.",
			})
			return
		}
		var user models.User
		// Update the username if it already exists in the database
		if result := database.ReadUserByName(authUser.Username); result != nil {
			user.Username = authUser.Username + fmt.Sprintf("%d", time.Now().Unix())
			if len(user.Username) > 32 {
				// If updated username is more than 32 characters, shorten it
				user.Username = user.Username[:32]
			}
		} else {
			user.Username = authUser.Username
		}
		user.CreatedAt = time.Now()
		user.Email = authUser.Email
		user.Verified = authUser.Verified
		user.Id = uuid.NewString()
		// Generate a random password for oauth user
		user.Password = uuid.NewString()
		user.HashPassword()
		if authUser.Avatar != nil {
			avatar := fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s", authUser.DiscordId, *authUser.Avatar)
			user.Avatar = &avatar
		}
		if res := database.CreateUser(&user); !res {
			c.HTML(http.StatusBadRequest, "error.tmpl.html", gin.H{
				"error":   "400 Bad Request",
				"message": "Unable to create account, try again later.",
			})
			return
		}
		// Add to table that identifies OAuth users
		database.CreateOAuthUser(user.Id)
		token, _ := middleware.CreateToken(user.Id)
		session := sessions.Default(c)
		session.Set("Authorization", token)
		session.Save()
		if user.Verified {
			c.Redirect(http.StatusFound, "/user/")
		} else {
			c.Redirect(http.StatusFound, "/auth/verify?signup=true")
		}
	}
}
