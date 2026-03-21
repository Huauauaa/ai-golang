package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Todo struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Title     string    `json:"title" gorm:"not null"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type createTodoRequest struct {
	Title string `json:"title" binding:"required,min=2,max=255"`
}

func main() {
	db, err := connectDatabase()
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	if err := db.AutoMigrate(&Todo{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	router := gin.Default()
	router.Use(cors.Default())

	router.GET("/api/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	router.GET("/api/todos", func(ctx *gin.Context) {
		var todos []Todo
		if err := db.Order("id DESC").Find(&todos).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, todos)
	})

	router.POST("/api/todos", func(ctx *gin.Context) {
		var req createTodoRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		todo := Todo{
			Title:     req.Title,
			Completed: false,
		}

		if err := db.Create(&todo).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusCreated, todo)
	})

	port := getEnv("PORT", "8080")
	log.Printf("backend server listening on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func connectDatabase() (*gorm.DB, error) {
	host := getEnv("DB_HOST", "127.0.0.1")
	port := getEnv("DB_PORT", "3306")
	user := getEnv("DB_USER", "app")
	password := getEnv("DB_PASSWORD", "app123456")
	name := getEnv("DB_NAME", "app")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user,
		password,
		host,
		port,
		name,
	)

	maxRetries := getEnvAsInt("DB_RETRY_TIMES", 20)
	retryInterval := getEnvAsInt("DB_RETRY_INTERVAL_SECONDS", 2)

	var db *gorm.DB
	var err error

	for i := 1; i <= maxRetries; i++ {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			sqlDB, dbErr := db.DB()
			if dbErr == nil && sqlDB.Ping() == nil {
				return db, nil
			}
			if dbErr != nil {
				err = dbErr
			} else {
				err = errors.New("ping database failed")
			}
		}

		log.Printf("database connection attempt %d/%d failed: %v", i, maxRetries, err)
		time.Sleep(time.Duration(retryInterval) * time.Second)
	}

	return nil, err
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvAsInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	result, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return result
}
