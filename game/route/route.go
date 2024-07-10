package route

import (
	"core/repo"
	"framework/node"
	"game/handler"
	"game/logic"
)

func RegisterHandler(r *repo.Manager) node.LogicHandler {

	handles := make(node.LogicHandler)
	um := logic.NewUnionManager()
	unionHandler := handler.NewUnionHandler(r, um)
	handles["unionHandler.createRoom"] = unionHandler.CreateRoom
	handles["unionHandler.joinRoom"] = unionHandler.JoinRoom
	gameHandler := handler.NewGameHandler(r, um)
	handles["gameHandler.roomMessageNotify"] = gameHandler.RoomMessageNotify
	handles["gameHandler.gameMessageNotify"] = gameHandler.GameMessageNotify
	return handles

}
