package state

import "errors"

type OpenValue string

const (
	NONE      OpenValue = "none"
	KEYHOLDER OpenValue = "keyholder"
	MEMBER    OpenValue = "member"
	OPEN      OpenValue = "open"
	OPEN_PLUS OpenValue = "open+"

	// this state don't come from the mqtt, it's calculated based on open and next_open state
	CLOSING OpenValue = "closing"
)

func (h *OpenValue) IsPublicOpen() bool {
	return *h == OPEN || *h == OPEN_PLUS
}

func ParseOpenValue(value string) (OpenValue, error) {
	ov := OpenValue(value)
	switch ov {
	case NONE:
		fallthrough
	case KEYHOLDER:
		fallthrough
	case MEMBER:
		fallthrough
	case OPEN:
		fallthrough
	case OPEN_PLUS:
		return ov, nil
	}

	// legacy state values
	switch value {
	// some clients send this for NONE
	case "":
		fallthrough
	case "closed":
		fallthrough
	case "off":
		return NONE, nil
	case "opened":
		fallthrough
	case "on":
		return OPEN, nil
	}

	return ov, errors.New("Invalid open value: " + value)
}
