package handler

import (
	"github.com/cocoide/fukaborikun/pkg/gateway"
	"github.com/labstack/echo/v4"
	"github.com/line/line-bot-sdk-go/linebot"
)

type WebHookHandler interface {
	HandleLineEvent(c echo.Context) error
}

type webHookHandler struct {
	lg gateway.LineAPIGateway
}

func NewWebHookHandler(lg gateway.LineAPIGateway) WebHookHandler {
	return &webHookHandler{lg: lg}
}

func (h *webHookHandler) HandleLineEvent(c echo.Context) error {
	events, _ := h.lg.SubscribeToEvents(c.Request())
	for _, e := range events {
		if e.Type != linebot.EventTypeMessage {
			continue
		}
		switch msg := e.Message.(type) {
		case *linebot.TextMessage:
			switch msg.Text {
			case "こんにちは":
				h.lg.ReplyWithText("こんにちは", e)
			default:
				h.lg.ReplyWithText(msg.Text, e)
			}
		}
	}
	return nil
}
