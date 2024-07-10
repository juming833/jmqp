package route

import (
	"core/repo"
	"framework/node"
	"hall/handler"
)

func RegisterHandler(r *repo.Manager) node.LogicHandler {

	handles := make(node.LogicHandler)
	userHandler := handler.NewUserHandler(r)
	handles["userHandler.updateUserAddress"] = userHandler.UpdateUserAddress

	return handles

}
