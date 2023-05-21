package gateway

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/line/line-bot-sdk-go/linebot"
)

type LineAPIGateway interface {
	SubscribeToEvents(req *http.Request) ([]*linebot.Event, error)
	ReplyWithText(msg string, event *linebot.Event) error
	ReplyWithMessage(msg *linebot.TextMessage, event *linebot.Event) error
	PushTextMessage(msg string, event *linebot.Event) error
	HandleErrorWithMessage(msg string, event *linebot.Event) error
	BroadcastMessage(msg *linebot.TextMessage, userIds []string) error
}

type lineAPIGateway struct {
	bot *linebot.Client
}

func NewLineAPIGateway() LineAPIGateway {
	channelSecret := os.Getenv("CHANNEL_SECRET")
	channelAccessToken := os.Getenv("CHANNEL_ACCESS_TOKEN")
	bot, err := linebot.New(channelSecret, channelAccessToken)
	if err != nil {
		log.Fatalf("failed to initialize Linebot client: %v", err)
	}
	return &lineAPIGateway{bot: bot}
}

func (gateway *lineAPIGateway) SubscribeToEvents(req *http.Request) ([]*linebot.Event, error) {
	events, err := gateway.bot.ParseRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LINEBOT request: %v", err)
	}
	return events, nil
}

func (gateway *lineAPIGateway) ReplyWithText(msg string, event *linebot.Event) error {
	if _, err := gateway.bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(msg)).Do(); err != nil {
		return fmt.Errorf("failed to reply with text message(%s): %v", msg, err)
	}
	return nil
}

func (gateway *lineAPIGateway) ReplyWithMessage(msg *linebot.TextMessage, event *linebot.Event) error {
	if _, err := gateway.bot.ReplyMessage(event.ReplyToken, msg).Do(); err != nil {
		return fmt.Errorf("failed to reply with Linebot message: %v", err)
	}
	return nil
}

func (gateway *lineAPIGateway) PushTextMessage(msg string, event *linebot.Event) error {
	if _, err := gateway.bot.PushMessage(event.Source.UserID, linebot.NewTextMessage(msg)).Do(); err != nil {
		return fmt.Errorf("failed to push text message(%s): %v", msg, err)
	}
	return nil
}

func (gateway *lineAPIGateway) HandleErrorWithMessage(msg string, event *linebot.Event) error {
	errMsg := "エラーが発生しました。もう一度お試し下さい"
	if msg == "" {
		msg = errMsg
	}
	if _, err := gateway.bot.PushMessage(event.Source.UserID, linebot.NewTextMessage(msg)).Do(); err != nil {
		return fmt.Errorf("failed to handle error with")
	}
	return nil
}

func (gateway *lineAPIGateway) BroadcastMessage(msg *linebot.TextMessage, userIds []string) error {
	for _, userID := range userIds {
		if _, err := gateway.bot.PushMessage(userID, msg).Do(); err != nil {
			return fmt.Errorf("failed to broadcast message: %v", err)
		}
	}
	return nil
}
