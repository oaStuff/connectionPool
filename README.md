This is an implementation of a connection pool manager in GO.

#### Getting the library
```
    go get github.com/oaStuff/connectionPool
```

#### Using the connection pool

```go
package main

import (
	"fmt"
	"github.com/oaStuff/connectionPool"
	"time"
)

func main() {
	fmt.Println("Using connection Pool.")
	//create a pool with 2 connections to the remote entity
	cpool, _ := pool.NewConnectionPool(2,"192.168.56.50:9998",time.Second * 1,time.Second * 1,nil)
	conn, err := cpool.Get(time.Second * 2)
	if err != nil {
		fmt.Printf("Error is %v\n",err)
		return
	}
	
	//use the connection
	
	conn.SendData([]byte("sample data been sent"))
	response, _ := conn.ReadData(12, time.Second * 5) //read 12 bytes from network
	fmt.Println(response)
	
	//more activity
	conn.SendData([]byte("more data"))
	ret := make([]byte,512)
	conn.Read(ret)
	fmt.Println(ret)
	conn.Close()
}
```
