package server

type ResponseData struct {
	Code      int    `json:"code"`
	Info      string `json:"info"`
	Data      any    `json:"data"`
	PublicKey string `json:"public_key"`
}

const (
	SuccessCode             = 10000
	ErrorCodePassword       = 20001
	ErrorCodeOnceToken      = 20002
	ErrorCodeUsernameRepeat = 20003
)

type PreCheckParam struct {
	RoomNumber string `json:"room_number"`
	UserName   string `json:"username"`
	Password   string `json:"password"`
}
