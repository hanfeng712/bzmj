package robot

import (
	"fmt"
	"rpc"
)

func (self *SRobot) HandleLoginRsp(conn rpc.RpcConn, login rpc.Login) error {
	fmt.Print("recv login")
	return nil
}
