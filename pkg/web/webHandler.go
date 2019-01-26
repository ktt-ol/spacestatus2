package web

import (
	"github.com/sirupsen/logrus"
	"github.com/ktt-ol/status2/pkg/conf"
	"github.com/ktt-ol/status2/pkg/events"
	"github.com/ktt-ol/status2/pkg/state"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ktt-ol/status2/pkg/db"
	"os"
	"github.com/ktt-ol/status2/pkg/mqtt"
)

var logger = logrus.WithField("where", "web")

func StartWebService(conf conf.WebServiceConf, ev events.EventManager, appState *state.State, dbMgr db.DbManager, mqttMgr *mqtt.MqttManager) {
	// our default is "release"
	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	// use logrus logging
	gin.DisableConsoleColor()
	gin.DefaultWriter = logrus.WithField("where", "gin").WriterLevel(logrus.DebugLevel)
	gin.DefaultErrorWriter = logrus.WithField("where", "gin").WriterLevel(logrus.ErrorLevel)
	router := gin.New()
	router.Use(SimpleNoTimeLogging(), gin.Recovery())

	legacyApiCall(router)

	api := router.Group("/api")
	StatusStream(ev, appState, api.Group("/statusStream"))
	SpaceInfo(appState, api.Group("/spaceInfo"))
	OpenState(appState, api.Group("/openState"))
	OpenStatistics(dbMgr, api.Group("/openStatistics"))

	SwitchPage(conf, mqttMgr, router.Group("/switch"))

	router.Static("/assets", "webUI/assets")
	router.LoadHTMLGlob("webUI/templates/*.html")
	router.StaticFile("/", "webUI/assets/index.html")
	router.StaticFile("/openStats", "webUI/assets/openStats.html")

	addr := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
	err := router.Run(addr);
	if err != nil {
		logger.Error("gin exit", err)
	}
}

func legacyApiCall(router *gin.Engine) {
	router.GET("/status", func(c *gin.Context) {
		c.Request.URL.Path = "/api/spaceInfo"
		router.HandleContext(c)
	})

	router.GET("/status.json", func(c *gin.Context) {
		c.Request.URL.Path = "/api/spaceInfo"
		router.HandleContext(c)
	})
}