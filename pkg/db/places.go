package db

import (
	"github.com/ktt-ol/status2/pkg/events"
	"errors"
)

type Place string

func (p Place) StrValue() string {
	return string(p)
}

const (
	PLACE_SPACE     Place = "space"
	PLACE_RADSTELLE Place = "radstelle"
	PLACE_LAB3D     Place = "lab3d"
	PLACE_MACHINING Place = "machining"
)

var validPlaces = [...]Place{PLACE_SPACE, PLACE_RADSTELLE, PLACE_LAB3D, PLACE_MACHINING}

func IsValidPlace(place Place) bool {
	for _, valid := range validPlaces {
		if place == valid {
			return true
		}
	}

	return false
}

func TopicToPlace(topic events.EventName) (Place, error) {
	switch topic {
	case events.TOPIC_SPACE_OPEN_STATE:
		return PLACE_SPACE, nil
	case events.TOPIC_RADSTELLE_OPEN_STATE:
		return PLACE_RADSTELLE, nil
	case events.TOPIC_LAB_3D_OPEN_STATE:
		return PLACE_LAB3D, nil
	case events.TOPIC_MACHINING_OPEN_STATE:
		return PLACE_MACHINING, nil
	default:
		return "", errors.New("Not an open state event: " + topic.StrValue())
	}
}
