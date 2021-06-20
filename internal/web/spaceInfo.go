package web

import (
	"github.com/ktt-ol/status2/internal/state"
	"github.com/gin-gonic/gin"
	"time"
	"fmt"
)

// ?spaceOpen=1&radstelleOpen=1&machining=1&spaceDevices=1&powerUsage=1&lab3dOpen=1&mqtt=1
func SpaceInfo(st *state.State, group *gin.RouterGroup) {
	group.GET("", func(c *gin.Context) {

		nowInSeconds := time.Now().Unix()

		data := map[string]interface{}{
			"api":   "0.13",
			"space": "Mainframe",
			"logo":  "https://status.mainframe.io/assets/images/mainframe.png",
			"url":   "https://mainframe.io/",
			"location": map[string]interface{}{
				"address": "Bahnhofsplatz 10, 26122 Oldenburg, Germany",
				"lat":     53.14402,
				"lon":     8.21988,
			},
			"contact": map[string]interface{}{
				"twitter":    "@HackspaceOL",
				"email":      "vorstand@kreativitaet-trifft-technik.de",
				"ml":         "https://mailman.ktt-ol.de/postorius/lists/diskussion.lists.ktt-ol.de/",
				"issue_mail": "hc@kreativitaet-trifft-technik.de",
			},
			"issue_report_channels": [...]string{"issue_mail"},
			"state": map[string]interface{}{
				"open":       st.Open.Space.Value.IsPublicOpen(),
				"lastchange": st.Open.Space.Timestamp,
				"message":    ifElse(st.Open.Space.Value.IsPublicOpen(), "Open!", "Close!"),
				"icon": map[string]interface{}{
					"open":   "https://www.kreativitaet-trifft-technik.de/media/img/mainframe-open.svg",
					"closed": "https://www.kreativitaet-trifft-technik.de/media/img/mainframe-closed.svg",
				},
			},
			"sensors": map[string]interface{}{
				"people_now_present": getPeopleSensor(st.SpaceDevices),
				"network_connections": [...]interface{}{
					map[string]interface{}{
						"value":    st.SpaceDevices.DeviceCount,
						"name":     "deviceCount",
						"location": "Inside",
					}, map[string]interface{}{
						"value":       ifElse(st.Mqtt.SpaceBrokerOnline, 1, 0),
						"name":        "internetStatus",
						"description": "0: no internet connection, 1: everything is fine",
					},
				},
			},
			"power_consumption": [...]interface{}{
				map[string]interface{}{
					"name":        "current consumption front",
					"location":    "Hackspace, front",
					"unit":        "W",
					"value":       st.PowerUsage.Front.Value,
					"description": fmt.Sprintf("Value changed %d sec. ago.", nowInSeconds-st.PowerUsage.Front.Timestamp),
				},
				map[string]interface{}{
					"name":        "current consumption back",
					"location":    "Hackspace, back",
					"unit":        "W",
					"value":       st.PowerUsage.Back.Value,
					"description": fmt.Sprintf("Value changed %d sec. ago.", nowInSeconds-st.PowerUsage.Back.Timestamp),
				},
				map[string]interface{}{
					"name":        "current consumption machining",
					"location":    "Hackspace, machining",
					"unit":        "W",
					"value":       st.PowerUsage.Machining.Value,
					"description": fmt.Sprintf("Value changed %d sec. ago.", nowInSeconds-st.PowerUsage.Machining.Timestamp),
				},

			},
			"feeds": map[string]interface{}{
				"calendar": map[string]interface{}{
					"type": "application/calendar",
					"url":  "https://www.kreativitaet-trifft-technik.de/calendar/ical/markusframer@gmail.com/public/basic.ics",
				},
			},
			"projects": [...]interface{}{
				"https://github.com/ktt-ol/",
			},
		}

		c.JSON(200, data)
	})

	group.GET("/asterisk", func(c *gin.Context) {
		c.Header("cache-control", "no-cache")
		c.String(200, "%d-%d",
			ifElse(st.Open.Space.Value.IsPublicOpen(), 1, 0),
			ifElse(st.Open.Radstelle.Value.IsPublicOpen(), 1, 0),
		)
	})
}

func getPeopleSensor(d *state.SpaceDevicesState) interface{} {
	peoplePresent := map[string]interface{}{
		"value": d.PeopleCount,
	}
	if len(d.People) > 0 {
		names := make([]string, len(d.People))
		for index, person := range d.People {
			names[index] = person.Name
		}
		peoplePresent["names"] = names
	}

	return [...]interface{}{peoplePresent}
}

func ifElse(check bool, ifTrue interface{}, ifFalse interface{}) interface{} {
	if check {
		return ifTrue
	}
	return ifFalse
}
