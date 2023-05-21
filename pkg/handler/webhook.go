package handler

import (
	"context"
	"log"
	"time"

	"github.com/cocoide/fukaborikun/pkg/gateway"
	"github.com/cocoide/fukaborikun/pkg/usecase"
	"github.com/cocoide/fukaborikun/utils"
	"github.com/labstack/echo/v4"
	"github.com/line/line-bot-sdk-go/linebot"
)

type WebHookHandler interface {
	HandleLineEvent(c echo.Context) error
}

type webHookHandler struct {
	lg gateway.LineAPIGateway
	du usecase.DialogUseCase
}

func NewWebHookHandler(lg gateway.LineAPIGateway, du usecase.DialogUseCase) WebHookHandler {
	return &webHookHandler{lg: lg, du: du}
}

var limiter = utils.NewRateLimiter(5, 10)

func (wh *webHookHandler) HandleLineEvent(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := limiter.Limit(ctx, func() error {
		events, _ := wh.lg.SubscribeToEvents(c.Request())
		for _, e := range events {
			if e.Type != linebot.EventTypeMessage {
				continue
			}
			uid := e.Source.UserID
			switch msg := e.Message.(type) {
			case *linebot.TextMessage:
				text := msg.Text
				switch text {
				case "深掘りを開始":
					if err := wh.du.BeginNewDialog(uid, e); err != nil {
						wh.lg.ReturnWithError("", e, err)
					}
				default:
					if err := wh.du.SubscribeDialogEvent(uid, text, e); err != nil {
						wh.lg.ReturnWithError("", e, err)
					}
				}
			}
		}
		return nil
	}); err != nil {
		log.Print(err.Error())
	}
	return nil
}
