// 对话管理

package controller

import (
	"git.jsjit.cn/customerService/customerService_Core/logic"
	"git.jsjit.cn/customerService/customerService_Core/wechat"
	"github.com/gin-gonic/gin"
	"log"
)

type DialogController struct {
	wxContext *wechat.Wechat
	rooms     map[string]*logic.Room
}

func InitDialog(wxContext *wechat.Wechat, rooms map[string]*logic.Room) *DialogController {
	return &DialogController{wxContext, rooms}
}

// @Summary 获取待回复消息列表
// @Description 获取待回复消息列表
// @Tags Dialog
// @Accept  json
// @Produce  json
// @Success 200 {string} json ""
// @Router /v1/dialog/list [get]
func (c *DialogController) List(context *gin.Context) {
	customerId := context.GetInt("customerId")

	for _, wxReceive := range c.rooms {
		log.Println(customerId, wxReceive)
	}

}

// @Summary 客服接入用户
// @Description 客服接入用户，创建一个会话房间
// @Tags Dialog
// @Accept  json
// @Produce  json
// @Success 200 {string} json ""
// @Router /v1/dialog/create [post]
func (c *DialogController) Create(context *gin.Context) {
}

// @Summary 获取一个用户的聊天记录
// @Description 获取一个用户的聊天记录
// @Tags Dialog
// @Accept  json
// @Produce  json
// @Param id path int true "客户 ID"
// @Success 200 {string} json ""
// @Router /v1/dialog/customer/{id}/message [get]
func (c *DialogController) History(context *gin.Context) {
}

// @Summary 客服发送消息给客户
// @Description 客服发送消息给客户
// @Tags Dialog
// @Accept  json
// @Produce  json
// @Param id path int true "客户 ID"
// @Success 200 {string} json ""
// @Router /v1/dialog/customer/{id}/message [post]
func (c *DialogController) SendMessage(context *gin.Context) {
}

// @Summary 客服撤回发送的消息
// @Description 客服撤回发送的消息
// @Tags Dialog
// @Accept  json
// @Produce  json
// @Param id path int true "客户 ID"
// @Success 200 {string} json ""
// @Router /v1/dialog/customer/{id}/message [delete]
func (c *DialogController) RecallMessage(context *gin.Context) {
}
