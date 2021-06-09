package action

import (
	"encoding/json"
	"errors"
	"github.com/itxaka/luet-mtree/pkg/helpers"
	"github.com/itxaka/luet-mtree/pkg/log"
	"github.com/mudler/go-pluggable"
	"github.com/mudler/luet/pkg/bus"
	"strings"
)

type UnpackEvent struct {
	Name string
	Data EventData
	File string
}

type EventData struct {
	Image string
	Dest  string
}

type LuetEvent struct {
	event   pluggable.EventType
	payload string
}

// ImageBlacklist has a list of images that we need to skip as they dont contain mtree checks
var ImageBlacklist = []string{"repository.yaml", "tree.tar", "repository.meta.yaml.tar", "compilertree.tar"}

func NewEventDispatcherAction(event string, payload string) *LuetEvent {
	return &LuetEvent{event: pluggable.EventType(event), payload: payload}
}

func (event LuetEvent) Run() (map[string]string, error) {

	log.Log("Got event: %s\n", event.event)

	switch event.event {
	case bus.EventImagePostUnPack:
		// Unpack payload
		dataTmp := UnpackEvent{}
		err := json.Unmarshal([]byte(event.payload), &dataTmp)
		if err != nil {
			log.Log("Error while unmarshalling payload")
			log.Log("Payload: %s", event.payload)
			return helpers.WrapErrorMap(err)
		}

		// Check correct payload data
		if dataTmp.Data.Image == "" || dataTmp.Data.Dest == "" {
			log.Log("Some fields are missing from the event, cannot continue")
			return helpers.WrapErrorMap(errors.New("fields missing from payload"))
		}

		// Check blacklist to skip images
		for _, s := range ImageBlacklist {
			if strings.Contains(dataTmp.Data.Image, s) {
				log.Log("Image type found in blacklist, skipping")
				return helpers.WrapErrorMap(nil)
			}
		}
		return UnpackAndMtree(dataTmp.Data.Image, dataTmp.Data.Dest)
	default:
		log.Log("No event that I can recognize")
		return helpers.WrapErrorMap(nil)
	}
}
