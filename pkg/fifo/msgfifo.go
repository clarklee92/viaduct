package fifo

import (
	"fmt"

	"github.com/clarklee92/beehive/pkg/common/log"
	"github.com/clarklee92/beehive/pkg/core/model"
	"github.com/clarklee92/viaduct/pkg/comm"
)

type MessageFifo struct {
	fifo chan model.Message
}

// set the fifo capacity to MessageFiFoSizeMax
func NewMessageFifo() *MessageFifo {
	return &MessageFifo{
		fifo: make(chan model.Message, comm.MessageFiFoSizeMax),
	}
}

// Put put the message into fifo
func (f *MessageFifo) Put(msg *model.Message) {
	select {
	case f.fifo <- *msg:
	default:
		// discard the old message
		<-f.fifo
		// push into fifo
		f.fifo <- *msg
		log.LOGGER.Warnf("too many message received, fifo overflow")
	}
}

// Get get message from fifo
// this api is blocked when the fifo is empty
func (f *MessageFifo) Get(msg *model.Message) error {
	var ok bool
	*msg, ok = <-f.fifo
	if !ok {
		return fmt.Errorf("the fifo is broken")
	}
	return nil
}
