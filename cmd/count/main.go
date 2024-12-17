package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
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

func (h *Handlers) PostCount(c echo.Context) error {
	countStr := c.FormValue("count")
	if countStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "count is empty")
	}

	count, err := strconv.Atoi(countStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid count value: %v", err))
	}

	if err := h.dbProvider.incrementCount(count); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error").SetInternal(err)
	}

	return c.String(http.StatusOK, fmt.Sprintf("Value %d has been added to count", count))
}

func (h *Handlers) GetCount(c echo.Context) error {
	val, err := h.dbProvider.SelectCount()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error").SetInternal(err)
	}

	return c.String(http.StatusOK, fmt.Sprintf("Current count value: %d", val))
}

func (dp *DatabaseProvider) SelectCount() (int, error) {
	var val int
	err := dp.db.QueryRow("SELECT value FROM count WHERE id = 1").Scan(&val)
	if err != nil {
		return 0, err
	}
	return val, nil
}

func (dp *DatabaseProvider) incrementCount(n int) error {
	_, err := dp.db.Exec("UPDATE count SET value = value + ($1) WHERE id = 1", n)
	return err
}

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS count (
		id SERIAL PRIMARY KEY,
		value INTEGER NOT NULL DEFAULT 0
	)
	`)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	_, err = db.Exec("INSERT INTO count (id, value) VALUES (1, 0) ON CONFLICT (id) DO NOTHING")
	if err != nil {
		log.Fatalf("Failed to insert initial data: %v", err)
	}

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	dp := DatabaseProvider{db: db}
	h := Handlers{dbProvider: dp}

	e.GET("/count", h.GetCount)
	e.POST("/count", h.PostCount)

	e.Logger.Fatal(e.Start(":3333"))
}
