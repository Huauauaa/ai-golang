package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
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

var mysqlIdentifierPattern = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

type dbConfig struct {
	Host          string
	Port          string
	User          string
	Password      string
	Name          string
	MaxRetries    int
	RetryInterval int
}

func main() {
	db, err := connectDatabase(loadDBConfig())
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
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

func connectDatabase(cfg dbConfig) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	for i := 1; i <= cfg.MaxRetries; i++ {
		if err = ensureDatabaseExists(cfg); err != nil {
			log.Printf("database setup attempt %d/%d failed: %v", i, cfg.MaxRetries, err)
			time.Sleep(time.Duration(cfg.RetryInterval) * time.Second)
			continue
		}

		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.User,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			cfg.Name,
		)

		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Printf("database connection attempt %d/%d failed: %v", i, cfg.MaxRetries, err)
			time.Sleep(time.Duration(cfg.RetryInterval) * time.Second)
			continue
		}

		sqlDB, dbErr := db.DB()
		if dbErr != nil {
			err = dbErr
			log.Printf("database connection attempt %d/%d failed: %v", i, cfg.MaxRetries, err)
			time.Sleep(time.Duration(cfg.RetryInterval) * time.Second)
			continue
		}

		if pingErr := sqlDB.Ping(); pingErr != nil {
			err = pingErr
			_ = sqlDB.Close()
			log.Printf("database connection attempt %d/%d failed: %v", i, cfg.MaxRetries, err)
			time.Sleep(time.Duration(cfg.RetryInterval) * time.Second)
			continue
		}

		if ddlErr := ensureTodoDDL(db); ddlErr != nil {
			err = ddlErr
			_ = sqlDB.Close()
			log.Printf("database connection attempt %d/%d failed: %v", i, cfg.MaxRetries, err)
			time.Sleep(time.Duration(cfg.RetryInterval) * time.Second)
			continue
		}

		return db, nil
	}

	return nil, err
}

func loadDBConfig() dbConfig {
	return dbConfig{
		Host:          getEnv("DB_HOST", "127.0.0.1"),
		Port:          getEnv("DB_PORT", "3306"),
		User:          getEnv("DB_USER", "root"),
		Password:      getEnv("DB_PASSWORD", "root"),
		Name:          getEnv("DB_NAME", "app"),
		MaxRetries:    getEnvAsInt("DB_RETRY_TIMES", 20),
		RetryInterval: getEnvAsInt("DB_RETRY_INTERVAL_SECONDS", 2),
	}
}

func ensureDatabaseExists(cfg dbConfig) error {
	if !mysqlIdentifierPattern.MatchString(cfg.Name) {
		return fmt.Errorf("invalid database name: %s", cfg.Name)
	}

	adminDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
	)

	db, err := gorm.Open(mysql.Open(adminDSN), &gorm.Config{})
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		return err
	}

	createDatabaseSQL := fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
		cfg.Name,
	)

	return db.Exec(createDatabaseSQL).Error
}

func ensureTodoDDL(db *gorm.DB) error {
	const createTodosTableDDL = `
CREATE TABLE IF NOT EXISTS todos (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  title VARCHAR(255) NOT NULL,
  completed TINYINT(1) NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`

	return db.Exec(createTodosTableDDL).Error
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
