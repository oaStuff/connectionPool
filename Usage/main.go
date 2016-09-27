package main

import (
	"fmt"
	"github.com/oaStuff/connectionPool"
	"time"
)

func main() {
	fmt.Println("Using connection Pool.")
	cpool, _ := pool.NewConnectionPool(1,"192.168.56.50:9998",time.Second * 1,time.Second * 1,nil)
	conn, err := cpool.Get(time.Second * 2)
	if err != nil {
		fmt.Printf("Error is %v\n",err)
		return
	}
	conn.Close()
	//fmt.Scanln()
}
