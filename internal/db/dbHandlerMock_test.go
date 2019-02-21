package db

import "github.com/ktt-ol/status2/internal/state"

type DbManagerMock struct {
	LastOpenStatesValues []LastOpenStates

	UpdateOpenStateCount int
	LastPlace            Place
	LastOpenValue        state.OpenValueTs

	UpdateDevicesAndPeopleCount int
	LastDevicesCount            int64
	LastPeopleCount             int64
}

func (dbm *DbManagerMock) GetLastOpenStates() []LastOpenStates {
	return dbm.LastOpenStatesValues
}

func (dbm *DbManagerMock) GetLastDevicesData() *LastDevices {
	panic("implement me")
}

func (dbm *DbManagerMock) GetAllSpaceOpenStates() []OpenState {
	panic("implement me")
}

func (dbm *DbManagerMock) UpdateOpenState(place Place, openValue state.OpenValueTs) {
	dbm.UpdateOpenStateCount++
	dbm.LastPlace = place
	dbm.LastOpenValue = openValue
}

func (dbm *DbManagerMock) UpdateDevicesAndPeople(devicesCount int64, peopleCount int64) {
	dbm.UpdateDevicesAndPeopleCount++
	dbm.LastDevicesCount = devicesCount
	dbm.LastPeopleCount = peopleCount
}
