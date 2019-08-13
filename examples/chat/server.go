package main

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"

	"github.com/clarklee92/beehive/pkg/common/log"
	"github.com/clarklee92/beehive/pkg/core/model"
	"github.com/clarklee92/viaduct/examples/chat/config"
	"github.com/clarklee92/viaduct/pkg/api"
	"github.com/clarklee92/viaduct/pkg/cmgr"
	"github.com/clarklee92/viaduct/pkg/conn"
	"github.com/clarklee92/viaduct/pkg/mux"
	"github.com/clarklee92/viaduct/pkg/server"
)

var serverStdWriter = bufio.NewWriter(os.Stdout)

func handleServer(container *mux.MessageContainer, writer mux.ResponseWriter) {
	fmt.Printf("receive message: %s", container.Message.GetContent())
	if container.Message.IsSync() {
		writer.WriteResponse(container.Message, "success")
	}
}

func initServerEntries() {
	mux.Entry(mux.NewPattern("*").Op("*"), handleServer)
}

func ConnNotify(conn conn.Connection) {
	log.LOGGER.Info("receive a connection")
}

// just for testing
func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{Certificates: []tls.Certificate{tlsCert}}
}

func StartServer(cfg *config.Config) error {
	//tls, err := GetTlsConfig(cfg)
	//if err != nil {
	//	return err
	//}

	// just for testing
	tls := generateTLSConfig()

	connMgr := cmgr.NewManager(nil)

	var exOpts interface{}
	switch cfg.Type {
	case api.ProtocolTypeQuic:
		exOpts = api.QuicServerOption{}
	case api.ProtocolTypeWS:
		exOpts = api.WSServerOption{
			Path: "/test",
		}
	}

	server := server.Server{
		Type:       cfg.Type,
		Addr:       cfg.Addr,
		TLSConfig:  tls,
		AutoRoute:  true,
		ConnMgr:    connMgr,
		ConnNotify: ConnNotify,
		ExOpts:     exOpts,
	}

	initServerEntries()

	go func() {
		err := server.ListenAndServeTLS("", "")
		if err != nil {
			log.LOGGER.Errorf("listen and serve tls failed, error: %+v", err)
		}
	}()

	input := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("send message: ")
		inputData, err := input.ReadString('\n')
		if err != nil {
			log.LOGGER.Errorf("failed to read input, error: %+v", err)
			return err
		}

		var conns []conn.Connection
		connMgr.Range(func(key, value interface{}) bool {
			conns = append(conns, value.(conn.Connection))
			return true
		})

		message := model.NewMessage("").
			BuildRouter("server", "", "viaduct_message", "update").
			FillBody([]byte(inputData))

		for _, conn := range conns {
			err = conn.WriteMessageAsync(message)
			if err != nil {
				log.LOGGER.Errorf("failed to write message async, error:%+v", err)
			}
		}
	}

	return nil
}
