package controller

import (
	"fmt"
	"git.jsjit.cn/customerService/customerService_Core/common"
	"git.jsjit.cn/customerService/customerService_Core/handle"
	"git.jsjit.cn/customerService/customerService_Core/model"
	"git.jsjit.cn/customerService/customerService_Core/wechat"
	"git.jsjit.cn/customerService/customerService_Core/wechat/message"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
	"log"
	"time"
)

type WeiXinController struct {
	wxContext *wechat.Wechat
	aiModule  *handle.AiSemantic
}

func NewWeiXin(wxContext *wechat.Wechat, aiModule *handle.AiSemantic) *WeiXinController {
	return &WeiXinController{wxContext: wxContext, aiModule: aiModule}
}

// 微信通信接口
func (c *WeiXinController) Listen(context *gin.Context) {
	wcServer := c.wxContext.GetServer(context.Request, context.Writer)

	//设置接收消息的处理方法
	wcServer.SetMessageHandler(func(msg message.MixMessage) (reply *message.Reply) {
		/*
			A 24小时新接入客户：
				1. 注册分配聊天房间
				2. 存储新客户、离线留言信息数据
			B 已在线的客户：
				1. 检索已分配的房间
				2. 存储聊天数据
		*/
		var (
			msgType    = string(msg.MsgType) // 消息类型
			MediaUrl   = ""                  // 多媒体地址
			msgText    = ""                  // 文本内容
			aiDialogue = ""                  // AI答复
		)
		switch msg.MsgType {
		case message.MsgTypeText:
			msgText = message.NewText(msg.Content).Content
		case message.MsgTypeImage:
			MediaUrl = msg.PicURL
		case message.MsgTypeVoice:
			msgText = msg.Recognition
			material := c.wxContext.GetMaterial()
			if mediaURL, err := material.GetMediaURL(msg.MediaID); err != nil {
				log.Printf("material.MsgTypeVoice is err: %#v", err)
			} else {
				MediaUrl = mediaURL
			}
		case message.MsgTypeVideo:
			material := c.wxContext.GetMaterial()
			if mediaURL, err := material.GetMediaURL(msg.MediaID); err != nil {
				log.Printf("material.MsgTypeVideo is err: %#v", err)
			} else {
				MediaUrl = mediaURL
			}
		case message.MsgTypeShortVideo:
			material := c.wxContext.GetMaterial()
			if mediaURL, err := material.GetMediaURL(msg.MediaID); err != nil {
				log.Printf("material.MsgTypeShortVideo is err: %#v", err)
			} else {
				MediaUrl = mediaURL
			}
		}

		log.Printf("用户[%s]发来信息：[%s] %s \n", msg.FromUserName, msgType, msgText)

		roomCollection := model.Db.C("room")
		customerCollection := model.Db.C("customer")
		var room = model.Room{}
		roomCollection.Find(bson.M{"room_customer.customer_id": msg.FromUserName}).One(&room)

		// 尝试机器人回答
		if msgText != "" {
			aiDialogue = c.aiModule.Dialogue(msgText)
		}

		if room.RoomCustomer.CustomerId == "" {
			// 新接入
			userInfo, err := c.wxContext.GetUser().GetUserInfo(msg.FromUserName)
			if err != nil {
				log.Printf("WeiXinController.wxContext.GetUser().GetUserInfo() is err：%v", err.Error())
			}

			// 客户数据持久化
			customerCollection.Insert(&model.Customer{
				CustomerId:   msg.FromUserName,
				NickName:     userInfo.Nickname,
				CustomerType: common.NormalCustomer,
				Sex:          userInfo.Sex,
				HeadImgUrl:   userInfo.Headimgurl,
				Address:      fmt.Sprintf("%s_%s", userInfo.Province, userInfo.City),
				CreateTime:   time.Now(),
				UpdateTime:   time.Now(),
			})

			// 实时会话数据更新
			roomCollection.Insert(&model.Room{
				RoomCustomer: model.RoomCustomer{
					CustomerId:         msg.FromUserName,
					CustomerNickName:   userInfo.Nickname,
					CustomerHeadImgUrl: userInfo.Headimgurl,
				},
				RoomMessages: []model.RoomMessage{
					{
						Id:         common.GetNewUUID(),
						Type:       msgType,
						Msg:        msgText,
						AiMsg:      aiDialogue,
						MediaUrl:   MediaUrl,
						OperCode:   common.MessageFromCustomer,
						CreateTime: time.Now(),
					},
				},
				CreateTime: time.Now(),
			})
		} else {
			var (
				kefuColection = model.Db.C("kefu")
				kefuModel     = model.Kf{}
			)
			kefuColection.Find(bson.M{"id": room.RoomKf.KfId}).One(&kefuModel)
			if kefuModel.Id != "" && kefuModel.IsOnline == false {
				// 若接待的客服已经下线，则将用户重新放入待接入
				roomCollection.Update(
					bson.M{"room_customer.customer_id": msg.FromUserName},
					bson.M{"$set": bson.M{"room_kf": &model.RoomKf{}}})
			}

			// 实时会话数据更新
			query := bson.M{
				"room_customer.customer_id": msg.FromUserName,
			}
			changes := bson.M{
				"$push": bson.M{"room_messages": bson.M{"$each": []model.RoomMessage{
					{
						Id:         common.GetNewUUID(),
						Type:       msgType,
						Msg:        msgText,
						AiMsg:      aiDialogue,
						MediaUrl:   MediaUrl,
						OperCode:   common.MessageFromCustomer,
						CreateTime: time.Now(),
					},
				},
					"$slice": -100}},
			}
			if err := roomCollection.Update(query, changes); err != nil {
				log.Printf("实时房型数据更新异常：%s", err.Error())
			}
		}

		// 存储历史消息
		model.InsertMessage(model.Message{
			Id:         common.GetNewUUID(),
			Type:       msgType,
			CustomerId: msg.FromUserName,
			KfId:       room.RoomKf.KfId,
			Msg:        msgText,
			AiMsg:      aiDialogue,
			MediaUrl:   MediaUrl,
			OperCode:   common.MessageFromCustomer,
			CreateTime: time.Now(),
		})

		onlines, _ := model.Kf{}.QueryOnlines()
		if len(onlines) == 0 {
			return &message.Reply{MsgType: message.MsgTypeText, MsgData: message.NewText(common.KF_REPLY)}
		}

		return
	})

	//处理消息接收以及回复
	err := wcServer.Serve()
	if err != nil {
		fmt.Println(err)
		return
	}

	//发送接收成功
	wcServer.Send()
}
