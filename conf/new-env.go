package conf

import (
	"log"

	"github.com/joho/godotenv"
)

func NewEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("failed to load .env file: %v" + err.Error())
	} else {
		log.Print(".env file properly loaded")
	}
}
