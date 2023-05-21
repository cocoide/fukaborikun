package main

import (
	"context"
	"log"

	"github.com/cocoide/fukaborikun/pkg/database"
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
	ctx := context.Background()
	database.NewDatabse()
	database.NewRedisCilent(ctx)

	lg := gateway.NewLineAPIGateway()
	wh := handler.NewWebHookHandler(lg)
	e.POST("/linebot-webhook", wh.HandleLineEvent)
	e.Logger.Fatal(e.Start(":8080"))
}
