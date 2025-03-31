// 工具函数
package utils

import (
	"crypto/rand"
	"encoding/base64"
	"math/big"
)

// GenerateRandomString 生成随机字符串
func GenerateRandomString(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GenerateRandomNumber 生成随机数字
func GenerateRandomNumber(min, max int) int {
	bigInt, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	if err != nil {
		return min
	}
	return min + int(bigInt.Int64())
}
