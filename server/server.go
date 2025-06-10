package server

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net"
	"net/rpc"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/xtaci/kcp-go"
	"go.uber.org/zap"
)

type CustomServeCodec struct {
	conn         *kcp.UDPSession
	defaultCodec rpc.ServerCodec
	sugar 		*zap.SugaredLogger
}


// This is default rpc ServeConn codec ------------------------------------------
type gobServerCodec struct {
	rwc    io.ReadWriteCloser
	dec    *gob.Decoder
	enc    *gob.Encoder
	encBuf *bufio.Writer
	closed bool
}

func (c *gobServerCodec) ReadRequestHeader(r *rpc.Request) error {
	return c.dec.Decode(r)
}

func (c *gobServerCodec) ReadRequestBody(body any) error {
	return c.dec.Decode(body)
}

func (c *gobServerCodec) WriteResponse(r *rpc.Response, body any) (err error) {
	if err = c.enc.Encode(r); err != nil {
		if c.encBuf.Flush() == nil {
			// Gob couldn't encode the header. Should not happen, so if it does,
			// shut down the connection to signal that the connection is broken.
			log.Println("rpc: gob error encoding response:", err)
			c.Close()
		}
		return
	}
	if err = c.enc.Encode(body); err != nil {
		if c.encBuf.Flush() == nil {
			// Was a gob problem encoding the body but the header has been written.
			// Shut down the connection to signal that the connection is broken.
			log.Println("rpc: gob error encoding body:", err)
			c.Close()
		}
		return
	}
	return c.encBuf.Flush()
}

func (c *gobServerCodec) Close() error {
	if c.closed {
		// Only call c.rwc.Close once; otherwise the semantics are undefined.
		return nil
	}
	c.closed = true
	return c.rwc.Close()
}
// ------------------------------------------------------------------------------



func NewCustomServeCodec(conn *kcp.UDPSession, sugar *zap.SugaredLogger) *CustomServeCodec {
	newCustomCodec := &CustomServeCodec{
		conn: conn,
		sugar: sugar,
	}

	buf := bufio.NewWriter(conn)
	srv := &gobServerCodec{
		rwc:    conn,
		dec:    gob.NewDecoder(conn),
		enc:    gob.NewEncoder(buf),
		encBuf: buf,
	}

	newCustomCodec.defaultCodec = srv
	return newCustomCodec
}

func (c *CustomServeCodec) ReadRequestHeader(r *rpc.Request) error {
	c.sugar.Info("ReadRequestHeader")
	return c.defaultCodec.ReadRequestHeader(r)
}

func (c *CustomServeCodec) ReadRequestBody(x interface{}) error {
	c.sugar.Info("ReadRequestBody")
	return c.defaultCodec.ReadRequestBody(x)
}

func (c *CustomServeCodec) WriteResponse(r *rpc.Response, x interface{}) error {
	c.sugar.Info("WriteResponse")
	return c.defaultCodec.WriteResponse(r,x)
}

// Close can be called multiple times and must be idempotent.
func (c *CustomServeCodec) Close() error {
	c.sugar.Info("Closing RPC Connection for - ", c.conn.RemoteAddr().String())
	if err := c.defaultCodec.Close(); err != nil {
		return err
	}

	if err := c.conn.Close(); err != nil {
		c.sugar.Debug("error in closing kcp session, err - %v", err)
		return fmt.Errorf("error in closing kcp session, err - %v", err)
	}

	return nil

	// return c.defaultCodec.Close()
}

func InitServer(port string, sugar *zap.SugaredLogger) error {

	udpaddr, err := net.ResolveUDPAddr("udp", port)
	if err != nil {
		return errors.WithStack(err)
	}

	conn, err := net.ListenUDP("udp", udpaddr)
	if err != nil {
		return errors.WithStack(err)
	}

	handler := Handler{
		Conn:  conn,
		sugar: sugar,
	}

	err = rpc.Register(&handler)
	if err != nil {
		sugar.Fatal("Could Not Register RPC to Handler...\nError: ", err)
		return err
	}

	// ServeConn(block BlockCrypt, dataShards, parityShards int, conn net.PacketConn)
	// ListenWithOptions(laddr string, block BlockCrypt, dataShards, parityShards int)
	// ListenWithOptions(laddr, nil, 0, 0)
	lis, err := kcp.ServeConn(nil, 0, 0, conn)
	if err != nil {
		sugar.Fatal("Could Not Start the UDP(KCP) Server...\nError: ", err)
		return err
	}

	sentResponse := make(map[string]bool)
	var sentResponseMutex sync.Mutex
	sugar.Info("Started KCP Server...")
	for {
		session, err := lis.AcceptKCP()
		if err != nil {
			sugar.Error("Error accepting KCP connection: ", err)
			continue
		}
		remoteAddr := session.RemoteAddr().String()
		sugar.Infof("New incoming connection from %s", remoteAddr)
		sugar.Infof("Sent Response Map %v", sentResponse)
		
		
		// Ik this is shit, but what can i do ....
		sentResponseMutex.Lock()
		sugar.Infof("Sent Response Map for Incomming Connection %v", sentResponse[remoteAddr])
		if sentResponse[remoteAddr] {
			var buff []byte = make([]byte, 1024)
			size, err := session.Read(buff)
			sugar.Infof("The Incoming connection from %s - Buffer - %v - Size: %v ", remoteAddr, buff, size)	
			if err != nil {
				sugar.Infof("The Incoming connection from %s", remoteAddr, "- Reading Buffer Error")	
			}

			if size == 24 {
				sugar.Infof("The Incoming connection from %s", remoteAddr, " is treated as response ack")
				sugar.Infof("Ignoring %s", remoteAddr, " is treated as response ack")
				sentResponseMutex.Unlock()

				sentResponse[remoteAddr] = false
				continue
			} else {
				sugar.Info("Buffer Size - ", size)
				sugar.Info("Treating as Connection")
			}
		}
		sentResponseMutex.Unlock()

		go func(session *kcp.UDPSession) {
			session.SetWindowSize(2, 32)                               // Only 2 unacked packets maximum
			session.SetWriteDeadline(time.Now().Add(10 * time.Second)) // Limits total retry time
			// session.SetNoDelay(0, 15000, 0, 0)
			session.SetNoDelay(1, 10, 0, 0)
			session.SetDeadline(time.Now().Add(30 * time.Second)) // Overall timeout
			// session.SetACKNoDelay(false)                          // Batch ACKs to reduce traffic
			session.SetACKNoDelay(true)                          // Don't Batch ACKs to reduce traffic

			sugar.Info("Creating CustomServeCodec for Session - ", remoteAddr)
			customCodec := NewCustomServeCodec(session, sugar)
			sugar.Info("Serving Connection to Codec for Session - ", remoteAddr)
			err := rpc.ServeRequest(customCodec)
			if err != nil {
				sugar.Debugf("error in serving request to session: %s, error - %v", remoteAddr, err)
			}
			sugar.Info("Response Sent 👍")	
			sugar.Info("Serve Codec Done Calling Close for Session - ", remoteAddr)
			customCodec.Close() // Could you believe this - This guy send the Response ... How ??
			
			sentResponseMutex.Lock()
			sentResponse[remoteAddr] = true
			sugar.Info("Locking Sent Response Map Close for Session - ", remoteAddr, " Map - ", sentResponse)
			sentResponseMutex.Unlock()
		}(session)
		
	}
}
