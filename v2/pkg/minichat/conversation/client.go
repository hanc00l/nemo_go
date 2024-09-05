package conversation

import (
	"encoding/base64"
	"encoding/json"
	"github.com/hanc00l/nemo_go/v2/pkg/minichat/constant"
	"github.com/hanc00l/nemo_go/v2/pkg/minichat/util"
	"log"
	"strings"
)

func (c *Client) Read() {
	defer func() {
		Manager.unregister <- c
	}()

	for {
		message, err := util.SocketReceive(c.Conn)
		if err != nil {
			return
		}
		if c.SecKey != "" {
			message, _ = base64.StdEncoding.DecodeString(string(message))
			message = util.AesDecryptCBC(message, []byte(c.SecKey))
		}
		if strings.HasPrefix(string(message), "img:") {
			Manager.broadcast <- Message{
				UserName:   c.UserName,
				RoomNumber: c.RoomNumber,
				Payload:    strings.ReplaceAll(string(message), "img:", ""),
				Cmd:        constant.CmdImg,
				HeadUlr:    c.HeadImg,
			}
		} else if strings.HasPrefix(string(message), "file:") {
			filenames := strings.Split(string(message), "||||")
			if len(filenames) == 2 {
				filename := strings.ReplaceAll(filenames[0], "file:", "")
				Manager.broadcast <- Message{
					UserName:   c.UserName,
					RoomNumber: c.RoomNumber,
					Payload:    filenames[1],
					Cmd:        constant.CmdFile,
					FileName:   filename,
					HeadUlr:    c.HeadImg,
				}
			}

		} else {
			Manager.broadcast <- Message{
				UserName:   c.UserName,
				RoomNumber: c.RoomNumber,
				Payload:    string(message),
				Cmd:        constant.CmdChat,
				HeadUlr:    c.HeadImg,
			}
		}
	}
}

func (c *Client) Write() {
	for {
		select {
		case message, isOpen := <-c.Send:
			if !isOpen {
				log.Printf("chan is closed")
				return
			}

			byteData, err := json.Marshal(message)
			if err != nil {
				log.Printf("json marshal error, error is %+v", err)
			} else {
				if c.SecKey != "" {
					byteData = util.AesEncryptCBC(byteData, []byte(c.SecKey))
					byteData = []byte(base64.StdEncoding.EncodeToString(byteData))
				}

				err = util.SocketSend(c.Conn, byteData)
				if err != nil {
					log.Printf("ocket send error, error is %+v", err)
					return
				}
			}
		case makeStop := <-c.Stop:
			if makeStop {
				break
			}
		}
	}
}
