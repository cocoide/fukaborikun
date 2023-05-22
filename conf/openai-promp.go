package conf

type OpenAIPrompt struct {
}

const (
	QuestionFormTopicsPrompt = "『%s』をテーマに相手に質問する内容を5点、文脈のつながりを意識した上で以下のような形式で箇条書きするだけして。\n-\n-\n-\n-\n-\n(ただし箇条書き以外何も話さないで)"
	SummaryDialogPrompt      = "[お題: %s][ 対話%s]以上、すべての回答を前後の文脈を自然な流れで繋げて、分かりやすく言い換えて"
	BrushupSummaryPrompt     = "[%s]この文章をさらに論理的で具体的に富んで読みやすく書き換えて"
)
