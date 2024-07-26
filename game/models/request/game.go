package request

import "game/component/proto"

type RoomMessageReq struct {
	Type proto.RoomMessageType `json:"type"`
	Data RoomMessageData       `json:"data"`
}

type RoomMessageData struct {
	IsExit  bool `json:"isExit"`
	IsReady bool `json:"isReady"`
}
