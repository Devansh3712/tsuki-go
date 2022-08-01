package routes

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/Devansh3712/tsuki-go/database"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

var postLimit = 5

func GetUser(c *gin.Context) {
	session := sessions.Default(c)
	id := session.Get("userId")
	if id == nil {
		c.HTML(http.StatusUnauthorized, "error.tmpl.html", gin.H{
			"error":   "401 Unauthorized",
			"message": "User not logged in.",
		})
		return
	}
	userId := id.(string)
	c.HTML(http.StatusOK, "user.tmpl.html", gin.H{
		"settings":  true,
		"user":      database.ReadUserById(userId),
		"postCount": database.ReadPostsCount(userId),
		"followers": database.ReadFollowers(userId),
		"following": database.ReadFollowing(userId),
		"posts":     database.ReadPosts(userId, 5, 0),
		"oauth":     database.IsOAuthUser(userId),
	})
}

func GetUserByName(c *gin.Context) {
	username := c.Param("username")
	session := sessions.Default(c)
	id := session.Get("userId")
	if id != nil {
		user := database.ReadUserById(id.(string))
		if username == user.Username {
			c.Redirect(http.StatusFound, "/user/")
			return
		}
	}
	user := database.ReadUserByName(username)
	if user == nil {
		c.HTML(http.StatusNotFound, "error.tmpl.html", gin.H{
			"error":   "404 Not Found",
			"message": "User not found",
		})
		return
	}
	user.Email = nil
	followers := database.ReadFollowers(user.Id)
	following := database.ReadFollowing(user.Id)
	postCount := database.ReadPostsCount(user.Id)
	posts := database.ReadPosts(user.Id, 5, 0)

	if id != nil {
		c.HTML(http.StatusOK, "user.tmpl.html", gin.H{
			"user":      user,
			"postCount": postCount,
			"followers": followers,
			"following": following,
			"posts":     posts,
			"follows":   database.Followed(id.(string), user.Id),
		})
		return
	}
	c.HTML(http.StatusOK, "user.tmpl.html", gin.H{
		"user":      user,
		"postCount": postCount,
		"followers": followers,
		"following": following,
		"posts":     posts,
	})
}

func GetUserPosts(c *gin.Context) {
	username := c.Param("username")
	user := database.ReadUserByName(username)
	if user == nil {
		c.HTML(http.StatusNotFound, "error.tmpl.html", gin.H{
			"error":   "404 Not Found",
			"message": "User not found",
		})
		return
	}
	postLimit = 10
	posts := database.ReadPosts(user.Id, 10, 0)
	c.HTML(http.StatusOK, "userPosts.tmpl.html", gin.H{
		"user":  user,
		"posts": posts,
	})
}

// Return posts for loading through AJAX
func LoadMorePosts(c *gin.Context) {
	username := c.Param("username")
	user := database.ReadUserByName(username)
	posts := database.ReadPosts(user.Id, 10, postLimit)
	postLimit += 10
	c.JSON(http.StatusOK, posts)
}

func UpdateAvatar(c *gin.Context) {
	session := sessions.Default(c)
	id := session.Get("userId")
	if id == nil {
		c.HTML(http.StatusUnauthorized, "error.tmpl.html", gin.H{
			"error":   "401 Unauthorized",
			"message": "User not logged in.",
		})
		return
	}
	switch c.Request.Method {
	case "GET":
		c.HTML(http.StatusOK, "update.tmpl.html", gin.H{
			"type": "avatar",
		})
	case "POST":
		// Read the image
		file, _, err := c.Request.FormFile("avatar")
		if err != nil {
			c.HTML(http.StatusBadRequest, "error.tmpl.html", gin.H{
				"error":   "400 Bad Request",
				"message": "Unable to process request, try again later.",
			})
			return
		}
		defer file.Close()
		fileData, err := ioutil.ReadAll(file)
		if err != nil {
			c.HTML(http.StatusBadRequest, "error.tmpl.html", gin.H{
				"error":   "400 Bad Request",
				"message": "Unable to read image, try again later.",
			})
			return
		}
		// Post the image to Freeimage API
		encoded := base64.StdEncoding.EncodeToString(fileData)
		response, err := http.PostForm(
			"https://freeimage.host/api/1/upload?key="+os.Getenv("FREEIMAGE_API_KEY")+"&format=json",
			url.Values{"source": {encoded}},
		)
		if err != nil {
			c.HTML(http.StatusBadRequest, "error.tmpl.html", gin.H{
				"error":   "400 Bad Request",
				"message": "Unable to read image, try again later.",
			})
			return
		}
		defer response.Body.Close()
		var responseData map[string]interface{}
		if err := json.NewDecoder(response.Body).Decode(&responseData); err != nil {
			c.HTML(http.StatusBadRequest, "error.tmpl.html", gin.H{
				"error":   "400 Bad Request",
				"message": "Unable to upload avatar, try again later.",
			})
			return
		}
		// Update user avatar URL
		if result := database.UpdateUser(
			id.(string),
			map[string]any{"avatar": responseData["image"].(map[string]interface{})["url"]},
		); !result {
			log.Println(responseData)
			c.HTML(http.StatusBadRequest, "error.tmpl.html", gin.H{
				"error":   "400 Bad Request",
				"message": "Unable to update avatar, try again later.",
			})
			return
		}
		c.HTML(http.StatusOK, "response.tmpl.html", gin.H{
			"message": "Avatar updated successfully.",
		})
	}
}

