package conf

type SessionKey struct {
	UID string
}

func (u *SessionKey) TopicsKey() string {
	return u.UID + "." + "topics"
}

func (u *SessionKey) AnswerKey() string {
	return u.UID + "." + "answer"
}

func (u *SessionKey) QuestionsKey() string {
	return u.UID + "." + "questions"
}

func (u *SessionKey) UIDKey() string {
	return u.UID
}
