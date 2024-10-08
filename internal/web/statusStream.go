package web

import (
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ktt-ol/status2/internal/events"
	"github.com/ktt-ol/status2/internal/state"
)

type ssEvent struct {
	name string
	data interface{}
}

// ?spaceOpen=1&radstelleOpen=1&machining=1&spaceDevices=1&powerUsage=1&lab3dOpen=1&mqtt=1
func StatusStream(ev events.EventManager, appState *state.State, group *gin.RouterGroup) {
	group.GET("", func(c *gin.Context) {

		// the default gin logger logs only at the request END, but this request is a stream
		logger.Debug("Starting statusStream: ", c.ClientIP(), " | ", c.Request.URL.RawQuery)

		// a small buffer to avoid getting the warning too early
		msgChannel := make(chan ssEvent, 5)

		registrations := make([]events.RegistrationId, 0, 8)
		defer func() {
			for _, token := range registrations {
				ev.Remove(token)
			}
		}()

		sendAndRegister := func(topic events.EventName, dataProvider func() interface{}) {
			if c.Request.URL.Query().Get(topic.StrValue()) == "1" {
				c.SSEvent(topic.StrValue(), dataProvider())
				registrations = append(registrations,
					ev.On(topic, func(topic events.EventName) {
						eventData := ssEvent{
							name: topic.StrValue(),
							data: dataProvider(),
						}
						select {
						case msgChannel <- eventData:
							// everything ok
						default:
							logger.Warn("No one there to get the msg.")
							// maybe we should close this channel here?
						}
					}))
			}
		}

		// sends the initial requested states and registers for events
		// those events will be written to the
		c.Stream(func(w io.Writer) bool {

			sendAndRegister(events.TOPIC_MQTT, func() interface{} {
				return appState.Mqtt
			})

			sendAndRegister(events.TOPIC_KEYHOLDER, func() interface{} {
				return appState.Open.Keyholder
			})
			sendAndRegister(events.TOPIC_KEYHOLDER_MACHINING, func() interface{} {
				return appState.Open.KeyholderMachining
			})
			sendAndRegister(events.TOPIC_KEYHOLDER_WOODWORKING, func() interface{} {
				return appState.Open.KeyholderWoodworking
			})

			sendAndRegister(events.TOPIC_SPACE_OPEN_STATE, func() interface{} {
				return appState.Open.Space
			})
			sendAndRegister(events.TOPIC_RADSTELLE_OPEN_STATE, func() interface{} {
				return appState.Open.Radstelle
			})
			sendAndRegister(events.TOPIC_LAB_3D_OPEN_STATE, func() interface{} {
				return appState.Open.Lab3d
			})
			sendAndRegister(events.TOPIC_MACHINING_OPEN_STATE, func() interface{} {
				return appState.Open.Machining
			})
			sendAndRegister(events.TOPIC_WOODWORKING_OPEN_STATE, func() interface{} {
				return appState.Open.Woodworking
			})

			sendAndRegister(events.TOPIC_SPACE_DEVICES, func() interface{} {
				return appState.SpaceDevices
			})

			sendAndRegister(events.TOPIC_POWER_USAGE, func() interface{} {
				return appState.PowerUsage
			})
			sendAndRegister(events.TOPIC_FREIFUNK, func() interface{} {
				return appState.Freifunk
			})

			sendAndRegister(events.TOPIC_BACKDOOR_BOLT_CONTACT, func() interface{} {
				return appState.Backdoor
			})

			return false
		})

		sendKeepAliveTicker := time.NewTicker(time.Minute * 10)
		// this seems to be needed, to avoid panics. If not used, the stream could be stuck in the sendKeepAliveTimer around 10 minutes
		// and this might lead to panics.
		connectionTicker := time.NewTicker(time.Second)

		defer func() {
			close(msgChannel)
			sendKeepAliveTicker.Stop()
			connectionTicker.Stop()
		}()

		c.Stream(func(w io.Writer) bool {
			select {
			case event := <-msgChannel:
				c.SSEvent(event.name, event.data)
			case <-sendKeepAliveTicker.C:
				c.SSEvent("keepalive", "")
			case <-connectionTicker.C:
			}

			// stream, until the client disconnects
			return true
		})
	})
}
