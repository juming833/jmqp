package route

import (
	"connector/handler"
	"core/repo"
	"framework/nets"
)

type Route struct {
}

func RegisterHandler(r *repo.Manager) nets.LogicHandler {
	handles := make(nets.LogicHandler)
	entryHandle := handler.NewEntryHandler(r)
	handles["entryHandler.entry"] = entryHandle.Entry

	return handles

}
