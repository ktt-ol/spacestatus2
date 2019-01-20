package main

import (
	"github.com/ktt-ol/status2/pkg/events"
	"github.com/ktt-ol/status2/pkg/conf"
	"github.com/ktt-ol/status2/pkg/mqtt"
	"github.com/ktt-ol/status2/pkg/state"
	"github.com/ktt-ol/status2/pkg/twitter"
	"github.com/ktt-ol/status2/pkg/web"
	"github.com/ktt-ol/status2/pkg/db"
)

const CONFIG_FILE = "config.toml"

func main() {
	config := conf.LoadConfig(CONFIG_FILE)

	conf.SetupLogging(config.Misc)

	st := state.NewDefaultState()
	ev := events.NewEventManager()

	dbMgr := db.NewManager(config.MySql)
	db.NewOpenStatePersistence(dbMgr, ev, st)
	db.NewDevicePersistence(config.MySql, dbMgr, st)

	twitter.NewTwitterHandler(config.Twitter, ev, st)
	mqttMgr := mqtt.NewMqttManager(config.Mqtt, ev, st)

	//fmt.Scanln()
	web.StartWebService(config.Web, ev, st, dbMgr, mqttMgr)
}
