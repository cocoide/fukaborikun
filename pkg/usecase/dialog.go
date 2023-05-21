package usecase

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/cocoide/fukaborikun/conf"
	"github.com/cocoide/fukaborikun/pkg/gateway"
	"github.com/cocoide/fukaborikun/pkg/repository"
	"github.com/cocoide/fukaborikun/utils"
	"github.com/line/line-bot-sdk-go/linebot"
)

type DialogUseCase interface {
	BeginNewDialog(uid string, e *linebot.Event) error
	SubscribeDialogEvent(uid, text string, e *linebot.Event) error
}

type dialogUseCase struct {
	cr repository.CacheRepo
	lg gateway.LineAPIGateway
	og gateway.OpenAIGateway
}

func NewDialogUseCase(cr repository.CacheRepo, lg gateway.LineAPIGateway, og gateway.OpenAIGateway) DialogUseCase {
	return &dialogUseCase{cr: cr, lg: lg, og: og}
}

func (u *dialogUseCase) BeginNewDialog(uid string, e *linebot.Event) error {
	ctx := context.Background()
	// cacheのリセット
	u.cr.Delete(ctx, uid+"."+"question")
	u.cr.Delete(ctx, uid+"."+"answer")
	u.lg.PushTextMessage("深掘りしたいお題を入力してね\n(例: 強みと弱み、大切な価値観、etc)", e)
	err := u.cr.Set(ctx, uid, conf.WaitingTopics, time.Minute*30)
	if err != nil {
		return err
	}
	return nil
}

func (u *dialogUseCase) SubscribeDialogEvent(uid, text string, e *linebot.Event) error {
	ctx := context.Background()
	value, _ := u.cr.Get(ctx, uid)
	if value == "" {
		u.lg.PushTextMessage("『深掘りを開始』と入力したら始まります。", e)
		return nil
	}
	switch value {
	case conf.WaitingTopics:
		u.cr.Set(ctx, uid+"."+"topics", text, time.Hour)
		promp := fmt.Sprintf("『%s』をテーマに相手に質問する内容を5点、文脈のつながりを意識した上で以下のような形式で箇条書きするだけして。\n-\n-\n-\n-\n-\n(ただし箇条書き以外何も話さないで)", text)
		questionCh := make(chan string, 1)
		go func() {
			answer, _ := u.og.GetAnswerFromQuery(promp)
			questionCh <- answer
		}()
		u.lg.PushTextMessage("質問項目を考え中...", e)
		question := <-questionCh
		u.cr.Set(ctx, uid+"."+"question", question, time.Hour)
		questions := utils.ExtractTextLines(question)
		u.lg.PushTextMessage(questions[0], e)
		u.cr.Set(ctx, uid, "1", time.Hour)
	case "1":
		u.setQuestionAndNextState(uid, 1, e, text)
	case "2":
		u.setQuestionAndNextState(uid, 2, e, text)
	case "3":
		u.setQuestionAndNextState(uid, 3, e, text)
	case "4":
		u.setQuestionAndNextState(uid, 4, e, text)
	case "5":
		answer, _ := u.cr.Get(ctx, uid+"."+"answer")
		question, _ := u.cr.Get(ctx, uid+"."+"question")
		questions := utils.ExtractTextLines(question)
		u.cr.Set(ctx, uid+"."+"answer", answer+"\n"+"[質問: "+questions[4]+"]"+" 回答:"+text, time.Hour)
		u.lg.PushTextMessage("ご回答ありがとうございました。\n最後に深掘りの結果をまとめます。", e)

		summaryCh := make(chan string, 1)
		topics, _ := u.cr.Get(ctx, uid+"."+"topics")
		go func() {
			answer, _ := u.cr.Get(ctx, uid+"."+"answer")
			promp := fmt.Sprintf("【%s】 以下全ての回答を前後の文脈を自然な流れで繋げて、かつ誰かに伝えるような口調で分かりやすく%s", topics, answer)
			response, _ := u.og.GetAnswerFromQuery(promp)
			summaryCh <- response
		}()
		u.lg.PushTextMessage("回答を要約中...", e)
		summary := <-summaryCh
		u.lg.PushTextMessage(summary, e)
		// dialogを保存するのかどうかを聞く
		u.cr.Delete(ctx, uid+"."+"question")
		u.cr.Delete(ctx, uid)
	}
	return nil
}

func (u *dialogUseCase) setQuestionAndNextState(uid string, index int, e *linebot.Event, text string) {
	ctx := context.Background()
	question, _ := u.cr.Get(ctx, uid+"."+"question")
	questions := utils.ExtractTextLines(question)

	answer, _ := u.cr.Get(ctx, uid+"."+"answer")
	u.cr.Set(ctx, uid+"."+"answer", answer+"\n"+"[質問: "+questions[index-1]+"]"+" 回答: "+text, time.Hour)
	u.lg.PushTextMessage(questions[index], e)
	u.cr.Set(ctx, uid, strconv.Itoa(index+1), time.Hour)
}
