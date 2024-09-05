package controllers

import (
	"github.com/google/uuid"
	minichatConfig "github.com/hanc00l/nemo_go/v2/pkg/minichat/config"
	"github.com/hanc00l/nemo_go/v2/pkg/minichat/server"
)

type MiniChatController struct {
	BaseController
}

var GlobalRoomNumber = make(map[string]string)

// IndexAction 显示列表页面
func (c *MiniChatController) IndexAction() {
	roomNumber := c.GetString("room", "")
	if minichatConfig.EnableAnonymous {
		//允许匿名访问，但房间号只能uuid生成提高安全性
		if roomNumber == "" {
			roomNumber = uuid.New().String()
			GlobalRoomNumber[roomNumber] = roomNumber
		} else {
			if !checkRoomNumber(roomNumber) {
				c.Abort("404")
			}
		}
		c.Data["roomNumber"] = roomNumber
	} else {
		// 根据当前工作区映射到房间号
		// 强制检查房间号是否是由服务端生成并对应
		if roomNumber != "" {
			if !checkRoomNumber(roomNumber) {
				c.Abort("404")
			}
			c.Data["roomNumber"] = roomNumber
		} else {
			workspaceGUID := c.GetCurrentWorkspaceGUID()
			if workspaceGUID != "" {
				if _, ok := GlobalRoomNumber[workspaceGUID]; !ok {
					GlobalRoomNumber[workspaceGUID] = uuid.New().String()
				}
				c.Data["roomNumber"] = GlobalRoomNumber[workspaceGUID]
			} else {
				c.Data["roomNumber"] = ""
			}
		}
	}
	c.TplName = "minichat-index.html"
}

func (c *MiniChatController) PreCheckAction() {
	//无法关闭自动渲染，所以传递一个空模板给beego
	c.TplName = "minichat-null.html"
	server.PreCheck(c.Ctx.ResponseWriter, c.Ctx.Request)
}

func (c *MiniChatController) WebSocketAction() {
	//无法关闭自动渲染，所以传递一个空模板给beego
	c.TplName = "minichat-null.html"
	server.HandleWs(c.Ctx.ResponseWriter, c.Ctx.Request)
}

func (c *MiniChatController) UploadAction() {
	//无法关闭自动渲染，所以传递一个空模板给beego
	c.TplName = "minichat-null.html"
	server.Upload(c.Ctx.ResponseWriter, c.Ctx.Request)
}

// 检查房间号是否存在
func checkRoomNumber(roomNumber string) bool {
	for _, v := range GlobalRoomNumber {
		if v == roomNumber {
			return true
		}
	}
	return false
}
