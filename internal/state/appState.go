package state

import (
	"github.com/ktt-ol/spaceDevices/pkg/structs"
	"github.com/ktt-ol/status2/internal/events"
	"errors"
)

type MqttState struct {
	Connected         bool `json:"connected"`
	SpaceBrokerOnline bool `json:"spaceBrokerOnline"`
}

type OpenValueTs struct {
	Value     OpenValue `json:"state"`
	Timestamp int64 `json:"timestamp"`
}

type OpenState struct {
	Space     *OpenValueTs
	Radstelle *OpenValueTs
	Lab3d     *OpenValueTs
	Machining *OpenValueTs
}

func (os *OpenState) OpenStateForEvent(event events.EventName) (*OpenValueTs, error) {
	switch event {
	case events.TOPIC_SPACE_OPEN_STATE:
		return os.Space, nil
	case events.TOPIC_RADSTELLE_OPEN_STATE:
		return os.Radstelle, nil
	case events.TOPIC_LAB_3D_OPEN_STATE:
		return os.Lab3d, nil
	case events.TOPIC_MACHINING_OPEN_STATE:
		return os.Machining, nil
	default:
		return nil, errors.New("Not an open state event: " + string(event))
	}
}

type SpaceDevicesState struct {
	structs.PeopleAndDevices
	Timestamp int64 `json:"timestamp"`
}

type PowerValueTs struct {
	Value     float64 `json:"value"`
	Timestamp int64   `json:"timestamp"`
}

type PowerUsageState struct {
	Front *PowerValueTs `json:"front"`
	Back  *PowerValueTs `json:"back"`
	Machining  *PowerValueTs `json:"machining"`
}

type FreifunkState struct {
	ClientCount uint
	Timestamp   int64
}

type State struct {
	Mqtt         *MqttState
	Open         *OpenState
	SpaceDevices *SpaceDevicesState
	PowerUsage   *PowerUsageState
	Freifunk     *FreifunkState
}

func NewDefaultState() *State {
	return &State{
		Mqtt: &MqttState{
			Connected:         false,
			SpaceBrokerOnline: false,
		},
		Open: &OpenState{
			Space:     &OpenValueTs{Value: NONE, Timestamp: 0},
			Radstelle: &OpenValueTs{Value: NONE, Timestamp: 0},
			Lab3d:     &OpenValueTs{Value: NONE, Timestamp: 0},
			Machining: &OpenValueTs{Value: NONE, Timestamp: 0},
		},
		SpaceDevices: &SpaceDevicesState{
			PeopleAndDevices: structs.PeopleAndDevices{
				DeviceCount:         0,
				PeopleCount:         0,
				UnknownDevicesCount: 0,
				People:              []structs.Person{},
			},
			Timestamp: 0,
		},
		PowerUsage: &PowerUsageState{
			Front: &PowerValueTs{Value: 0.0, Timestamp: 0},
			Back:  &PowerValueTs{Value: 0.0, Timestamp: 0},
			Machining:  &PowerValueTs{Value: 0.0, Timestamp: 0},
		},
		Freifunk: &FreifunkState{
			ClientCount: 0,
			Timestamp:   0,
		},
	}
}
