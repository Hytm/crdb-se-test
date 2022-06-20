package main

import (
	"context"
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
)

type Request struct {
	Url       string `json:"url"`
	Frequency int    `json:"frequency"`
}

type Response struct {
	Id        uuid.UUID `json:"id"`
	Url       string    `json:"url"`
	Frequency int       `json:"frequency"`
}

type DeleteCommand struct {
	Id uuid.UUID `json:"id"`
}

type feeds struct {
	id  uuid.UUID
	url string
}

type Feed struct {
	Entries []Entry `xml:"entry"`
}

type Entry struct {
	Link struct {
		Href string `xml:"href,attr"`
	} `xml:"link"`
	Thumbnail struct {
		URL string `xml:"url,attr"`
	} `xml:"thumbnail"`
	Title string `xml:"title"`
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

	//start a worker to process feeds into recipes
	go func() {
		for range time.Tick(1 * time.Hour) {
			log.Println("processing feeds")
			processFeeds(false)
		}
	}()
	router := gin.Default()
	router.POST("/", AddFeedHandler)
	router.GET("/", ListFeedHandler)
	router.DELETE("/", DeleteFeedHandler)
	router.PUT("/", ForceProcessFeedsHandler)
	router.Run(":8090")
}

func AddFeedHandler(c *gin.Context) {
	var r Request
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := connection.Exec(context.Background(), "INSERT INTO feeds (url, frequency) VALUES ($1, $2)", r.Url, r.Frequency)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while publishing to queue"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

func ListFeedHandler(c *gin.Context) {
	rows, err := connection.Query(context.Background(), "SELECT id, url, frequency FROM feeds")
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while reading feeds table"})
		return
	}
	defer rows.Close()

	var feeds []Response
	for rows.Next() {
		var r Response
		if err := rows.Scan(&r.Id, &r.Url, &r.Frequency); err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while reading feeds"})
			return
		}
		feeds = append(feeds, r)
	}
	c.JSON(http.StatusOK, gin.H{"feeds": feeds})
}

func DeleteFeedHandler(c *gin.Context) {
	var r DeleteCommand
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := connection.Exec(context.Background(), "DELETE FROM feeds WHERE id = $1", r.Id)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while deleting feed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

func ForceProcessFeedsHandler(c *gin.Context) {
	processFeeds(true)
	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

func processFeeds(force bool) {
	q := "SELECT id, url FROM feeds WHERE COALESCE(AGE(last_update)::INT, 10000) > frequency"
	if force {
		q = "SELECT id, url FROM feeds"
	}

	rows, err := connection.Query(context.Background(), q)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var f feeds
		if err := rows.Scan(&f.id, &f.url); err != nil {
			log.Println(err)
			return
		}
		getUpdateFromFeed(f.id, f.url)
	}
}

//Adding recipes update from feeds
func getUpdateFromFeed(id uuid.UUID, url string) {
	log.Println("getting update from feed", url)
	c := &http.Client{}
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
		return
	}
	r.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.110 Safari/537.36")
	resp, err := c.Do(r)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	bValue, _ := ioutil.ReadAll(resp.Body)
	var feed Feed
	xml.Unmarshal(bValue, &feed)

	tx, err := connection.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			log.Println("rolling back transaction for feed", url)
			tx.Rollback(context.Background())
		} else {
			log.Println("committing transaction for feed", url)
			tx.Commit(context.Background())
		}
	}()
	err = addNewRecipes(tx, feed.Entries)
	if err != nil {
		return
	}
	err = updateFeedLastUpdate(tx, id)
	if err != nil {
		return
	}
}

func addNewRecipes(tx pgx.Tx, entries []Entry) error {
	for _, entry := range entries {
		log.Println("adding recipe", entry.Link.Href)
		_, err := tx.Exec(context.Background(), "INSERT INTO recipes (title, thumbnail, url) VALUES ($1, $2, $3)", entry.Title, entry.Thumbnail.URL, entry.Link.Href)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}

func updateFeedLastUpdate(tx pgx.Tx, id uuid.UUID) error {
	_, err := tx.Exec(context.Background(), "UPDATE feeds SET last_update = NOW() WHERE id = $1", id)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
