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

type Message struct {
	Msg string `json:"msg"`
}

func (h *Handlers) GetHello(c echo.Context) error {
	msg, err := h.dbProvider.SelectHello()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{"message": msg})
}

func (h *Handlers) PostHello(c echo.Context) error {
	var input Message
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input: "+err.Error())
	}

	if err := h.dbProvider.InsertHello(input.Msg); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error: "+err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "Message inserted"})
}

func (dp *DatabaseProvider) SelectHello() (string, error) {
	var msg string

	row := dp.db.QueryRow("SELECT message FROM hello ORDER BY RANDOM() LIMIT 1")
	err := row.Scan(&msg)
	if err != nil {
		return "", err
	}

	return msg, nil
}

func (dp *DatabaseProvider) InsertHello(msg string) error {
	_, err := dp.db.Exec("INSERT INTO hello (message) VALUES ($1)", msg)
	return err
}

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS hello (
		id SERIAL PRIMARY KEY,
		message TEXT NOT NULL
	)
	`)
	if err != nil {
		log.Fatal(err)
	}

	dp := DatabaseProvider{db: db}
	h := Handlers{dbProvider: dp}

	e := echo.New()


	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/get", h.GetHello)
	e.POST("/post", h.PostHello)


/* 	rand.Seed(time.Now().UnixNano()) */
	if err := e.Start(":8081"); err != nil {
		log.Fatal(err)
	}
}
