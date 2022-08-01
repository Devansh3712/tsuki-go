package auth

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/Devansh3712/tsuki-go/database"
	"github.com/Devansh3712/tsuki-go/internal"
	"github.com/Devansh3712/tsuki-go/middleware"
	"github.com/Devansh3712/tsuki-go/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const GITHUB_URL = "https://github.com/login/oauth"

func GitHubSignUp(c *gin.Context) {
	api, _ := url.Parse(GITHUB_URL + "/authorize")
	params := url.Values{
		"client_id":    []string{os.Getenv("GITHUB_CLIENT_ID")},
		"redirect_uri": []string{"https://tsukigo.herokuapp.com/auth/github"},
	}
	api.RawQuery = params.Encode()
	api.RawQuery += "&scope=read:user,user:email"
	c.Redirect(http.StatusFound, api.String())
}

func GitHubLogin(c *gin.Context) {
	api, _ := url.Parse(GITHUB_URL + "/authorize")
	params := url.Values{
		"client_id":    []string{os.Getenv("GITHUB_CLIENT_ID")},
		"redirect_uri": []string{"https://tsukigo.herokuapp.com/auth/github?login=true"},
	}
	api.RawQuery = params.Encode()
	api.RawQuery += "&scope=read:user,user:email"
	c.Redirect(http.StatusFound, api.String())
}

func GitHubAuth(c *gin.Context) {
	// Retrieve user access token
	authCode := c.Query("code")
	client := &http.Client{}
	data := url.Values{
		"client_id":     []string{os.Getenv("GITHUB_CLIENT_ID")},
		"client_secret": []string{os.Getenv("GITHUB_CLIENT_SECRET")},
		"code":          []string{authCode},
	}
	switch c.Query("login") {
	case "true":
		data.Add("redirect_uri", "https://tsukigo.herokuapp.com/auth/github?login=true")
	default:
		data.Add("redirect_uri", "https://tsukigo.herokuapp.com/auth/github")
	}
	request, _ := http.NewRequest(
		"POST", GITHUB_URL+"/access_token", bytes.NewBuffer([]byte(data.Encode())),
	)
	request.Header.Set("Accept", "application/json")
	response, err := client.Do(request)
	if err != nil {
		log.Println(err)
		c.HTML(http.StatusBadRequest, "error.tmpl.html", gin.H{
			"error":   "400 Bad Request",
			"message": "Unable to retrieve access token, try again later.",
		})
		return
	}
	defer response.Body.Close()
	var responseData map[string]interface{}
	if err := json.NewDecoder(response.Body).Decode(&responseData); err != nil {
		log.Println(err)
		c.HTML(http.StatusBadRequest, "error.tmpl.html", gin.H{
			"error":   "400 Bad Request",
			"message": "Unable to parse authentication response, try again later.",
		})
		return
	}
	accessToken := responseData["access_token"].(string)
	// Fetch user data
	request, _ = http.NewRequest("GET", "https://api.github.com/user", nil)
	request.Header.Add("Authorization", "Bearer "+accessToken)
	response, err = client.Do(request)
	if err != nil {
		log.Println(err)
		c.HTML(http.StatusBadRequest, "error.tmpl.html", gin.H{
			"error":   "400 Bad Request",
			"message": "Unable to make authorization request, try again later.",
		})
		return
	}
	defer response.Body.Close()
	var authUser models.GitHubUser
	if err := json.NewDecoder(response.Body).Decode(&authUser); err != nil {
		log.Println(err)
		c.HTML(http.StatusBadRequest, "error.tmpl.html", gin.H{
			"error":   "400 Bad Request",
			"message": "Unable to parse authorization response, try again later.",
		})
		return
	}
	// If email is null, make another GET request
	if authUser.Email == nil {
		request, _ = http.NewRequest("GET", "https://api.github.com/user/emails", nil)
		request.Header.Add("Authorization", "Bearer "+accessToken)
		response, err = client.Do(request)
		if err != nil {
			log.Println(err)
			c.HTML(http.StatusBadRequest, "error.tmpl.html", gin.H{
				"error":   "400 Bad Request",
				"message": "Unable to make authorization request, try again later.",
			})
			return
		}
		defer response.Body.Close()
		var emails []map[string]interface{}
		if err := json.NewDecoder(response.Body).Decode(&emails); err != nil {
			log.Println(err)
			c.HTML(http.StatusBadRequest, "error.tmpl.html", gin.H{
				"error":   "400 Bad Request",
				"message": "Unable to parse authorization response, try again later.",
			})
			return
		}
		email := emails[0]["email"].(string)
		authUser.Email = &email
		authUser.Verified = emails[0]["verified"].(bool)
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
		user.Username = authUser.Username
		// Update the username if it already exists in the database
		if result := database.ReadUserByName(user.Username); result != nil {
			user.Username += internal.RandomString(32 - len(authUser.Username))
		}
		user.CreatedAt = time.Now()
		user.Email = authUser.Email
		user.Verified = authUser.Verified
		user.Id = uuid.NewString()
		// Generate a random password for oauth user
		user.Password = uuid.NewString()
		user.HashPassword()
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
