package gateway_test

import (
	"context"
	"testing"

	"github.com/cocoide/fukaborikun/conf"
	"github.com/cocoide/fukaborikun/pkg/gateway"
)

func TestGetAnswerFromQuery(t *testing.T) {
	conf.NewEnv()
	ctx := context.Background()
	og := gateway.NewOpenAIGateway(ctx)
	res, err := og.GetAnswerFromPrompt("hello")
	if err != nil {
		t.Errorf("func error: %v", err)
	}
	t.Logf("answer: %s", res)
}
