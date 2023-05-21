package main

import (
	"log"

	"github.com/cocoide/fukaborikun/pkg/gateway"
	"github.com/cocoide/fukaborikun/pkg/handler"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("failed to load .env file: %v" + err.Error())
	} else {
		log.Print(".env file properly loaded")
	}
	// DSN := "kazuki:secret@tcp(db:3306)/mydb?charset=utf8mb4&parseTime=True&loc=Asia%2FTokyo"
	// _, err := gorm.Open(mysql.Open(DSN))
	// if err != nil {
	// 	log.Fatalf("failed to connect with databse: %s", err.Error())
	// }
	lg := gateway.NewLineAPIGateway()
	wh := handler.NewWebHookHandler(lg)
	e.POST("/linebot-webhook", wh.HandleLineEvent)
	e.Logger.Fatal(e.Start(":8080"))
}
