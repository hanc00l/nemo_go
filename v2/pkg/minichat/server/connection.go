package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/hanc00l/nemo_go/v2/pkg/minichat/config"
	"github.com/hanc00l/nemo_go/v2/pkg/minichat/constant"
	"github.com/hanc00l/nemo_go/v2/pkg/minichat/conversation"
	"github.com/hanc00l/nemo_go/v2/pkg/minichat/util"
	"html/template"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		//CheckOrigin: func(r *http.Request) bool {
		//	return r.Header.Get("Origin") == ""
		//},
	}
)

var OnceTokenMap = make(map[string]map[string]string)

func PreCheck(w http.ResponseWriter, r *http.Request) {
	//query := r.URL.Query()
	//roomNumber := query.Get("room_number")
	//userName := query.Get("username")
	//password := query.Get("password")

	// 确保请求方法是POST
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}

	// 读取请求体
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("body close error: %+v", err)
		}
	}(r.Body)

	// 解析JSON数据到结构体中
	var requestBody PreCheckParam
	err = json.Unmarshal(body, &requestBody)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	roomNumber := requestBody.RoomNumber
	userName := requestBody.UserName
	password := requestBody.Password

	log.Printf("PreCheck Info: RoomNumber is %s, UserName is %s, Password is %s\n", roomNumber, userName, password)

	if roomNumber == "" || roomNumber == "null" || roomNumber == "undefined" {
		log.Println("roomNumber is invalid")
		http.Error(w, "roomNumber is invalid", http.StatusBadRequest)
		return
	}

	if room, ok := conversation.Manager.Rooms[roomNumber]; ok {
		// check password
		if room.Password != password { // password error
			log.Println("password is invalid")
			res, err := json.Marshal(ResponseData{
				Code: ErrorCodePassword,
				Info: "password is invalid",
			})
			if err != nil {
				log.Printf("json marshal error is: %+v ", err)
				return
			}
			_, err = w.Write(res)
			if err != nil {
				log.Printf("response write error is: %+v ", err)
				return
			}
			return
		}

		// check name repeat
		for client, _ := range conversation.Manager.Rooms[roomNumber].Clients {
			if client.UserName == userName {
				//randomStr := util.RandomString(10)
				//userName = userName + randomStr
				log.Println("username repeat")
				res, err := json.Marshal(ResponseData{
					Code: ErrorCodeUsernameRepeat,
					Info: "username repeat",
				})
				if err != nil {
					log.Printf("json marshal error is: %+v ", err)
					return
				}
				_, err = w.Write(res)
				if err != nil {
					log.Printf("response write error is: %+v ", err)
					return
				}
				return
			}
		}
		for halfUsername, _ := range OnceTokenMap[roomNumber] {
			if halfUsername == userName {
				log.Println("username repeat")
				res, err := json.Marshal(ResponseData{
					Code: ErrorCodeUsernameRepeat,
					Info: "username repeat",
				})
				if err != nil {
					log.Printf("json marshal error is: %+v ", err)
					return
				}
				_, err = w.Write(res)
				if err != nil {
					log.Printf("response write error is: %+v ", err)
					return
				}
				return
			}
		}
	}

	if _, ok := OnceTokenMap[roomNumber]; !ok {
		OnceTokenMap[roomNumber] = make(map[string]string)
	}
	onceToken := util.RandomString(12)
	OnceTokenMap[roomNumber][userName] = onceToken
	res, err := json.Marshal(ResponseData{
		Code:      SuccessCode,
		Info:      "success",
		Data:      onceToken,
		PublicKey: config.RsaPublicKey,
	})
	if err != nil {
		log.Printf("json marshal error is: %+v ", err)
		return
	}
	_, err = w.Write(res)
	if err != nil {
		log.Printf("response write error is: %+v ", err)
		return
	}

}

func HandleWs(w http.ResponseWriter, r *http.Request) {

	// params := mux.Vars(r)
	// roomNumber := params["room_number"]
	// userName := params["username"]
	// password := params["password"]
	// cmd := params["cmd"]

	query := r.URL.Query()
	roomNumber := query.Get("room_number")
	userName := query.Get("username")
	onceToken := query.Get("once_token")
	password := query.Get("password")
	cmd := query.Get("cmd")
	secKey := query.Get("sec_key")
	headImg := query.Get("head_img")
	log.Printf("Connection Info: RoomNumber is %s, Cmd is %s, UserName is %s, OnceToken is %s SecKey is %s\n", roomNumber, cmd, userName, onceToken, secKey)

	if roomNumber == "" || roomNumber == "null" || roomNumber == "undefined" || len(roomNumber) < 1 {
		log.Println("roomNumber is invalid")
		http.Error(w, "roomNumber is invalid", http.StatusBadRequest)
		return
	}

	if _, ok := conversation.Manager.Rooms[roomNumber]; ok {
		if cacheOnceToken, onceOk := OnceTokenMap[roomNumber][userName]; !onceOk || cacheOnceToken != onceToken { // password error
			log.Println("once token is invalid")
			res, err := json.Marshal(ResponseData{
				Code: ErrorCodeOnceToken,
				Info: "once token is invalid",
			})
			if err != nil {
				log.Printf("json marshal error is: %+v ", err)
				return
			}
			_, err = w.Write(res)
			if err != nil {
				log.Printf("response write error is: %+v ", err)
				return
			}
			return
		}
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	delete(OnceTokenMap[roomNumber], userName)

	// timeout
	timeoutDuration := 1440 * time.Minute
	err = ws.SetReadDeadline(time.Now().Add(timeoutDuration))
	if err != nil {
		log.Printf("SetReadDeadline error: %+v\n", err)
		return
	}

	// heart
	go func() {
		heartCount := 0
		for {
			time.Sleep(30 * time.Second)
			if err := ws.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second)); err != nil {
				//if err := ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
			heartCount++
			if heartCount >= 10 {
				log.Println("websocket heartbeat!")
				heartCount = 0
			}
		}
	}()
	bkey, err := base64.StdEncoding.DecodeString(secKey)
	sec_key := ""
	if bkey != nil {
		plaintext, err := util.RSADecryptFromPemText(bkey, []byte(config.RsaPrivateKey))
		if err == nil {
			sec_key = string(plaintext)
		}
	}

	head_img := ""
	if strings.HasPrefix(headImg, constant.HeadImageHttpPath) && (strings.HasSuffix(headImg, ".png") || strings.HasSuffix(headImg, ".jpg") || strings.HasSuffix(headImg, ".gif") || strings.HasSuffix(headImg, ".jpeg")) {
		head_img = headImg
	}
	// register new client
	client := &conversation.Client{RoomNumber: roomNumber, UserName: userName, Password: password, Conn: ws, Send: make(chan conversation.Message), SecKey: sec_key, HeadImg: head_img}

	go client.Write()
	go client.Read()

	conversation.Manager.Register <- client

}

func HandleFiles(w http.ResponseWriter, _ *http.Request, dirTemplate fs.FS) {
	data := struct {
		Url string
	}{
		Url: "",
	}

	tmpl, err := template.ParseFS(dirTemplate, "templates/index.html")
	if err != nil {
		fmt.Printf("failed to parse the template: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	err = tmpl.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		fmt.Printf("failed to execute the template: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
