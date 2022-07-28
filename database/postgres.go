package database

import (
	"database/sql"
	"io/ioutil"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var db *sql.DB

func init() {
	godotenv.Load(".env")
	var err error
	db, err = sql.Open("postgres", os.Getenv("POSTGRES_URI"))
	if err != nil {
		panic(err)
	}
	data, err := ioutil.ReadFile("database/init.sql")
	if err != nil {
		panic(err)
	}
	script := string(data)
	if _, err := db.Exec(script); err != nil {
		panic(err)
	}
}
