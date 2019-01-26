package mqtt

import (
	"testing"
	"github.com/stretchr/testify/require"
	"github.com/ktt-ol/status2/internal/events"
	"github.com/ktt-ol/status2/internal/state"
	"github.com/ktt-ol/status2/internal/test"
)

func Test_newSpaceState(t *testing.T) {
	eventsMock := new(test.EventManagerMock)
	appState := state.NewDefaultState()
	manager := MqttManager{state: appState, events: eventsMock}

	manager.newSpaceState()
	require.Equal(t, appState.Open.Space.Value, state.NONE)
	require.Equal(t, eventsMock.EmitCount, 0)
	require.Equal(t, events.EventName(""), eventsMock.LastEvent)

	manager.lastOpenState = &state.OpenValueTs{Value: state.OPEN_PLUS}
	manager.newSpaceState()
	require.Equal(t, appState.Open.Space.Value, state.OPEN_PLUS)
	require.Equal(t, eventsMock.EmitCount, 1)
	require.Equal(t, events.TOPIC_SPACE_OPEN_STATE, eventsMock.LastEvent)

	manager.lastOpenStateNext = &state.OpenValueTs{Value: state.NONE}
	manager.newSpaceState()
	require.Equal(t, state.CLOSING, appState.Open.Space.Value)
	require.Equal(t, eventsMock.EmitCount, 2)

	manager.lastOpenState = &state.OpenValueTs{Value: state.MEMBER}
	manager.lastOpenStateNext = nil
	manager.newSpaceState()
	require.Equal(t, state.MEMBER, appState.Open.Space.Value)
	require.Equal(t, eventsMock.EmitCount, 3)

	manager.lastOpenState = &state.OpenValueTs{Value: state.NONE}
	manager.lastOpenStateNext = &state.OpenValueTs{Value: state.OPEN}
	manager.newSpaceState()
	require.Equal(t, state.NONE, appState.Open.Space.Value)
	require.Equal(t, eventsMock.EmitCount, 4)
}



func Test_onDevicesChange(t *testing.T) {
	eventsMock := new(test.EventManagerMock)
	appState := state.NewDefaultState()
	manager := MqttManager{state: appState, events: eventsMock}

	mMock := new(test.MessageMock)

	// some real live data
	mMock.PayloadData = []byte(`{"people":[{"name":"Holger","devices":[{"name":"Handy","location":"Space"},{"name":"Mac","location":"Space"}]},{"name":"MarvinGS","devices":[{"name":"Handy","location":"Space"},{"name":"Notebook","location":"Space"}]},{"name":"larsh404","devices":[{"name":"Aquaris X","location":"Space"},{"name":"SE","location":"Space"}]},{"name":"larsho","devices":null},{"name":"mbl","devices":[{"name":"ring²","location":"Space"}]},{"name":"sre","devices":[{"name":"Droid4","location":"Space"},{"name":"X250","location":"Space"}]}],"peopleCount":7,"deviceCount":28,"unknownDevicesCount":8}`)
	manager.onDevicesChange(nil, mMock)
	require.Equal(t, uint16(28), appState.SpaceDevices.DeviceCount)
	require.Equal(t, uint16(8), appState.SpaceDevices.UnknownDevicesCount)
	require.Equal(t, uint16(7), appState.SpaceDevices.PeopleCount)
	require.Equal(t, 6, len(appState.SpaceDevices.People))
	person := appState.SpaceDevices.People[0]
	require.Equal(t, "Holger", person.Name)
	require.Equal(t, 2, len(person.Devices))
	require.Equal(t, "Handy", person.Devices[0].Name)
	require.Equal(t, "Space", person.Devices[0].Location)
	require.Equal(t, 1, eventsMock.EmitCount)
	require.Equal(t, events.TOPIC_SPACE_DEVICES, eventsMock.LastEvent)

	mMock.PayloadData = []byte(`{"people":[{"name":"Holger","devices":[{"name":"Handy","location":"Space"}]}],"peopleCount":1,"deviceCount":4,"unknownDevicesCount":4}`)
	manager.onDevicesChange(nil, mMock)
	require.Equal(t, uint16(1), appState.SpaceDevices.PeopleCount)
	require.Equal(t, 1, len(appState.SpaceDevices.People))
	require.Equal(t, 2, eventsMock.EmitCount)

	// missing and invalid attributes are still ok for the parser
	mMock.PayloadData = []byte(`{"people":[],"apeopleCount":1,"bdeviceCount":4,"cunknownDevicesCount":4}`)
	manager.onDevicesChange(nil, mMock)
	require.Equal(t, uint16(0), appState.SpaceDevices.DeviceCount)
	require.Equal(t, uint16(0), appState.SpaceDevices.UnknownDevicesCount)
	require.Equal(t, uint16(0), appState.SpaceDevices.PeopleCount)
	require.Equal(t, 3, eventsMock.EmitCount)

	// what about parsing errors?
	mMock.PayloadData = []byte("{invalid_json")
	manager.onDevicesChange(nil, mMock)
	require.Equal(t, 3, eventsMock.EmitCount)
	mMock.PayloadData = []byte(`{"peopleCount":"moin"}`)
	manager.onDevicesChange(nil, mMock)
	require.Equal(t, 3, eventsMock.EmitCount)
}
