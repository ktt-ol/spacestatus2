package db

import (
	"github.com/ktt-ol/status2/pkg/state"
	"github.com/ktt-ol/status2/pkg/conf"
	"time"
)

type DevicePersistence struct {
	dbManager     DbManager
	st            *state.State
	timerInterval time.Duration
	stopChan      chan bool
	ticker        *time.Ticker
}

func NewDevicePersistence(config conf.MySqlConf, dbManager DbManager, st *state.State) *DevicePersistence {
	dp := DevicePersistence{dbManager, st,
		time.Duration(config.SaveDevicesIntervalInSec) * time.Second,
		make(chan bool), nil}
	dp.startTimer()
	return &dp
}

func (dp *DevicePersistence) startTimer() {
	logger.Info("Starting DevicePersistence timer.")
	dp.ticker = time.NewTicker(dp.timerInterval)
	go func() {
		for {
			select {
			case <-dp.ticker.C:
				dp.dbManager.UpdateDevicesAndPeople(int64(dp.st.SpaceDevices.DeviceCount), int64(dp.st.SpaceDevices.PeopleCount))
			case <-dp.stopChan:
				return
			}
		}
	}()
}

func (dp *DevicePersistence) StopTimer() {
	if dp.ticker != nil {
		dp.ticker.Stop()
		dp.ticker = nil
		dp.stopChan <- true
	}
}
