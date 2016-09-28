package pool

import (
	"net"
	"bufio"
	"time"
	"errors"
	"io"
	"strings"
)

var (
	errConnectionUnusable = errors.New("connection is not usable")
	errTimeout = errors.New("i/o timeout")
)

//defines a connection to a remote peer
type Connection struct {
	Remote      string
	Uid         string
	conn        *net.TCPConn
	buffReader  *bufio.Reader
	readTimeout time.Duration
	Usable      bool
	connPool    *ConnectionPool
}


//Create a new client connection . This will initiate a TCP connection to a remote peer
func newClientConnction(pool *ConnectionPool) (*Connection) {

	c := &Connection{connPool:pool, Usable:false}
	return c
}

func (c *Connection) setReadTimeout(timeout time.Duration)  {
	c.readTimeout = timeout
}

func (c *Connection) Read(p []byte) (int, error) {
	if !c.Usable {
		return 0, errConnectionUnusable
	}

	if c.readTimeout != 0 {
		c.conn.SetReadDeadline(time.Now().Add(c.readTimeout))
	}

	n, err := c.conn.Read(p)
	if err != nil {
		if err == io.EOF {
			c.Usable = false;
		}
		if strings.Contains(err.Error(),"timeout"){
			err = errors.New("i/o timeout")
		}
	}

	return n, err
}

//Connect to the remote entity
func (c *Connection) connect(endpoint string, connectionTimeout time.Duration, keepAlive bool) error  {

	conn, err :=   net.DialTimeout("tcp", endpoint, connectionTimeout)
	if err != nil {
		return err
	}

	c.conn = conn.(*net.TCPConn)

	c.Uid = c.conn.LocalAddr().String()
	c.Remote = endpoint
	c.conn.SetKeepAlive(keepAlive)
	c.buffReader = bufio.NewReader(c)
	c.Usable = true;

	return nil
}

//Send a []byte over the network
func (c *Connection) SendData(data []byte) error {

	if !c.Usable {
		return errConnectionUnusable
	}

	count := 0
	size := len(data)
	for count < size {
		n, err := c.conn.Write(data[count:])
		if err != nil {
			if err == io.EOF {
				c.Usable = false;
			}
			return err
		}

		count += n
	}

	return nil
}

//Read size byte of data and return is to the caller
func (c *Connection) ReadData(size uint, timeout time.Duration) ([]byte, error) {

	ret := make([]byte, size)
	var err error

	_, err = io.ReadFull(c.buffReader, ret)
	return ret, err

}

func (c *Connection) Close()  {
	c.connPool.put(c)
}

func (c *Connection) Shutdown() {
	c.conn.Close()
}

