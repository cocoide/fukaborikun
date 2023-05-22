package usecase

import (
	"context"
	"fmt"
	"strconv"

	v "github.com/cocoide/fukaborikun/conf"
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
	key := v.SessionKey{UID: uid}
	ctx := context.Background()
	// cacheのリセット
	u.cr.Delete(ctx, key.QuestionsKey())
	u.cr.Delete(ctx, key.AnswerKey())
	u.lg.PushText("深掘りしたいお題を入力してね\n(例: 強みと弱み、大切な価値観、etc)", e)
	err := u.cr.Set(ctx, uid, v.WaitingForTopics, v.DialogExpireTime)
	if err != nil {
		return err
	}
	return nil
}

func (u *dialogUseCase) SubscribeDialogEvent(uid, text string, e *linebot.Event) error {
	key := v.SessionKey{UID: uid}

	ctx := context.Background()
	state, _ := u.cr.Get(ctx, key.UIDKey())
	if state == "" {
		u.lg.PushText("『深掘りを開始』と入力したら始まります。", e)
		return nil
	}
	switch state {
	case v.ThinkingQuestions:
		u.lg.PushText("質問を生成中...", e)
	case v.ThinkingSummary:
		u.lg.PushText("回答を要約中....", e)
	case v.WaitingForTopics:
		u.cr.Set(ctx, key.TopicsKey(), text, v.DialogExpireTime)
		prompt := fmt.Sprintf("『%s』をテーマに相手に質問する内容を5点、文脈のつながりを意識した上で以下のような形式で箇条書きするだけして。\n-\n-\n-\n-\n-\n(ただし箇条書き以外何も話さないで)", text)
		questionCh := make(chan string, 1)
		go func() {
			answer, _ := u.og.GetAnswerFromPrompt(prompt)
			questionCh <- answer
		}()
		u.cr.Set(ctx, uid, v.ThinkingQuestions, v.DialogExpireTime)
		// 非同期処理が完了するまで、valueに『ThinkingQuestions』を設定
		u.lg.PushText("質問項目を考え中...", e)
		question := <-questionCh
		u.cr.Set(ctx, key.QuestionsKey(), question, v.DialogExpireTime)
		questions := utils.ExtractTextLines(question)
		u.lg.PushText(questions[0], e)
		u.cr.Set(ctx, key.UIDKey(), "1", v.DialogExpireTime)
	case "1", "2", "3", "4":
		index, err := strconv.Atoi(state)
		if err != nil {
			return err
		}
		if err := u.manageSingleDialog(uid, index, e, text); err != nil {
			return err
		}
	case "5":
		answer, _ := u.cr.Get(ctx, key.AnswerKey())
		question, _ := u.cr.Get(ctx, key.QuestionsKey())
		questions := utils.ExtractTextLines(question)
		u.cr.Set(ctx, key.AnswerKey(), answer+"\n"+"[質問: "+questions[4]+"]"+" 回答:"+text, v.DialogExpireTime)
		u.lg.PushText("ご回答ありがとうございました。\n最後に深掘りの結果をまとめます。", e)
		u.cr.Set(ctx, uid, v.ThinkingSummary, v.DialogExpireTime)
		summaryCh := make(chan string, 1)
		topics, _ := u.cr.Get(ctx, key.TopicsKey())
		go func() {
			answer, _ := u.cr.Get(ctx, key.AnswerKey())
			prompt := fmt.Sprintf("[お題: %s][ 対話%s]以上、すべての回答を前後の文脈を自然な流れで繋げて、分かりやすく言い換えて", topics, answer)
			response, _ := u.og.GetAnswerFromPrompt(prompt)
			summaryCh <- response
		}()
		u.lg.PushText("回答を要約中...", e)

		summary := <-summaryCh
		u.lg.PushText("【"+topics+"】\n"+summary, e)
		// dialogを保存するのかどうかを聞く処理を入れる
		// u.cr.Delete(ctx, key.QuestionsKey())
		u.cr.Delete(ctx, uid)
	}
	return nil
}

func (u *dialogUseCase) manageSingleDialog(uid string, index int, e *linebot.Event, text string) error {
	key := v.SessionKey{UID: uid}
	ctx := context.Background()
	question, err := u.cr.Get(ctx, key.AnswerKey())
	if err != nil {
		return err
	}
	questions := utils.ExtractTextLines(question)
	answer, err := u.cr.Get(ctx, key.AnswerKey())
	if err != nil {
		return err
	}
	u.cr.Set(ctx, key.AnswerKey(), answer+"\n"+"[質問: "+questions[index-1]+"]"+"[ 回答: "+text+"]", v.DialogExpireTime)
	u.lg.PushText(questions[index], e)
	u.cr.Set(ctx, key.UIDKey(), strconv.Itoa(index+1), v.DialogExpireTime)
	return nil
}
