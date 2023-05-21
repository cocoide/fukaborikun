package database

import (
	"context"
	"log"
	"os"

	"github.com/go-redis/redis"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewDatabse() *gorm.DB {
	env := os.Getenv("APP_ENV")
	var DSN string
	switch env {
	case "pro":
		DSN = os.Getenv("MYSQL_URL")
	case "dev":
		DSN = "kazuki:secret@tcp(db:3306)/mydb?charset=utf8mb4&parseTime=True&loc=Asia%2FTokyo"
	default:
		log.Print("databse environment isn't properly selected")
	}
	db, err := gorm.Open(
		mysql.Open(DSN),
		&gorm.Config{DisableForeignKeyConstraintWhenMigrating: true},
	)
	if err != nil {
		log.Printf("failed to connect with %s databse: %s", env, err.Error())
	}
	log.Printf("%s databse connected !! 📦", env)
	return db
}

func NewRedisCilent(ctx context.Context) *redis.Client {
	env := os.Getenv("APP_ENV")
	client := &redis.Client{}
	switch env {
	case "pro":
		REDIS_URL := os.Getenv("REDIS_URL")
		option, _ := redis.ParseURL(REDIS_URL)
		client = redis.NewClient(option)
	case "dev":
		client = redis.NewClient(&redis.Options{
			Addr:     "redis:6379",
			Password: "",
			DB:       0,
		})
	}
	_, err := client.Ping().Result()
	if err != nil {
		log.Fatalf("failed to connect to %s redis: %v", env, err)
	}
	log.Printf("%s redis connected !!", env)
	return client
}
