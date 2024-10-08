package events

type EventName string

func (en EventName) StrValue() string {
	return string(en)
}

const (
	// the values are used in the web service as parameter, too

	TOPIC_SPACE_OPEN_STATE       EventName = "spaceOpen"
	TOPIC_RADSTELLE_OPEN_STATE   EventName = "radstelleOpen"
	TOPIC_LAB_3D_OPEN_STATE      EventName = "lab3dOpen"
	TOPIC_MACHINING_OPEN_STATE   EventName = "machining"
	TOPIC_WOODWORKING_OPEN_STATE EventName = "woodworking"

	TOPIC_SPACE_DEVICES EventName = "spaceDevices"
	TOPIC_POWER_USAGE   EventName = "powerUsage"
	TOPIC_FREIFUNK      EventName = "freifunk"
	TOPIC_WEATHER       EventName = "weather"

	TOPIC_MQTT                  EventName = "mqtt"
	TOPIC_KEYHOLDER             EventName = "keyholder"
	TOPIC_KEYHOLDER_MACHINING   EventName = "keyholder_machining"
	TOPIC_KEYHOLDER_WOODWORKING EventName = "keyholder_woodworking"

	// temporary pass through for the Hacs app
	TOPIC_BACKDOOR_BOLT_CONTACT EventName = "backdoor"
)
