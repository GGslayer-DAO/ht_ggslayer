package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	base64Table = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
)

var coder = base64.NewEncoding(base64Table)

//Base64Encode base64加密
func Base64Encode(src []byte) []byte {
	return []byte(coder.EncodeToString(src))
}

//Base64Decode base64解密
func Base64Decode(src []byte) ([]byte, error) {
	return coder.DecodeString(string(src))
}

func MetaEncode(params map[string]interface{}) string {
	stb, _ := json.Marshal(params)
	enstr := fmt.Sprintf("%s%s%s", "yfi6n9", string(Base64Encode(stb)), "c39lzpk")
	return string(Base64Encode([]byte(enstr)))
}

func MetaDecode(str string) string {
	bm, _ := Base64Decode([]byte(str))
	bs := strings.ReplaceAll(string(bm), "yfi6n9", "")
	bs = strings.ReplaceAll(string(bs), "c39lzpk", "")
	st, _ := Base64Decode([]byte(bs))
	return string(st)
}
