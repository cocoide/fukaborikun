package gateway

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/line/line-bot-sdk-go/linebot"
)

type LineAPIGateway interface {
	SubscribeLineBotEvents(req *http.Request) ([]*linebot.Event, error)
	ReplyWithText(msg string, event *linebot.Event) error
	ReplyWithMessage(msg *linebot.TextMessage, event *linebot.Event) error
	PushText(msg string, event *linebot.Event) error
	ReturnWithError(msg string, event *linebot.Event, err error)
	ReplyQuickReplyButtons(text string, labels []string, event *linebot.Event) error
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

func (gateway *lineAPIGateway) SubscribeLineBotEvents(req *http.Request) ([]*linebot.Event, error) {
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

func (gateway *lineAPIGateway) PushText(msg string, event *linebot.Event) error {
	if _, err := gateway.bot.PushMessage(event.Source.UserID, linebot.NewTextMessage(msg)).Do(); err != nil {
		return fmt.Errorf("failed to push text message(%s): %v", msg, err)
	}
	return nil
}

func (gateway *lineAPIGateway) ReturnWithError(msg string, event *linebot.Event, err error) {
	errMsg := "エラーが発生しました。もう一度お試し下さい"
	if msg == "" {
		log.Printf("unexpected error occured: %v", err)
		msg = errMsg
	}
	if _, err := gateway.bot.PushMessage(event.Source.UserID, linebot.NewTextMessage(msg)).Do(); err != nil {
		log.Printf("failed to handle error: %v", err)
	}
	return
}

func (gateway *lineAPIGateway) ReplyQuickReplyButtons(text string, labels []string, event *linebot.Event) error {
	msg := linebot.NewTextMessage(text)
	var quickReplyButtons []*linebot.QuickReplyButton
	for _, label := range labels {
		button := linebot.NewQuickReplyButton("", linebot.NewMessageAction(label, label))
		quickReplyButtons = append(quickReplyButtons, button)
	}
	msg.WithQuickReplies(linebot.NewQuickReplyItems(quickReplyButtons...))
	if err := gateway.ReplyWithMessage(msg, event); err != nil {
		return err
	}
	return nil
}
