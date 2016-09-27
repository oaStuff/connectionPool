package pool

import (
	"time"
	"github.com/op/go-logging"
	"os"
	"fmt"
	"sync/atomic"
	"io"
	"errors"
	"math"
)

const (
	INFINITE = math.MaxInt64
)

//Nnotification status that will be sent in the notifier interface method 'Notify()'
const (
	DISCONNECT = iota
	CONNECTED
	POOL_EMPTY
)

var log *logging.Logger

//Notification callback definition
type Notifier interface {
	Notify(condition uint)
}


//Connnection pool definition
type ConnectionPool struct {
	activeConnections chan *Connection
	deadConnections chan *Connection
	totalConnections uint32
	remoteAddr string
	readTimeout time.Duration
	closed bool
	connectedCount int32
	notifier Notifier
}


func init() {

	log = logging.MustGetLogger("connectionPool")
	format := logging.MustStringFormatter(`%{color}%{time:15:04:05.000} [%{level:.4s}] %{message} %{color:reset}`)
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)
	logging.SetBackend(backendFormatter)
}

//Create a connection pool of size numConnections to remote address addr
func NewConnectionPool (numConnections uint32, addr string, conReadTimeout time.Duration, connectionTimeout time.Duration, notify Notifier) (*ConnectionPool, error) {

	if numConnections < 1{
		return nil, errors.New("Connection pool size must be greater than zero.")
	}

	cp := &ConnectionPool{remoteAddr: addr,
		activeConnections:make(chan *Connection,numConnections + 2),
		deadConnections:make(chan *Connection, numConnections + 2),
		readTimeout:conReadTimeout,
		totalConnections:numConnections,
		closed:false,
		connectedCount:0,
		notifier:notify}

	var x uint32
	for x = 0; x < numConnections; x++ {
		con := newClientConnction(cp)
		con.Uid = string(x)
		cp.deadConnections <- con
	}

	//spin off the goroutine to handle connection to remote
	go cp.handleConnectToRemote(connectionTimeout)

	return cp, nil
}


//This gorouting make the connection to remote system
//it gets the dead connections and connects to the remote side
//if the connection attempt fails, it pulses for 5secs
func (cp *ConnectionPool) handleConnectToRemote(connectionTimeout time.Duration)  {

	for conn := range cp.deadConnections {

		//if closed just exit the loop
		if cp.closed{
			break
		}

		go func(c *Connection) {
			err := c.connect(cp.remoteAddr, connectionTimeout, true)
			if err != nil {
				log.Error(fmt.Sprintf("Error connecting : %v",err))
				time.Sleep(time.Second * 2) //wait for a while b4 retrying
				cp.deadConnections <- c
			} else {
				//check if we have been closed, so just shut it down
				if cp.closed{
					c.Shutdown()
					return
				}

				log.Info(fmt.Sprintf("Successfully connected [%s] to [%s] ",c.Uid, cp.remoteAddr))
				c.setReadTimeout(cp.readTimeout)
				cp.activeConnections <- c
				atomic.AddInt32(&cp.connectedCount, 1)
			}
		}(conn)

	}

	log.Info(fmt.Sprintf("Connection pool: connection gorouting exiting for [%s]",cp.remoteAddr))
}


//Places a connection in the dead channel in order for automatic connection to remote
func (cp *ConnectionPool) queueForReConnection(con *Connection)  {
	atomic.AddInt32(&cp.connectedCount, -1)
	log.Info(fmt.Sprintf("Requeuing [%s/%s] for reconnection because it disconnected",con.Remote, con.Uid))
	con.Shutdown()
	cp.deadConnections <- con

	if cp.connectedCount < 1 {
		if cp.notifier != nil {
			go cp.notifier.Notify(POOL_EMPTY)
		}
	}
}


//Check if the connection is closed
func (cp *ConnectionPool) checkError(conn *Connection, err error)  {
	if err == io.EOF{
		cp.queueForReConnection(conn)
	}
}


//Fetch a connection from the pool or nil if timeout elapses
func (cp *ConnectionPool) Get(timeout time.Duration) (*Connection, error) {
	var conn *Connection
	if timeout == INFINITE {
		conn = <-cp.activeConnections
	} else {
		select {
		case conn = <-cp.activeConnections:
		case <-time.After(timeout):
			log.Warning("Timed out trying to get connection from the pool")
			return nil, errors.New(fmt.Sprintf("Timed out trying to get a connection after wating for %v",timeout))
		}
	}

	return conn, nil
}


//Return back the connection to the pool. If it is unusable, place it on the dead queue
func (cp *ConnectionPool) put(conn *Connection)  {

	if conn.Usable {
		cp.activeConnections <- conn
	}else {
		cp.queueForReConnection(conn)
	}
}

func (cp *ConnectionPool) IsConnected() bool {
	return atomic.LoadInt32(&cp.connectedCount) > 0
}

//Shut down the connection pool.
//Just iterate over the connections and call their respective Shutdown()
func (cp *ConnectionPool) Shutdown()  {
	//if we are closed just return
	if cp.closed {
		return
	}

	cp.closed = true;

	for{
		select {
		case conn := <-cp.activeConnections:
			conn.Shutdown()
			cp.deadConnections <- conn //to force the connection goroutine to exit
		default:
			return
		}
	}

}
