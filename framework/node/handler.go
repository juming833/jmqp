package node

import "framework/remote"

type LogicHandler map[string]HandlerFunc
type HandlerFunc func(session *remote.Session, msg []byte) any
