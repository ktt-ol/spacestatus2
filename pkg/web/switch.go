package web

import (
	"github.com/gin-gonic/gin"
	"github.com/ktt-ol/status2/pkg/mqtt"
	"net/http"
	"github.com/ktt-ol/status2/pkg/conf"
	"github.com/ktt-ol/status2/pkg/state"
)

func SwitchPage(conf conf.WebServiceConf, mqttMgr *mqtt.MqttManager, group *gin.RouterGroup) {

	if conf.SwitchPassword == "" {
		logger.Info("/switch page is disabled, because no password is set.")
		return
	}

	group.GET("", func(c *gin.Context) {
		password := c.Query("password")
		showPwField := password == ""

		c.HTML(http.StatusOK, "switch.html", gin.H{
			"password":    password,
			"showPwField": showPwField,
		})
	})

	group.POST("", func(c *gin.Context) {
		password := c.Query("password")
		if password == "" {
			password = c.PostForm("password")
		}

		if password != conf.SwitchPassword {
			logger.WithField("tried", password).Warn("Invalid switch password!")
			query := c.Request.URL.Query()
			query.Set("wrongPw", "1")
			c.Request.URL.RawQuery = query.Encode()
			c.Redirect(http.StatusSeeOther, c.Request.URL.String())
			return
		}

		action := c.PostForm("action")

		if action == "open" {
			mqttMgr.SendNewSpaceStatus(state.OPEN)
		} else {
			mqttMgr.SendNewSpaceStatus(state.NONE)
		}


		c.Redirect(http.StatusSeeOther, c.Request.RequestURI)
	})

}
