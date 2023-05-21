package main

import (
	"context"

	"github.com/cocoide/fukaborikun/conf"
	"github.com/cocoide/fukaborikun/pkg/database"
	"github.com/cocoide/fukaborikun/pkg/gateway"
	"github.com/cocoide/fukaborikun/pkg/handler"
	"github.com/cocoide/fukaborikun/pkg/repository"
	"github.com/cocoide/fukaborikun/pkg/usecase"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	conf.NewEnv()
	ctx := context.Background()
	database.NewDatabse()
	rb := database.NewRedisCilent(ctx)

	cr := repository.NewCacheRepo(rb)
	og := gateway.NewOpenAIGateway(ctx)
	lg := gateway.NewLineAPIGateway()
	du := usecase.NewDialogUseCase(cr, lg, og)
	wh := handler.NewWebHookHandler(lg, du)
	e.POST("/linebot-webhook", wh.HandleLineEvent)
	e.Logger.Fatal(e.Start(":8080"))
}
