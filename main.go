package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

var db *sql.DB

func initDB() {
	var err error
	connStr := "postgres://bookstore:Singh1337@35.222.18.238:5432/bookstore?sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error opening database: %q", err)
	}

	// Verify connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Cannot connect to the database: %v", err)
	}
	log.Println("Connected to the database")
	// Create table if it doesn't exist
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS albums (
		id SERIAL PRIMARY KEY,
		title TEXT NOT NULL,
		artist TEXT NOT NULL,
		price NUMERIC(10, 2) NOT NULL
	);`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatalf("Error creating table: %v", err)
	}
	log.Println("Table 'albums' is ready")

}

// album represents data about a record album.
type album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

// getAlbums responds with the list of all albums as JSON.
func getAlbums(c *gin.Context) {
	rows, err := db.Query("SELECT id, title, artist, price FROM albums")
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	var albums []album
	for rows.Next() {
		var a album
		if err := rows.Scan(&a.ID, &a.Title, &a.Artist, &a.Price); err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Error scanning data"})
			return
		}
		albums = append(albums, a)
	}

	c.IndentedJSON(http.StatusOK, albums)
}

// postAlbums adds an album from JSON received in the request body.

func postAlbums(c *gin.Context) {
	var newAlbum album

	if err := c.BindJSON(&newAlbum); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Insert the album and return the generated ID
	var lastInsertID int
	err := db.QueryRow("INSERT INTO albums (title, artist, price) VALUES ($1, $2, $3) RETURNING id",
		newAlbum.Title, newAlbum.Artist, newAlbum.Price).Scan(&lastInsertID)

	if err != nil {
		log.Printf("Error inserting album: %v", err)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert album"})
		return
	}

	// Set the ID for the new album
	newAlbum.ID = fmt.Sprintf("%d", lastInsertID)

	c.IndentedJSON(http.StatusCreated, newAlbum)
}

// getAlbumByID locates the album whose ID value matches the id
// parameter sent by the client, then returns that album as a response.
func getAlbumByID(c *gin.Context) {
	id := c.Param("id")
	var a album

	err := db.QueryRow("SELECT id, title, artist, price FROM albums WHERE id = $1", id).Scan(&a.ID, &a.Title, &a.Artist, &a.Price)
	if err == sql.ErrNoRows {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
		return
	} else if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.IndentedJSON(http.StatusOK, a)
}

func deleteAlbumByID(c *gin.Context) {
	id := c.Param("id")

	result, err := db.Exec("DELETE FROM albums WHERE id = $1", id)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete album"})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Error checking rows affected"})
		return
	}

	if rowsAffected == 0 {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "album deleted"})
}

func main() {
	initDB()
	defer db.Close()
	router := gin.Default()
	router.Use(cors.Default())
	router.GET("/albums", getAlbums)
	router.GET("/albums/:id", getAlbumByID)
	router.POST("/albums", postAlbums)
	router.DELETE("/albums/:id", deleteAlbumByID)

	router.Run("0.0.0.0:8080")
}
