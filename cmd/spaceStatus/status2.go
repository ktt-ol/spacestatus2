package main

import (
	"github.com/ktt-ol/status2/internal/events"
	"github.com/ktt-ol/status2/internal/conf"
	"github.com/ktt-ol/status2/internal/mqtt"
	"github.com/ktt-ol/status2/internal/state"
	"github.com/ktt-ol/status2/internal/twitter"
	"github.com/ktt-ol/status2/internal/web"
	"github.com/ktt-ol/status2/internal/db"
	"github.com/sirupsen/logrus"
)

const CONFIG_FILE = "config.toml"

var buildVersion = "unkown"

func main() {
	config := conf.LoadConfig(CONFIG_FILE)

	conf.SetupLogging(config.Misc)

	logrus.Info("\n" +
		"-------------\n" +
		"Starting status2\n" +
		"Version: " + buildVersion + "\n" +
		"-------------")

	st := state.NewDefaultState()
	ev := events.NewEventManager()

	dbMgr := db.NewManager(config.MySql)
	db.NewOpenStatePersistence(dbMgr, ev, st)
	db.NewDevicePersistence(config.MySql, dbMgr, st)

	twitter.NewTwitterHandler(config.Twitter, ev, st)
	mqttMgr := mqtt.NewMqttManager(config.Mqtt, ev, st)

	web.StartWebService(config.Web, ev, st, dbMgr, mqttMgr)
}
