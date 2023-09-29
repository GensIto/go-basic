package main

import (
	"database/sql"
	"log"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func initDB(filepath string) *sql.DB {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		log.Fatal()
	}
	return db
}

func validateUser(name string, age int) error {
	if name == "" {
		return echo.NewHTTPError(400, "name is required field")
	}
	if len(name) > 100 {
		return echo.NewHTTPError(400, "name is too long")
	}
	if age < 0 || age > 150 {
		return echo.NewHTTPError(400, "age must be between 0 and 150")
	}
	return nil
}

func main() {
	db := initDB("example.db")

	e := echo.New()
	e.Use(middleware.Logger())

	e.POST("/users", func(c echo.Context) error {
		name := c.FormValue("name")
		age, _ := strconv.Atoi(c.FormValue("age"))
		result, err := db.Exec("INSERT INTO users (name, age) VALUES (?, ?)", name, age)

		if err := validateUser(name, age); err != nil {
			return err
		}

		if err != nil {
			return echo.NewHTTPError(500, err.Error())
		}

		id, _ := result.LastInsertId()
		return c.JSON(201, &User{
			ID:   int(id),
			Name: name,
			Age:  age,
		})
	})

	e.DELETE("/users/:id", func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(500, err.Error())
		}
		result, err := db.Exec("DELETE FROM users WHERE id = ?", id)
		if err != nil {
			return echo.NewHTTPError(500, err.Error())
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return echo.NewHTTPError(404, "Not Found")
		}
		return c.NoContent(204)
	})

	e.PUT("/users/:id", func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(500, err.Error())
		}
		name := c.FormValue("name")
		age, _ := strconv.Atoi(c.FormValue("age"))

		if err := validateUser(name, age); err != nil {
			return err
		}

		result, err := db.Exec("UPDATE users SET name = ?, age = ? WHERE id = ?", name, age, id)
		if err != nil {
			return echo.NewHTTPError(500, err.Error())
		}

		rows, _ := result.RowsAffected()
		if rows == 0 {
			return echo.NewHTTPError(404, "Not Found")
		}

		return c.JSON(200, &User{
			ID:   id,
			Name: name,
			Age:  age,
		})
	})

	e.GET("/users", func(c echo.Context) error {
		rows, err := db.Query("SELECT id, name, age FROM users")
		if err != nil {
			return echo.NewHTTPError(500, err.Error())
		}
		defer rows.Close()

		users := []User{}
		for rows.Next() {
			var user User
			if err := rows.Scan(&user.ID, &user.Name, &user.Age); err != nil {
				return echo.NewHTTPError(500, err.Error())
			}
			users = append(users, user)
		}
		return c.JSON(200, users)
	})

	e.GET("/users/:id", func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(500, err.Error())
		}
		row := db.QueryRow("SELECT id, name, age FROM users WHERE id = ?", id)
		var user User
		if err := row.Scan(&user.ID, &user.Name, &user.Age); err != nil {
			return echo.NewHTTPError(500, err.Error())
		}
		return c.JSON(200, user)
	})

	e.Start(":8080")
}