func UpdateUsername(c *gin.Context) {
	session := sessions.Default(c)
	id := session.Get("userId")
	if id == nil {
		c.HTML(http.StatusUnauthorized, "error.tmpl.html", gin.H{
			"error":   "401 Unauthorized",
			"message": "User not logged in.",
		})
		return
	}
	switch c.Request.Method {
	case "GET":
		c.HTML(http.StatusOK, "update.tmpl.html", gin.H{
			"type": "username",
		})
	case "POST":
		newUsername := c.PostForm("username")
		user := database.ReadUserById(id.(string))
		if user.Username == newUsername {
			c.HTML(http.StatusForbidden, "error.tmpl.html", gin.H{
				"error":   "403 Forbidden",
				"message": "New username cannot be the same as current.",
			})
			return
		}
		if exists := database.ReadUserByName(newUsername); exists != nil {
			c.HTML(http.StatusForbidden, "error.tmpl.html", gin.H{
				"error":   "403 Forbidden",
				"message": "Username not available or already taken.",
			})
			return
		}
		if result := database.UpdateUser(user.Id, map[string]any{"username": newUsername}); !result {
			c.HTML(http.StatusInternalServerError, "error.tmpl.html", gin.H{
				"error":   "500 Internal Server Error",
				"message": "Unable to change username, try again later.",
			})
			return
		}
		c.HTML(http.StatusOK, "response.tmpl.html", gin.H{
			"message": "Username updated successfully",
		})
	}
}

func UpdatePassword(c *gin.Context) {
	session := sessions.Default(c)
	id := session.Get("userId")
	if id == nil {
		c.HTML(http.StatusUnauthorized, "error.tmpl.html", gin.H{
			"error":   "401 Unauthorized",
			"message": "User not logged in.",
		})
		return
	}
	switch c.Request.Method {
	case "GET":
		c.HTML(http.StatusOK, "update.tmpl.html", gin.H{
			"type": "password",
		})
	case "POST":
		newPassword := c.PostForm("password")
		user := database.ReadUserById(id.(string))
		if user.CheckPassword(newPassword) {
			c.HTML(http.StatusForbidden, "error.tmpl.html", gin.H{
				"error":   "403 Forbidden",
				"message": "New password cannot cannot be same as the current.",
			})
			return
		}
		// Create hash of new password and update it
		user.Password = newPassword
		user.HashPassword()
		if result := database.UpdateUser(id.(string), map[string]any{"password": user.Password}); !result {
			c.HTML(http.StatusBadRequest, "error.tmpl.html", gin.H{
				"error":   "400 Bad Request",
				"message": "Unable to change password, try again later.",
			})
			return
		}
		c.HTML(http.StatusOK, "response.tmpl.html", gin.H{
			"message": "Password updated successfully",
		})
	}
}

func DeleteUser(c *gin.Context) {
	session := sessions.Default(c)
	id := session.Get("userId")
	if id == nil {
		c.HTML(http.StatusUnauthorized, "error.tmpl.html", gin.H{
			"error":   "401 Unauthorized",
			"message": "User not logged in.",
		})
		return
	}
	switch c.Request.Method {
	case "GET":
		c.HTML(http.StatusOK, "delete.tmpl.html", gin.H{
			"oauth": database.IsOAuthUser(id.(string)),
		})
	case "POST":
		user := database.ReadUserById(id.(string))
		// Password required for users who didn't sign up through OAuth
		if !database.IsOAuthUser(user.Id) {
			password := c.PostForm("password")
			if !user.CheckPassword(password) {
				c.HTML(http.StatusForbidden, "error.tmpl.html", gin.H{
					"error":   "403 Forbidden",
					"message": "Incorrect password.",
				})
				return
			}
		}
		if result := database.DeleteUser(user.Id); !result {
			c.HTML(http.StatusBadRequest, "error.tmpl.html", gin.H{
				"error":   "400 Bad Request",
				"message": "Unable to delete account, try again later.",
			})
			return
		}
		session := sessions.Default(c)
		session.Clear()
		session.Options(sessions.Options{Path: "/", MaxAge: -1})
		session.Save()
		c.HTML(http.StatusOK, "response.tmpl.html", gin.H{
			"message": "Account deleted succesfully. つき が つかって くれて ありがとう ございました。",
		})
	}
}

func ToggleFollow(c *gin.Context) {
	session := sessions.Default(c)
	id := session.Get("userId")
	if id == nil {
		c.HTML(http.StatusUnauthorized, "error.tmpl.html", gin.H{
			"error":   "401 Unauthorized",
			"message": "User not logged in.",
		})
		return
	}
	username := c.Param("username")
	toFollow := database.ReadUserByName(username)
	database.ToggleFollow(id.(string), toFollow.Id)
	c.Redirect(http.StatusFound, "/user/"+username)
}
