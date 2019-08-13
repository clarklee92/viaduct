package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	"github.com/clarklee92/beehive/pkg/common/log"
	"github.com/clarklee92/beehive/pkg/core/model"
	"github.com/clarklee92/viaduct/examples/chat/config"
	"github.com/clarklee92/viaduct/pkg/api"
	"github.com/clarklee92/viaduct/pkg/client"
	"github.com/clarklee92/viaduct/pkg/conn"
	"github.com/clarklee92/viaduct/pkg/mux"
)

var clientStdWriter = bufio.NewWriter(os.Stdout)

func handleClient(container *mux.MessageContainer, writer mux.ResponseWriter) {
	fmt.Printf("receive message: %s", container.Message.GetContent())
	if container.Message.IsSync() {
		writer.WriteResponse(container.Message, "ack")
	}
}

func initClientEntries() {
	mux.Entry(mux.NewPattern("*").Op("*"), handleClient)
}

func StartClient(cfg *config.Config) error {
	//tls, err := GetTlsConfig(cfg)
	//if err != nil {
	//	return err
	//}

	initClientEntries()

	// just for testing
	tls := &tls.Config{InsecureSkipVerify: true}

	var exOpts interface{}

	header := make(http.Header)
	header.Add("client_id", "client1")
	switch cfg.Type {
	case api.ProtocolTypeQuic:
		exOpts = api.QuicClientOption{
			Header: header,
		}
	case api.ProtocolTypeWS:
		exOpts = api.WSClientOption{
			Header: header,
		}
	}

	client := client.Client{
		Options: client.Options{
			Type:      cfg.Type,
			Addr:      cfg.Addr,
			TLSConfig: tls,
			AutoRoute: true,
			ConnUse:   api.UseTypeMessage,
		},
		ExOpts: exOpts,
	}

	connClient, err := client.Connect()
	if err != nil {
		return err
	}
	stat := connClient.ConnectionState()
	log.LOGGER.Infof("connect stat:%+v", stat)

	return SendStdin([]conn.Connection{connClient}, "client")
}

func SendStdin(conns []conn.Connection, source string) error {
	input := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("send message: ")
		inputData, err := input.ReadString('\n')
		if err != nil {
			log.LOGGER.Errorf("failed to read input, error: %+v", err)
			return err
		}
		message := model.NewMessage("").
			BuildRouter(source, "", "viaduct_message", "update").
			FillBody(inputData)

		for _, conn := range conns {
			err = conn.WriteMessageAsync(message)
			if err != nil {
				log.LOGGER.Errorf("failed to write message async, error:%+v", err)
			}
		}
	}
	return nil
}
