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
	err = conn.SendData([]byte{00, 12, 48, 48, 48, 48, 48, 48, 48, 48, 78, 79, 48, 48})
	if err != nil {
		panic(err)
	}
	header, err := conn.ReadData(2,time.Second * 1) //read header length firs
	dLen := (uint(header[0] << 8) | uint(header[1]))
	data, err := conn.ReadData(dLen, time.Second * 1)
	fmt.Println(string(data))
	conn.Close()
	//fmt.Scanln()
}
