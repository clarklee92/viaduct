package main

import (
	"strings"

	"github.com/clarklee92/beehive/pkg/common/log"
	"github.com/clarklee92/viaduct/examples/chat/config"
)

func main() {
	cfg := config.InitConfig()

	var err error
	if strings.Compare(cfg.CmdType, "server") == 0 {
		err = StartServer(cfg)
	} else {
		err = StartClient(cfg)
	}
	if err != nil {
		log.LOGGER.Errorf("start %s failed, error: %+v", cfg.CmdType, err)
	}
}
