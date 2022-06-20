package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
)

type Recipe struct {
	Title     string `json:"title"`
	URL       string `json:"url"`
	Thumbnail string `json:"thumbnail"`
}

var connection *pgx.Conn

//export DATABASE_URL=postgres://root@127.0.0.1:26257/recipes?sslmode=disable
func main() {
	var err error
	connection, err = pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
		return
	}
	defer connection.Close(context.Background())

	log.Println("Connected to database")

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "OPTIONS"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	router.GET("/", GetRecipesHandler)
	router.Run(":8091")
}

func GetRecipesHandler(c *gin.Context) {
	rows, err := connection.Query(context.Background(), "SELECT title, url, thumbnail FROM recipes")
	if err != nil {
		log.Fatal(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not get recipes"})
		return
	}

	var recipes []Recipe
	for rows.Next() {
		var recipe Recipe
		err := rows.Scan(&recipe.Title, &recipe.URL, &recipe.Thumbnail)
		if err != nil {
			log.Fatal(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not get recipes"})
			return
		}
		recipes = append(recipes, recipe)
	}
	c.JSON(http.StatusOK, gin.H{"recipes": recipes})
}
