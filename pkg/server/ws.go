package server

import (
	glog "log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/clarklee92/beehive/pkg/common/log"
	"github.com/clarklee92/viaduct/pkg/api"
	"github.com/clarklee92/viaduct/pkg/conn"
	"github.com/clarklee92/viaduct/pkg/lane"
	"github.com/clarklee92/viaduct/pkg/utils"
)

// websocket protocol server
type WSServer struct {
	options Options
	exOpts  api.WSServerOption
	server  *http.Server
}

// http server logger filter
type LoggerFilter struct{}

func (f *LoggerFilter) Write(p []byte) (n int, err error) {
	output := string(p)
	if strings.Contains(output, "http: TLS handshake error from") {
		return 0, nil
	}
	log.LOGGER.Error(output)
	return os.Stderr.Write(p)
}

func NewWSServer(opts Options, exOpts interface{}) *WSServer {
	extendOption, ok := exOpts.(api.WSServerOption)
	if !ok {
		panic("bad websocket option")
	}

	server := http.Server{
		Addr:      opts.Addr,
		TLSConfig: opts.TLS,
		ErrorLog:  glog.New(&LoggerFilter{}, "", glog.LstdFlags),
	}

	wsServer := &WSServer{
		options: opts,
		exOpts:  extendOption,
		server:  &server,
	}
	http.HandleFunc(extendOption.Path, wsServer.ServeHTTP)
	return wsServer
}

func (srv *WSServer) upgrade(w http.ResponseWriter, r *http.Request) *websocket.Conn {
	upgrader := websocket.Upgrader{
		HandshakeTimeout: srv.options.HandshakeTimeout,
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.LOGGER.Errorf("failed to upgrade to websocket")
		return nil
	}
	return conn
}

func (srv *WSServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if srv.exOpts.Filter != nil {
		if filtered := srv.exOpts.Filter(w, req); filtered {
			log.LOGGER.Warnf("failed to filter req")
			return
		}
	}

	wsConn := srv.upgrade(w, req)
	if wsConn == nil {
		return
	}

	conn := conn.NewConnection(&conn.ConnectionOptions{
		ConnType: api.ProtocolTypeWS,
		Base:     wsConn,
		ConnUse:  api.UseType(req.Header.Get("ConnectionUse")),
		Consumer: srv.options.Consumer,
		Handler:  srv.options.Handler,
		CtrlLane: lane.NewLane(api.ProtocolTypeWS, wsConn),
		State: &conn.ConnectionState{
			State:   api.StatConnected,
			Headers: utils.DeepCopyHeader(req.Header),
		},
		AutoRoute: srv.options.AutoRoute,
	})

	// connection callback
	if srv.options.ConnNotify != nil {
		srv.options.ConnNotify(conn)
	}

	// connection manager
	if srv.options.ConnMgr != nil {
		srv.options.ConnMgr.AddConnection(conn)
	}

	// serve connection
	go conn.ServeConn()
}

func (srv *WSServer) ListenAndServeTLS() error {
	return srv.server.ListenAndServeTLS("", "")
}

func (srv *WSServer) Close() error {
	if srv.server != nil {
		return srv.server.Close()
	}
	return nil
}
