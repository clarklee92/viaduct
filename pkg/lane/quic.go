package lane

import (
	"time"

	"github.com/clarklee92/beehive/pkg/common/log"
	"github.com/clarklee92/beehive/pkg/core/model"
	"github.com/clarklee92/viaduct/pkg/packer"
	"github.com/clarklee92/viaduct/pkg/translator"
	"github.com/lucas-clemente/quic-go"
)

type QuicLane struct {
	writeDeadline time.Time
	readDeadline  time.Time
	stream        quic.Stream
}

func NewQuicLane(van interface{}) *QuicLane {
	if stream, ok := van.(quic.Stream); ok {
		return &QuicLane{stream: stream}
	}
	log.LOGGER.Errorf("oops! bad type of van")
	return nil
}

func (l *QuicLane) ReadMessage(msg *model.Message) error {
	rawData, err := packer.NewReader(l.stream).Read()
	if err != nil {
		return err
	}

	err = translator.NewTran().Decode(rawData, msg)
	if err != nil {
		log.LOGGER.Errorf("failed to decode message")
		return err
	}

	return nil
}

func (l *QuicLane) WriteMessage(msg *model.Message) error {
	rawData, err := translator.NewTran().Encode(msg)
	if err != nil {
		log.LOGGER.Errorf("failed to encode message")
		return err
	}

	_, err = packer.NewWriter(l.stream).Write(rawData)
	return err
}

func (l *QuicLane) Read(raw []byte) (int, error) {
	return l.stream.Read(raw)
}

func (l *QuicLane) Write(raw []byte) (int, error) {
	return l.stream.Write(raw)
}

func (l *QuicLane) SetReadDeadline(t time.Time) error {
	l.readDeadline = t
	return l.stream.SetReadDeadline(t)
}

func (l *QuicLane) SetWriteDeadline(t time.Time) error {
	l.writeDeadline = t
	return l.stream.SetWriteDeadline(t)
}
