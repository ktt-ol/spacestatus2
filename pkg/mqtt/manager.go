package mqtt

import (
	"github.com/eclipse/paho.mqtt.golang"
	"crypto/tls"
	"time"
	"github.com/ktt-ol/status2/pkg/conf"
	"github.com/sirupsen/logrus"
	"crypto/x509"
	"io/ioutil"
	"github.com/ktt-ol/status2/pkg/events"
	"github.com/ktt-ol/status2/pkg/state"
	"strconv"
	"github.com/bep/debounce"
	"encoding/json"
	"github.com/ktt-ol/spaceDevices/pkg/structs"
)

const CLIENT_ID = "status2Go"

var mqttLogger = logrus.WithField("where", "mqtt")

type MqttManager struct {
	client mqtt.Client
	config conf.MqttConf
	events events.EventManager
	state  *state.State
	// internal state to calculate a combined state (with closing)
	lastOpenState     *state.OpenValueTs
	lastOpenStateNext *state.OpenValueTs
	debounceFunc      func(f func())
	//watchDog     *watchDog
}

func NewMqttManager(conf conf.MqttConf, events events.EventManager, appState *state.State) *MqttManager {
	opts := mqtt.NewClientOptions()

	opts.AddBroker(conf.Url)

	if conf.Username != "" {
		opts.SetUsername(conf.Username)
	}
	if conf.Password != "" {
		opts.SetPassword(conf.Password)
	}

	certs := defaultCertPool(conf.CertFile)
	tlsConf := &tls.Config{
		RootCAs: certs,
	}
	opts.SetTLSConfig(tlsConf)

	opts.SetClientID(CLIENT_ID + GenerateRandomString(4))
	opts.SetAutoReconnect(true)
	opts.SetKeepAlive(10 * time.Second)
	opts.SetMaxReconnectInterval(5 * time.Minute)

	debounced, _, _ := debounce.New(500 * time.Millisecond)
	handler := MqttManager{
		config:            conf,
		events:            events,
		state:             appState,
		lastOpenState:     nil,
		lastOpenStateNext: nil,
		debounceFunc:      debounced,
	}

	opts.SetOnConnectHandler(handler.onConnect)
	opts.SetConnectionLostHandler(handler.onConnectionLost)

	handler.client = mqtt.NewClient(opts)
	if tok := handler.client.Connect(); tok.WaitTimeout(5*time.Second) && tok.Error() != nil {
		mqttLogger.WithError(tok.Error()).Fatal("Could not connect to mqtt server.")
	}

	//if conf.WatchDogTimeoutInMinutes > 0 {
	//	mqttLogger.Println("Enable mqtt watch dog, timeout in minutes is", conf.WatchDogTimeoutInMinutes)
	//	handler.watchDog = NewWatchDog(time.Duration(conf.WatchDogTimeoutInMinutes) * time.Minute)
	//}

	return &handler
}

func (h *MqttManager) SendNewSpaceStatus(status state.OpenValue) {
	stLogger := mqttLogger.WithField("newStatus", status)
	stLogger.Info("Sending new space status mqtt value.")
	token := h.client.Publish(h.config.Topics.StateSpace, 0, false, string(status))
	token.WaitTimeout(5 * time.Second)

	if token.Error() != nil {
		stLogger.WithError(token.Error()).Error("Could not send new status.")
	}
}

