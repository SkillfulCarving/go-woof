package public

import (
	"context"
	"io"
	"net"
)

func DataExchange(conn1, conn2 net.Conn) {
	defer func() {
		_ = conn1.Close()
		_ = conn2.Close()
	}()
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer cancel()
		_, _ = io.Copy(conn1, conn2)
	}()
	go func() {
		defer cancel()
		_, _ = io.Copy(conn2, conn1)
	}()
	<-ctx.Done()
}
