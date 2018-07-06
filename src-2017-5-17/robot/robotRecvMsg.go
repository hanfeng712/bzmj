package robot

import (
	"bzmj/rpc"
	"fmt"
)

func (self *SRobot) HandleLoginRsp(conn rpc.RpcConn, login rpc.Login) error {
	fmt.Print("recv login")
	return nil
}