func (h *MqttManager) onConnect(client mqtt.Client) {
	mqttLogger.Info("connected")
	h.state.Mqtt.Connected = true
	h.events.Emit(events.TOPIC_MQTT)

	h.subscribe(h.config.Topics.SpaceInternalBrokerTopic, func(client mqtt.Client, message mqtt.Message) {
		msg := string(message.Payload())
		mqttLogger.WithField("data", msg).Info("SpaceInternalBrokerTopic")
		h.state.Mqtt.SpaceBrokerOnline = msg == "1"
		h.events.Emit(events.TOPIC_MQTT)
	})

	// closing state + debouncing
	h.subscribe(h.config.Topics.StateSpace, h.onSpaceOpenChange)
	h.subscribe(h.config.Topics.StateSpaceNext, h.onSpaceOpenChange)

	h.subscribeToOpenState(h.config.Topics.StateRadstelle, events.TOPIC_RADSTELLE_OPEN_STATE, h.state.Open.Radstelle)
	h.subscribeToOpenState(h.config.Topics.StateLab3d, events.TOPIC_LAB_3D_OPEN_STATE, h.state.Open.Lab3d)
	h.subscribeToOpenState(h.config.Topics.StateMachining, events.TOPIC_MACHINING_OPEN_STATE, h.state.Open.Machining)

	h.subscribe(h.config.Topics.Devices, h.onDevicesChange)

	h.subscribeToPower(h.config.Topics.EnergyFront, events.TOPIC_POWER_USAGE, h.state.PowerUsage.Front)
	h.subscribeToPower(h.config.Topics.EnergyBack, events.TOPIC_POWER_USAGE, h.state.PowerUsage.Back)


	//token := h.client.Publish("/access-control-system/footest", 0, false, "barbar")
	//token.WaitTimeout(5 * time.Second)

}

func (h *MqttManager) onConnectionLost(client mqtt.Client, err error) {
	mqttLogger.WithError(err).Error("Connection lost.")
	h.state.Mqtt.Connected = false
	h.state.Mqtt.SpaceBrokerOnline = false
	h.events.Emit(events.TOPIC_MQTT)
}

func (h *MqttManager) subscribe(topic string, cb mqtt.MessageHandler) {
	qos := 0
	tok := h.client.Subscribe(topic, byte(qos), cb)
	tok.WaitTimeout(5 * time.Second)

	if tok.Error() != nil {
		mqttLogger.WithField("topic", topic).WithError(tok.Error()).Fatal("Could not subscribe.")
	}
}

// subscribe to an open state change (e.g. radstelle)
// on event does: parse the new open state, change the value in the state and emit the event
func (h *MqttManager) subscribeToOpenState(topic string, eventName events.EventName, openState *state.OpenValueTs) {

	h.subscribe(topic, func(client mqtt.Client, message mqtt.Message) {
		topicLogger := mqttLogger.WithField("topic", topic)

		strMessage := string(message.Payload())
		if strMessage == "" {
			topicLogger.Debug("Empty message.")
			return
		}
		openValue, err := state.ParseOpenValue(strMessage)
		if err != nil {
			topicLogger.WithError(err).Warn("Got invalid open value from mqtt")
			return
		}

		topicLogger.WithField("state", openValue).Info("new open state")

		openState.Value = openValue
		openState.Timestamp = time.Now().Unix()
		h.events.Emit(eventName)
	})
}

// subscribe to a power state change(e.g. front/back)
// on event does: parse the new open state, change the value in the state and emit the event
func (h *MqttManager) subscribeToPower(topic string, eventName events.EventName, powerState *state.PowerValueTs) {

	h.subscribe(topic, func(client mqtt.Client, message mqtt.Message) {
		strMessage := string(message.Payload())

		energy, err := strconv.ParseFloat(strMessage, 64);
		if err != nil {
			mqttLogger.WithError(err).WithField("topic", topic).Warn("Invalid float value for energy front: ", strMessage)
			return
		}

		//mqttLogger.WithFields(logrus.Fields{
		//	"topic": topic,
		//	"state": strMessage,
		//}).Debug("new power state")

		powerState.Value = energy
		powerState.Timestamp = time.Now().Unix()
		h.events.Emit(eventName)
	})
}

