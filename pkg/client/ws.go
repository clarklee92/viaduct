package client

import (
	"fmt"
	"io/ioutil"

	"github.com/gorilla/websocket"
	"github.com/clarklee92/beehive/pkg/common/log"
	"github.com/clarklee92/viaduct/pkg/api"
	"github.com/clarklee92/viaduct/pkg/conn"
	"github.com/clarklee92/viaduct/pkg/lane"
	"github.com/clarklee92/viaduct/pkg/utils"
)

// the client based on websocket
type WSClient struct {
	options Options
	exOpts  api.WSClientOption
	dialer  *websocket.Dialer
}

// new websocket client instance
func NewWSClient(options Options, exOpts interface{}) *WSClient {
	extendOption, ok := exOpts.(api.WSClientOption)
	if !ok {
		panic("bad websocket extend option")
	}

	return &WSClient{
		options: options,
		exOpts:  extendOption,
		dialer: &websocket.Dialer{
			TLSClientConfig:  options.TLSConfig,
			HandshakeTimeout: options.HandshakeTimeout,
		},
	}
}

// Connect try to connect remote server
func (c *WSClient) Connect() (conn.Connection, error) {
	header := c.exOpts.Header
	header.Add("ConnectionUse", string(c.options.ConnUse))
	wsConn, resp, err := c.dialer.Dial(c.options.Addr, header)
	if err == nil {
		log.LOGGER.Infof("dial %s successfully", c.options.Addr)

		// do user's processing on connection or response
		if c.exOpts.Callback != nil {
			c.exOpts.Callback(wsConn, resp)
		}
		return conn.NewConnection(&conn.ConnectionOptions{
			ConnType: api.ProtocolTypeWS,
			ConnUse:  c.options.ConnUse,
			Base:     wsConn,
			Consumer: c.options.Consumer,
			Handler:  c.options.Handler,
			CtrlLane: lane.NewLane(api.ProtocolTypeWS, wsConn),
			State: &conn.ConnectionState{
				State:   api.StatConnected,
				Headers: utils.DeepCopyHeader(c.exOpts.Header),
			},
			AutoRoute: c.options.AutoRoute,
		}), nil
	}

	// something wrong!!
	var respMsg string
	if resp != nil {
		body, errRead := ioutil.ReadAll(resp.Body)
		if errRead != nil {
			respMsg = fmt.Sprintf("response code: %d, response body: %s", resp.StatusCode, string(body))
		} else {
			respMsg = fmt.Sprintf("response code: %d", resp.StatusCode)
		}
		resp.Body.Close()
		return nil, err
	}
	log.LOGGER.Errorf("dial websocket error(%+v), response message: %s", err, respMsg)

	return nil, err
}
