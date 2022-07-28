package routes

import (
	"net/http"
	"time"

	"github.com/Devansh3712/tsuki/database"
	"github.com/Devansh3712/tsuki/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
)

var commentLimit = 10

func NewPost(c *gin.Context) {
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
		c.HTML(http.StatusOK, "makePost.tmpl.html", nil)
	case "POST":
		var post models.Post
		if err := c.Request.ParseForm(); err != nil {
			c.HTML(http.StatusBadRequest, "error.tmpl.html", gin.H{
				"error":   "400 Bad Request",
				"message": "Unable to parse form.",
			})
			return
		}
		if err := c.ShouldBindWith(&post, binding.Form); err != nil {
			c.HTML(http.StatusBadRequest, "error.tmpl.html", gin.H{
				"error":   "400 Bad Request",
				"message": err.Error(),
			})
			return
		}
		post.Id = uuid.NewString()
		post.CreatedAt = time.Now()
		if result := database.CreatePost(id.(string), &post); !result {
			c.HTML(http.StatusBadRequest, "error.tmpl.html", gin.H{
				"error":   "400 Bad Request",
				"message": "Unable to create post, try again later.",
			})
			return
		}
		c.Redirect(http.StatusFound, "/post/"+post.Id)
	}
}

func GetPost(c *gin.Context) {
	var self, voted bool
	session := sessions.Default(c)
	id := session.Get("userId")
	postId := c.Param("id")
	post := database.ReadPost(postId)
	if post == nil {
		c.HTML(http.StatusNotFound, "error.tmpl.html", gin.H{
			"error":   "404 Not Found",
			"message": "Post not found or doesn't exist.",
		})
		return
	}
	if c.Query("more") == "true" {
		commentLimit += 10
	} else {
		commentLimit = 10
	}
	comments := database.ReadComments(post.Id, commentLimit)
	if id != nil {
		// Check if current user has voted on post
		voted = database.Voted(id.(string), post.Id)
		// Enable delete post if its current user's post
		if id.(string) == post.UserId {
			self = true
		}
		for index := range comments {
			comments[index].Username = database.ReadUserById(comments[index].UserId).Username
			// Enable delete comment if its current user's comment
			if id.(string) == comments[index].UserId {
				comments[index].Self = true
				break
			}
		}
	}
	c.HTML(http.StatusOK, "getPost.tmpl.html", gin.H{
		"author":   database.ReadUserById(post.UserId),
		"post":     post,
		"self":     self,
		"voted":    voted,
		"voters":   database.ReadVotes(post.Id),
		"comments": comments,
	})
}

func DeletePost(c *gin.Context) {
	session := sessions.Default(c)
	id := session.Get("userId")
	if id == nil {
		c.HTML(http.StatusUnauthorized, "error.tmpl.html", gin.H{
			"error":   "401 Unauthorized",
			"message": "User not logged in.",
		})
		return
	}
	postId := c.Param("id")
	post := database.ReadPost(postId)
	if id.(string) != post.UserId {
		c.HTML(http.StatusUnauthorized, "error.tmpl.html", gin.H{
			"error":   "401 Unauthorized",
			"message": "Cannot perform this task.",
		})
		return
	}
	if result := database.DeletePost(post.Id); !result {
		c.HTML(http.StatusBadRequest, "error.tmpl.html", gin.H{
			"error":   "400 Bad Request",
			"message": "Unable to delete post, try again later.",
		})
		return
	}
	c.HTML(http.StatusOK, "response.tmpl.html", gin.H{
		"message": "Post deleted successfully.",
	})
}

func ToggleVote(c *gin.Context) {
	session := sessions.Default(c)
	id := session.Get("userId")
	if id == nil {
		c.HTML(http.StatusUnauthorized, "error.tmpl.html", gin.H{
			"error":   "401 Unauthorized",
			"message": "User not logged in.",
		})
		return
	}
	postId := c.Param("id")
	database.ToggleVote(id.(string), postId)
	c.Redirect(http.StatusFound, "/post/"+postId)
}

func Comment(c *gin.Context) {
	session := sessions.Default(c)
	id := session.Get("userId")
	if id == nil {
		c.HTML(http.StatusUnauthorized, "error.tmpl.html", gin.H{
			"error":   "401 Unauthorized",
			"message": "User not logged in.",
		})
		return
	}
	var comment models.Comment
	if err := c.Request.ParseForm(); err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl.html", gin.H{
			"error":   "400 Bad Request",
			"message": "Unable to parse form.",
		})
		return
	}
	if err := c.ShouldBindWith(&comment, binding.Form); err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl.html", gin.H{
			"error":   "400 Bad Request",
			"message": err.Error(),
		})
		return
	}
	postId := c.Param("id")
	comment.Id = uuid.NewString()
	comment.CreatedAt = time.Now()
	if result := database.CreateComment(id.(string), postId, &comment); !result {
		c.HTML(http.StatusBadRequest, "error.tmpl.html", gin.H{
			"error":   "400 Bad Request",
			"message": "Unable to add comment, try again later.",
		})
		return
	}
	c.Redirect(http.StatusFound, "/post/"+postId)
}

func DeleteComment(c *gin.Context) {
	session := sessions.Default(c)
	id := session.Get("userId")
	if id == nil {
		c.HTML(http.StatusUnauthorized, "error.tmpl.html", gin.H{
			"error":   "401 Unauthorized",
			"message": "User not logged in.",
		})
		return
	}
	postId := c.Param("id")
	commentId := c.Query("commentId")
	comment := database.ReadComment(commentId)
	if comment == nil {
		c.HTML(http.StatusNotFound, "error.tmpl.html", gin.H{
			"error":   "404 Not Found",
			"message": "Comment not found.",
		})
		return
	}
	if id.(string) != comment.UserId {
		c.HTML(http.StatusUnauthorized, "error.tmpl.html", gin.H{
			"error":   "401 Unauthorized",
			"message": "Cannot perform this task.",
		})
		return
	}
	if result := database.DeleteComment(commentId); !result {
		c.HTML(http.StatusBadRequest, "error.tmpl.html", gin.H{
			"error":   "400 Bad Request",
			"message": "Unable to delete comment, try again later.",
		})
		return
	}
	c.Redirect(http.StatusFound, "/post/"+postId)
}
