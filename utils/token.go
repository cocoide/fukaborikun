package utils

import "github.com/google/uuid"

// 36文字の固定長のIDを生成する
func GenerateToken() (string, error) {
	uuidObj, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	// UUIDを文字列に変換
	token := uuidObj.String()
	return token, nil
}