func (h *MqttManager) onSpaceOpenChange(client mqtt.Client, message mqtt.Message) {
	topicLogger := mqttLogger.WithField("topic", message.Topic())

	strMessage := string(message.Payload())
	if strMessage == "" {
		topicLogger.Debug("Empty message.")
		return
	}
	openValue, err := state.ParseOpenValue(strMessage)
	if err != nil {
		topicLogger.WithError(err).Warn("Got invalid open value from mqtt")
		return
	}
	topicLogger.WithField("openValue", openValue).Info("onSpaceOpenChange")

	if message.Topic() == h.config.Topics.StateSpace {
		h.lastOpenState = &state.OpenValueTs{Value: openValue, Timestamp: time.Now().Unix()}
		h.debounceFunc(h.newSpaceState)
		return
	}

	if message.Topic() == h.config.Topics.StateSpaceNext {
		h.lastOpenStateNext = &state.OpenValueTs{Value: openValue, Timestamp: time.Now().Unix()}
		h.debounceFunc(h.newSpaceState)
		return
	}

	mqttLogger.Warn("Unexpected topic: ", message.Topic())
}

func (h *MqttManager) onDevicesChange(client mqtt.Client, message mqtt.Message) {
	/*
	{
	  "people": [
		{
		  "name": "Hans",
		  "devices": [
			{
			  "name": "S8",
			  "location": "Space"
			},
			{
			  "name": "T430",
			  "location": "Space"
			}
		  ]
		}
	  ],
	  "peopleCount": 1,
	  "deviceCount": 25,
	  "unknownDevicesCount": 12
	}
	*/

	devices := structs.PeopleAndDevices{}
	if err := json.Unmarshal(message.Payload(), &devices); err != nil {
		logrus.WithField("payload", string(message.Payload())).WithError(err).Warn("Invalid json payload for devices.")
		return
	}

	logrus.Debug("New devices data. ", string(message.Payload()))

	h.state.SpaceDevices.PeopleAndDevices = devices
	h.state.SpaceDevices.Timestamp = time.Now().Unix()
	h.events.Emit(events.TOPIC_SPACE_DEVICES)
}

func (h *MqttManager) newSpaceState() {
	if h.lastOpenState == nil {
		mqttLogger.Error("lastOpenState is not set!")
		return
	}
	mqttLogger.WithField("lastOpenState", h.lastOpenState).WithField("lastOpenStateNext", h.lastOpenStateNext).
		Debug("newSpaceState.")

	if !h.lastOpenState.Value.IsPublicOpen() {
		h.changeOpenState(h.lastOpenState.Value, h.lastOpenState.Timestamp)
		return
	}

	if h.lastOpenStateNext != nil {
		// is the next state close for guests?
		nextValue := h.lastOpenStateNext.Value
		if nextValue == state.NONE || nextValue == state.KEYHOLDER || nextValue == state.MEMBER {
			h.changeOpenState(state.CLOSING, time.Now().Unix())
			return
		}
	}
	// no special closing state
	h.changeOpenState(h.lastOpenState.Value, h.lastOpenState.Timestamp)
}

// changes the state, logs and emits the event
func (h *MqttManager) changeOpenState(value state.OpenValue, timestamp int64) {
	mqttLogger.WithFields(logrus.Fields{
		"state": value,
	}).Info("new SPACE open state")

	h.state.Open.Space.Value = value
	h.state.Open.Space.Timestamp = timestamp
	h.events.Emit(events.TOPIC_SPACE_OPEN_STATE)
}

func defaultCertPool(certFile string) *x509.CertPool {
	if certFile == "" {
		mqttLogger.Debug("No certFile given, using system pool")
		pool, err := x509.SystemCertPool()
		if err != nil {
			mqttLogger.WithError(err).Fatal("Could not create system cert pool.")
		}
		return pool
	}

	fileData, err := ioutil.ReadFile(certFile)
	if err != nil {
		mqttLogger.WithError(err).Fatal("Could not read given cert file.")
	}

	certs := x509.NewCertPool()
	if !certs.AppendCertsFromPEM(fileData) {
		mqttLogger.Fatal("unable to add given certificate to CertPool")
	}

	return certs
}
