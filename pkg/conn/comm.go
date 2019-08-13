package conn

import (
	"github.com/clarklee92/beehive/pkg/common/log"
	"github.com/clarklee92/beehive/pkg/core/model"
	"github.com/clarklee92/viaduct/pkg/lane"
)

type responseWriter struct {
	Type string
	Van  interface{}
}

// write response
func (r *responseWriter) WriteResponse(msg *model.Message, content interface{}) {
	response := msg.NewRespByMessage(msg, content)
	err := lane.NewLane(r.Type, r.Van).WriteMessage(response)
	if err != nil {
		log.LOGGER.Errorf("failed to write response, error: %+v", err)
	}
}

// write error
func (r *responseWriter) WriteError(msg *model.Message, errMsg string) {
	response := model.NewErrorMessage(msg, errMsg)
	err := lane.NewLane(r.Type, r.Van).WriteMessage(response)
	if err != nil {
		log.LOGGER.Errorf("failed to write error, error: %+v", err)
	}
}
