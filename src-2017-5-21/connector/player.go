package connector

import (
	"fmt"
	"logger"
	"rpc"
)

type player struct {
	uid  uint64
	conn rpc.RpcConn
}

func LoadPlayer(uid uint64) *player {
	return NewPlayer(uid)
}

func NewPlayer(uid uint64) *player {
	lRet := &player{uid: uid}
	return lRet
}

func (p *player) GetUid() uint64 {
	return p.uid
}

func (p *player) OnQuit() {
	ts("player:OnQuit", p.GetUid())
	defer te("player:OnQuit", p.GetUid())
	fmt.Println("退出 ")

	if p.conn != nil {
		p.conn.Lock()
		defer p.conn.Unlock()
	}

	logger.Info("OnQuit p.conn.Lock() end")
}
