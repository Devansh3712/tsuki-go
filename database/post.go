package database

import (
	"log"

	"github.com/Devansh3712/tsuki/models"
)

func CreatePost(userId string, post *models.Post) bool {
	if _, err := db.Exec(
		`INSERT INTO posts(user_id, id, body, created_at)
		VALUES ($1, $2, $3, $4)`,
		userId, post.Id, post.Body, post.CreatedAt,
	); err != nil {
		log.Println(err)
		return false
	}
	return true
}

func ReadPost(id string) *models.Post {
	var post models.Post
	if err := db.QueryRow(`SELECT * FROM posts WHERE id = $1`, id).Scan(
		&post.UserId, &post.Id, &post.Body, &post.CreatedAt,
	); err != nil {
		log.Println(err)
		return nil
	}
	return &post
}

func ReadPostsCount(userId string) int {
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM posts WHERE user_id = $1`, userId).Scan(&count); err != nil {
		log.Println(err)
		return 0
	}
	return count
}

func ReadPosts(userId string, limit int) []models.Post {
	var posts []models.Post
	rows, err := db.Query(
		`SELECT * FROM posts WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`,
		userId, limit,
	)
	if err != nil {
		log.Println(err)
		return nil
	}

	defer rows.Close()
	for rows.Next() {
		var post models.Post
		rows.Scan(&post.UserId, &post.Id, &post.Body, &post.CreatedAt)
		posts = append(posts, post)
	}
	return posts
}

func ReadFeedPosts(userId string, limit int) []models.Post {
	var posts []models.Post
	rows, err := db.Query(
		`SELECT * FROM posts WHERE user_id IN
		(SELECT follow_id FROM follows WHERE user_id = $1)
		ORDER BY created_at DESC LIMIT $2`,
		userId, limit,
	)
	if err != nil {
		log.Println(err)
		return nil
	}

	defer rows.Close()
	for rows.Next() {
		var post models.Post
		rows.Scan(&post.UserId, &post.Id, &post.Body, &post.CreatedAt)
		posts = append(posts, post)
	}
	return posts
}

func DeletePost(id string) bool {
	if _, err := db.Exec(`DELETE FROM posts WHERE id = $1`, id); err != nil {
		log.Println(err)
		return false
	}
	return true
}

func Voted(userId string, id string) bool {
	var count int
	db.QueryRow(
		`SELECT COUNT(*) FROM votes WHERE user_id = $1 AND id = $2`,
		userId, id,
	).Scan(&count)

	switch count {
	case 0:
		return false
	default:
		return true
	}
}

func ToggleVote(userId string, id string) {
	var query string
	voted := Voted(userId, id)

	switch voted {
	case false:
		query = `INSERT INTO votes (user_id, id) VALUES ($1, $2)`
	default:
		query = `DELETE FROM votes WHERE user_id = $1 AND id = $2`
	}
	if _, err := db.Exec(query, userId, id); err != nil {
		log.Println(err)
	}
}

func ReadVotes(id string) []string {
	var voters []string
	rows, err := db.Query(
		`SELECT username FROM t_users WHERE id IN
		(SELECT user_id FROM votes WHERE id = $1)`,
		id,
	)
	if err != nil {
		log.Println(err)
		return nil
	}

	defer rows.Close()
	for rows.Next() {
		var username string
		rows.Scan(&username)
		voters = append(voters, username)
	}
	return voters
}

func CreateComment(userId string, postId string, comment *models.Comment) bool {
	if _, err := db.Exec(
		`INSERT INTO comments (user_id, post_id, id, body, created_at)
		VALUES ($1, $2, $3, $4, $5)`,
		userId, postId, comment.Id, comment.Body, comment.CreatedAt,
	); err != nil {
		log.Println(err)
		return false
	}
	return true
}

func ReadComment(id string) *models.Comment {
	var comment models.Comment
	if err := db.QueryRow(`SELECT * FROM comments WHERE id = $1`, id).Scan(
		&comment.UserId,
		&comment.PostId,
		&comment.Id,
		&comment.Body,
		&comment.CreatedAt,
	); err != nil {
		log.Println(err)
		return nil
	}
	return &comment
}

func ReadComments(postId string, limit int) []models.Comment {
	var comments []models.Comment
	rows, err := db.Query(`SELECT * FROM comments WHERE post_id = $1 LIMIT $2`, postId, limit)
	if err != nil {
		log.Println(err)
		return nil
	}

	defer rows.Close()
	for rows.Next() {
		var comment models.Comment
		rows.Scan(
			&comment.UserId,
			&comment.PostId,
			&comment.Id,
			&comment.Body,
			&comment.CreatedAt,
		)
		comments = append(comments, comment)
	}
	return comments
}

func DeleteComment(id string) bool {
	if _, err := db.Exec(`DELETE FROM comments WHERE id = $1`, id); err != nil {
		log.Println(err)
		return false
	}
	return true
}
