package service

type Output struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Result  interface{} `json:"result"`
}
