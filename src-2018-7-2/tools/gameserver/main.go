// bzmj project main.go
package main

//protoc --go_out=. *.proto
import (
	"connector"
	"fmt"
	"sync"
)

func main() {
	fmt.Println("Hello World!")

	wg := &sync.WaitGroup{}
	cnServer := connector.NewCNServer()
	cnServer.StartClientService(1, wg)
	for {
	}
}
