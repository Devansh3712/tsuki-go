package routes

import (
	"net/http"

	"github.com/Devansh3712/tsuki/database"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

var feedLimit = 10

func UserFeed(c *gin.Context) {
	session := sessions.Default(c)
	id := session.Get("userId")
	if id == nil {
		c.HTML(http.StatusUnauthorized, "error.tmpl.html", gin.H{
			"error":   "401 Unauthorized",
			"message": "User not logged in.",
		})
		return
	}
	if c.Query("more") == "true" {
		feedLimit += 10
	} else {
		feedLimit = 10
	}
	posts := database.ReadFeedPosts(id.(string), feedLimit)
	for index := range posts {
		author := database.ReadUserById(posts[index].UserId)
		posts[index].Username = author.Username
		posts[index].Avatar = author.Avatar
	}
	c.HTML(http.StatusOK, "feed.tmpl.html", gin.H{
		"posts": posts,
	})
}
