package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	_ "modernc.org/sqlite"
)

type Task struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

type App struct {
	DB    *sql.DB
	Redis *redis.Client
}

func main() {
	addr := getEnv("APP_ADDR", ":8080")
	dbPath := getEnv("SQLITE_PATH", "/var/lib/gotasker.db")
	redisAddr := getEnv("REDIS_ADDR", "127.0.0.1:6379")

	if err := os.MkdirAll("/var/lib/gotasker", 0755); err != nil {
		log.Fatalf("mkdir /var/lib/gotasker: %v", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	if err := initSchema(db); err != nil {
		log.Fatalf("init schema: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("warning: redis ping failed: %v", err)
	}

	app := &App{
		DB:    db,
		Redis: rdb,
	}

	r := gin.Default()

	r.GET("/healthz", app.handleHealth)
	api := r.Group("/api/tasks")
	{
		api.GET("/", app.handleListTasks)
		api.POST("/", app.handleCreateTask)
		api.DELETE("/", app.handleClearTasks)
		api.POST("/generate", app.handleGenerateTasks)
		api.DELETE("/:id", app.handleDeleteTask)
	}

	log.Printf("listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("listen: %v", err)
	}
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}

	return def
}

func initSchema(db *sql.DB) error {
	schema := `
		CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			completed BOOLEAN NOT NULL DEFAULT 0
		);`

	_, err := db.Exec(schema)
	return err
}

func (a *App) handleHealth(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

func (a *App) handleListTasks(c *gin.Context) {
	ctx := c.Request.Context()

	//Try redis cache
	if a.Redis != nil {
		if data, err := a.Redis.Get(ctx, "tasks:all").Result(); err == nil {
			c.Header("X-Cache-Hit", "true")
			c.Header("Content-Type", "application/json")
			c.String(http.StatusOK, data)
			return
		}
	}

	c.Header("X-Cache-Hit", "false")

	//Fallback to db
	rows, err := a.DB.QueryContext(ctx, "SELECT id, title, completed FROM tasks ORDER BY id DESC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Completed); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		tasks = append(tasks, t)
	}

	// Handle empty list
	if tasks == nil {
		tasks = []Task{}
	}

	jsonData, err := json.Marshal(tasks)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "json error"})
		return
	}

	if a.Redis != nil {
		_ = a.Redis.Set(ctx, "tasks:all", jsonData, 5*time.Second).Err()
	}

	c.Data(http.StatusOK, "application/json", jsonData)
}

func (a *App) handleCreateTask(c *gin.Context) {
	ctx := c.Request.Context()
	var input struct {
		Title string `json:"title"`
	}
	if err := c.BindJSON(&input); err != nil || input.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	res, err := a.DB.ExecContext(ctx, "INSERT INTO tasks(title, completed) VALUES (?, 0)", input.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	id, _ := res.LastInsertId()

	if a.Redis != nil {
		_ = a.Redis.Del(ctx, "tasks:all").Err()
	}

	task := Task{
		ID:        id,
		Title:     input.Title,
		Completed: false,
	}
	c.JSON(http.StatusCreated, task)
}

func (a *App) handleDeleteTask(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if _, err := a.DB.ExecContext(ctx, "DELETE FROM tasks WHERE id = ?", id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	if a.Redis != nil {
		_ = a.Redis.Del(ctx, "tasks:all").Err()
	}

	c.Status(http.StatusNoContent)
}

func (a *App) handleGenerateTasks(c *gin.Context) {
	ctx := c.Request.Context()
	count := 1000
	if c := c.Query("count"); c != "" {
		if n, err := strconv.Atoi(c); err == nil && n > 0 {
			count = n
		}
	}

	tx, err := a.DB.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO tasks(title, completed) VALUES (?, 0)")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	defer stmt.Close()

	for i := 1; i <= count; i++ {
		if _, err := stmt.ExecContext(ctx, fmt.Sprintf("Task %d", i)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	if a.Redis != nil {
		_ = a.Redis.Del(ctx, "tasks:all").Err()
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("generated %d tasks", count)})
}

func (a *App) handleClearTasks(c *gin.Context) {
	ctx := c.Request.Context()

	if _, err := a.DB.ExecContext(ctx, "DELETE FROM tasks"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	if a.Redis != nil {
		_ = a.Redis.Del(ctx, "tasks:all").Err()
	}

	c.Status(http.StatusNoContent)
}
