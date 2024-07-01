package conversation

import (
	"github.com/gorilla/websocket"
	"github.com/hanc00l/nemo_go/pkg/minichat/config"
	"github.com/hanc00l/nemo_go/pkg/minichat/constant"
	"github.com/hanc00l/nemo_go/pkg/minichat/util"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type ConversationManager struct {
	Rooms          map[string]*Room
	Register       chan *Client
	unregister     chan *Client
	broadcast      chan Message
	registerLock   *sync.RWMutex
	unregisterLock *sync.RWMutex
	broadcastLock  *sync.RWMutex
}

type Message struct {
	RoomNumber string `json:"room_number"`
	UserName   string `json:"username"`
	Cmd        string `json:"cmd"`
	Payload    string `json:"payload"`
	HeadUlr    string `json:"head_ulr"`
	FileName   string `json:"file_name"`
	//SecKey     string `json:"sec_key"`
}

type Client struct {
	//cmd    string
	RoomNumber string
	UserName   string
	Password   string
	Send       chan Message
	Conn       *websocket.Conn
	Stop       chan bool
	//用户头像
	HeadImg string
	//每个用户一个密钥
	SecKey string
}

type Room struct {
	Clients  map[*Client]*Client
	RoomName string
	Password string
	Messages []*Message
	//PrivateKey string
	//PublicKey  string
}

var Manager = ConversationManager{
	broadcast:      make(chan Message),
	Register:       make(chan *Client),
	unregister:     make(chan *Client),
	Rooms:          make(map[string]*Room),
	registerLock:   new(sync.RWMutex),
	unregisterLock: new(sync.RWMutex),
	broadcastLock:  new(sync.RWMutex),
}

func (manager *ConversationManager) Start() {
	for {
		select {
		case client := <-manager.Register:
			// 新客户端链接
			manager.registerLock.Lock()
			if _, ok := manager.Rooms[client.RoomNumber]; !ok {
				//由于交互逻辑问题，目前还是全局使用一个privateKey，不能对每个房间建立key
				//err, publicKey, privateKey := util.GenerateRSAKey(2048)
				//if err != nil {
				//	log.Println(err.Error())
				//}
				manager.Rooms[client.RoomNumber] = &Room{
					Clients:  make(map[*Client]*Client),
					Password: client.Password,
					//PrivateKey: string(privateKey),
					//PublicKey:  string(publicKey),
				}
			}
			// 加载历史消息
			if config.LoadHistory {
				lenmsg := len(manager.Rooms[client.RoomNumber].Messages)
				indx := 0
				if config.MaxHistoryMessage > 0 && lenmsg > config.MaxHistoryMessage {
					indx = lenmsg - config.MaxHistoryMessage
				}
				for i := indx; i < lenmsg; i++ {
					client.Send <- *manager.Rooms[client.RoomNumber].Messages[i]
				}
				//go func() {
				//	bkmsg := manager.Rooms[client.RoomNumber].Messages
				//	for _, msg := range bkmsg {
				//		if msg != nil {
				//			client.Send <- *msg
				//		}
				//	}
				//}()
			}
			// 塞入房间初次数据
			manager.Rooms[client.RoomNumber].Clients[client] = client
			go func() {
				names := ""
				for key, _ := range manager.Rooms[client.RoomNumber].Clients {
					names += "[ " + key.UserName + " ], "
				}
				names = strings.TrimSuffix(names, ", ")
				manager.broadcast <- Message{
					UserName:   client.UserName,
					Payload:    constant.JoinSuccess + constant.Online + names,
					RoomNumber: client.RoomNumber,
					Cmd:        constant.CmdJoin,
					//SecKey:     manager.Rooms[client.RoomNumber].PublicKey,
				}

			}()
			manager.registerLock.Unlock()

		case client := <-manager.unregister:
			// 客户端断开链接
			manager.unregisterLock.Lock()
			err := client.Conn.Close()
			if err != nil {
				log.Println(err.Error())
			}
			if _, ok := manager.Rooms[client.RoomNumber]; ok {
				delete(manager.Rooms[client.RoomNumber].Clients, client)
				if len(manager.Rooms[client.RoomNumber].Clients) == 0 {
					delete(manager.Rooms, client.RoomNumber)
					if !config.IsNotDelFileDir {
						os.RemoveAll(filepath.Join(constant.UploadSavePath, util.Md5Crypt(client.RoomNumber)))
					}
				}
				//client.stop <- true
				safeClose(client.Send)

				if manager.Rooms != nil && len(manager.Rooms) != 0 && manager.Rooms[client.RoomNumber] != nil && client.RoomNumber != "" {
					for c, _ := range manager.Rooms[client.RoomNumber].Clients {
						names := ""
						for key, _ := range manager.Rooms[client.RoomNumber].Clients {
							names += "[ " + key.UserName + " ], "
						}
						names = strings.TrimSuffix(names, ", ")
						c.Send <- Message{
							UserName:   client.UserName,
							Payload:    "[ " + client.UserName + " ] " + constant.ExitSuccess + constant.Online + names,
							RoomNumber: client.RoomNumber,
							Cmd:        constant.CmdExit,
						}
					}
				}
			}
			manager.unregisterLock.Unlock()

		case message := <-manager.broadcast:
			// 存储历史消息
			if config.LoadHistory {
				manager.Rooms[message.RoomNumber].Messages = append(manager.Rooms[message.RoomNumber].Messages, &message)
				// 超过最大历史消息数，删除前半部分
				if config.MaxHistoryMessage > 0 && len(manager.Rooms[message.RoomNumber].Messages) > 2*config.MaxHistoryMessage {
					manager.Rooms[message.RoomNumber].Messages = manager.Rooms[message.RoomNumber].Messages[config.MaxHistoryMessage:]
				}
			}
			// 广播消息
			manager.broadcastLock.RLock()
			if manager.Rooms != nil && len(manager.Rooms) != 0 && manager.Rooms[message.RoomNumber] != nil && message.RoomNumber != "" {
				for c, _ := range manager.Rooms[message.RoomNumber].Clients {
					if c != nil && c.Conn != nil && c.Send != nil {
						c.Send <- message
					}
				}
			}
			manager.broadcastLock.RUnlock()
		}

	}
}

func safeClose(ch chan Message) {
	defer func() {
		if recover() != nil {
			log.Println("ch is closed")
		}
	}()
	close(ch)
	log.Println("ch closed successfully")
}
