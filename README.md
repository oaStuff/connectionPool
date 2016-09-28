A golang implementation of a connection pool manager

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
#### Explanation:
```go
pool.NewConnectionPool(2,"192.168.56.50:9998",time.Second * 1,time.Second * 1,nil)
```
The above will create a connection pool with **2** connections to the remote system
at IP address 192.168.56.50 on port 9998. The pool will try to always keep the 
2 connections active. There is a background gorountine that ensures that the
connections are active. The first time.Duration value specifies the read timeout
while the second specified the connection timeout. The last argument expected is
an interface of type **Notifier**. (Notifier notifies the "listening" object about
conditions of the pool.)

```go
conn, err := cpool.Get(time.Second * 2)
````
You get a connection from the pool using the **Get(timeout time.Duration)** method.
This will either return the connection object or will timeout after waiting for
the value specified.

```go
conn.Close()
````
The **close** method should alway be called after using the connection.
This ensures that the connection is returned back to the pool.
you can always do the following to ensure that

```go
conn, err := cpool.Get(time.Second * 2)
defer conn.Close()
````
This uses the **defer** construct to achieve returning the connection to the pool.
```go
conn.SendData([]byte("sample data been sent"))
```
Sending data is done using the **SendData([]byte)** method.

```go
response, _ := conn.ReadData(12, time.Second * 5)
````
Reading of data can be done using 2 different methods. The above will read 
exactly 12 bytes of data or timeout after waiting for the specified time duration.

```go
ret := make([]byte,512)
conn.Read(ret)
````
Reading of data can also be done using the above method. This will block until
the size of the slice is read or the connection is cloased or an error occurs.