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
	u.cr.Delete(ctx, key.DialogStateKey())
	u.lg.PushText("深掘りしたいお題を入力してね\n(例: 強みと弱み、大切な価値観、etc)", e)
	err := u.cr.Set(ctx, key.DialogStateKey(), v.WaitingForTopics, v.DialogExpireTime)
	if err != nil {
		return err
	}
	return nil
}

func (u *dialogUseCase) SubscribeDialogEvent(uid, text string, e *linebot.Event) error {
	key := v.SessionKey{UID: uid}
	ctx := context.Background()
	state, _ := u.cr.Get(ctx, key.DialogStateKey())
	if state == "" {
		u.lg.PushText("『深掘りを開始』と入力したら始まります。", e)
		return nil
	}
	switch state {
	case v.ThinkingQuestions:
		u.lg.PushText("質問を生成中...", e)
	case v.ThinkingSummary:
		u.lg.PushText("要約を要約中....", e)
	case v.BrushupSummary:
		u.lg.PushText("要約を再生成中....", e)
	case v.WaitingForTopics:
		u.cr.Set(ctx, key.TopicsKey(), text, v.DialogExpireTime)
		prompt := fmt.Sprintf(v.QuestionFormTopicsPrompt, text)
		questionCh := u.og.AsyncGetAnswerFromPrompt(prompt)
		u.cr.Set(ctx, uid, v.ThinkingQuestions, v.DialogExpireTime)
		u.lg.PushText("質問を生成中...", e)
		// Questionの生成が完了するまでブロック
		question := <-questionCh
		u.cr.Set(ctx, key.QuestionsKey(), question, v.DialogExpireTime)
		questions := utils.ExtractTextLines(question)
		u.lg.PushText(questions[0], e)
		u.cr.Set(ctx, key.DialogStateKey(), "1", v.DialogExpireTime)
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
		cacheAllDialogValue := answer + "\n" + "[質問: " + questions[4] + "]" + " 回答:" + text
		u.cr.Set(ctx, key.AnswerKey(), cacheAllDialogValue, v.DialogExpireTime)
		u.lg.PushText("ご回答ありがとうございました。\n最後に深掘りの結果をまとめます。", e)
		u.cr.Set(ctx, key.DialogStateKey(), v.ThinkingSummary, v.DialogExpireTime)

		topics, _ := u.cr.Get(ctx, key.TopicsKey())

		prompt := fmt.Sprintf(v.SummaryDialogPrompt, topics, answer)
		summaryCh := u.og.AsyncGetAnswerFromPrompt(prompt)
		u.cr.Set(ctx, key.DialogStateKey(), v.ThinkingSummary, v.DialogExpireTime)
		u.lg.PushText("回答を要約中...", e)
		// Summaryの生成が完了するまでブロック
		summary := <-summaryCh
		u.lg.PushText("【"+topics+"】\n"+summary, e)
		// AnswerKeyに再生成用のSummaryを保存
		u.cr.Set(ctx, key.AnswerKey(), summary, v.DialogExpireTime)

		u.lg.SendQuickReplyButtons("作成した要約を再生成する？", []string{"お願いします", "大丈夫です"}, e)
		u.cr.Set(ctx, key.DialogStateKey(), "6", v.DialogExpireTime)
	case "6":
		if text == "大丈夫です" {
			u.lg.PushText("わかりました。それでは深掘りを終了いたします。\nもう一度遊びたいときは、再度『深掘りを開始』とお送り下さい", e)
		}
		if text == "お願いします" {
			summary, _ := u.cr.Get(ctx, key.AnswerKey())
			topics, _ := u.cr.Get(ctx, key.TopicsKey())
			prompt := fmt.Sprintf(v.BrushupSummaryPrompt, summary)
			brushupCh := u.og.AsyncGetAnswerFromPrompt(prompt)
			u.cr.Set(ctx, key.DialogStateKey(), v.BrushupSummary, v.DialogExpireTime)
			u.lg.PushText("要約を再生成中....", e)
			brushup := <-brushupCh
			u.lg.PushText("【"+topics+"】\n"+brushup, e)
			u.lg.PushText("以上で終了いたします。\nもう一度遊びたいときは、再度『深掘りを開始』とお送り下さい", e)
		}
		u.cr.Delete(ctx, key.QuestionsKey())
		u.cr.Delete(ctx, key.DialogStateKey())
		u.cr.Delete(ctx, key.AnswerKey())
		u.cr.Delete(ctx, key.TopicsKey())
	}
	return nil
}

func (u *dialogUseCase) manageSingleDialog(uid string, index int, e *linebot.Event, text string) error {
	key := v.SessionKey{UID: uid}
	ctx := context.Background()
	question, err := u.cr.Get(ctx, key.QuestionsKey())
	if err != nil {
		return err
	}
	questions := utils.ExtractTextLines(question)
	answer, err := u.cr.Get(ctx, key.AnswerKey())
	if err != nil {
		return err
	}
	cacheDialogValue := answer + "\n" + "[質問: " + questions[index-1] + "]" + "[ 回答: " + text + "]"
	u.cr.Set(ctx, key.AnswerKey(), cacheDialogValue, v.DialogExpireTime)
	u.lg.PushText(questions[index], e)

	nextIndex := strconv.Itoa(index + 1)
	u.cr.Set(ctx, key.DialogStateKey(), nextIndex, v.DialogExpireTime)
	return nil
}
