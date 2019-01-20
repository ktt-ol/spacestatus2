package db

import (
	"testing"
	"github.com/ktt-ol/status2/pkg/state"
	"github.com/stretchr/testify/require"
	"github.com/ktt-ol/status2/pkg/conf"
	"time"
)

func Test_DevicePersistence(t *testing.T) {
	dbConf := conf.MySqlConf{SaveDevicesIntervalInSec: 1}
	dbMock := new(DbManagerMock)
	appState := state.NewDefaultState()

	appState.SpaceDevices.DeviceCount = 10
	appState.SpaceDevices.PeopleCount = 2

	dp := NewDevicePersistence(dbConf, dbMock, appState)
	require.Equal(t, 0, dbMock.UpdateDevicesAndPeopleCount)

	waitingTime := 1010
	time.Sleep(time.Duration(waitingTime) * time.Millisecond)
	require.Equal(t, 1, dbMock.UpdateDevicesAndPeopleCount)
	require.Equal(t, int64(10), dbMock.LastDevicesCount)
	require.Equal(t, int64(2), dbMock.LastPeopleCount)

	// changed data
	appState.SpaceDevices.DeviceCount = 11
	appState.SpaceDevices.PeopleCount = 3
	time.Sleep(time.Duration(waitingTime) * time.Millisecond)
	require.Equal(t, 2, dbMock.UpdateDevicesAndPeopleCount)
	require.Equal(t, int64(11), dbMock.LastDevicesCount)
	require.Equal(t, int64(3), dbMock.LastPeopleCount)

	// data hasn't changed
	time.Sleep(time.Duration(waitingTime) * time.Millisecond)
	require.Equal(t, 3, dbMock.UpdateDevicesAndPeopleCount)
	require.Equal(t, int64(11), dbMock.LastDevicesCount)
	require.Equal(t, int64(3), dbMock.LastPeopleCount)

	// stop
	dp.StopTimer()
	appState.SpaceDevices.DeviceCount = 12
	appState.SpaceDevices.PeopleCount = 4
	// nothing should have changed
	time.Sleep(time.Duration(waitingTime) * time.Millisecond)
	require.Equal(t, 3, dbMock.UpdateDevicesAndPeopleCount)
	require.Equal(t, int64(11), dbMock.LastDevicesCount)
	require.Equal(t, int64(3), dbMock.LastPeopleCount)
}
