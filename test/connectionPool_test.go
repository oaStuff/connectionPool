package test

import (
	"testing"
	"github.com/oaStuff/connectionPool"
	"time"
)

func TestConnectionPool(t *testing.T)  {
	_, err := pool.NewConnectionPool(0,"localhost:9999",time.Second * 2, time.Second * 2,nil)
	if err == nil {
		t.Fatal("Connection should return non nill")
	}

	t.Log(err)
}

func TestGetTimeout(t *testing.T)  {
	cPool, err := pool.NewConnectionPool(1,"localhost:9999",time.Second * 2, time.Second * 2,nil)
	_, err = cPool.Get(time.Second * 2)
	if err == nil {
		t.Fatal("Get should have timed out")
	}

	t.Log(err)
}