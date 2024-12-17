package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "BKMZ661248082"
	dbname   = "Sandbox"
)

type Handlers struct {
	dbProvider DatabaseProvider
}

type DatabaseProvider struct {
	db *sql.DB
}


func (h *Handlers) PostQuery(c echo.Context) error {
	name := c.QueryParam("name") 
	if name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Name is empty")
	}

	if err := h.dbProvider.insertName(name); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error: "+err.Error())
	}

	return c.String(http.StatusOK, "Name is posted")
}


func (h *Handlers) GetQuery(c echo.Context) error {
	name := c.QueryParam("name") 
	if name == "" {
		return c.String(http.StatusBadRequest, "Name is empty")
	}

	if err := h.dbProvider.selectName(name); err != nil {
		return c.String(http.StatusNotFound, "User does not exist!")
	}

	return c.String(http.StatusOK, "Hello, "+name+"!")
}

func (dp *DatabaseProvider) insertName(name string) error {
	_, err := dp.db.Exec("INSERT INTO query (name) VALUES ($1)", name)
	return err
}


func (dp *DatabaseProvider) selectName(name string) error {
	var exist string
	err := dp.db.QueryRow("SELECT name FROM query WHERE name = $1", name).Scan(&exist)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("Database connection error:", err)
	}
	defer db.Close()

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS query (
        id SERIAL PRIMARY KEY,
        name VARCHAR(50)
	)
	`)
	if err != nil {
		log.Fatal("Table creation error:", err)
	}

	dp := DatabaseProvider{db: db}
	h := Handlers{dbProvider: dp}

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	api := e.Group("/api/user")
	api.GET("/get", h.GetQuery)  
	api.POST("/post", h.PostQuery) 

	if err := e.Start(":9000"); err != nil {
		log.Fatal("Server error:", err)
	}
}
