package twitter

import (
	"github.com/sirupsen/logrus"
	"github.com/ktt-ol/status2/pkg/conf"
	"github.com/ktt-ol/status2/pkg/events"
	"github.com/ktt-ol/status2/pkg/state"
	"github.com/bep/debounce"
	"time"
	"fmt"
)

var logger = logrus.WithField("where", "twitter")

const TWEET_TEMPLATE_OPEN = "%s ist seit %s Uhr geöffnet, kommt vorbei! Details unter https://status.mainframe.io/"
const TWEET_TEMPLATE_CLOSED = "%s ist leider seit %s Uhr geschlossen. Details unter https://status.mainframe.io/"

func getPlaceName(event events.EventName) string {
	switch event {
	case events.TOPIC_SPACE_OPEN_STATE:
		return "Der Mainframe"
	case events.TOPIC_RADSTELLE_OPEN_STATE:
		return "Die Radstelle"
	case events.TOPIC_LAB_3D_OPEN_STATE:
		return "Das 3DLab"
	case events.TOPIC_MACHINING_OPEN_STATE:
		return "Machining"
	default:
		return "?"
	}
}

type TwitterHandler struct {
	config        conf.TwitterConf
	api           TwitterApi
	state         *state.State
	lastStateSend map[events.EventName]state.OpenValueTs
	debounceFuncs map[events.EventName]func(f func())
}

func NewTwitterHandler(config conf.TwitterConf, evManager events.EventManager, appState *state.State) *TwitterHandler {
	twitter := TwitterHandler{
		config:        config,
		state:         appState,
		lastStateSend: make(map[events.EventName]state.OpenValueTs),
		debounceFuncs: make(map[events.EventName]func(f func())),
	}
	if !config.Enabled {
		return &twitter
	}

	logger.Info("Starting twitter module, mocking is ", config.Mocking)
	if config.Mocking {
		twitter.api = NewMockingImpl()
	} else {
		twitter.api = NewTwitterImpl(config)
	}

	//twitter.lastStates = make(map[events.EventName]*state.OpenValueTs)
	//twitter.debounceFuncs = make(map[events.EventName]func(f func()))

	evManager.On(events.TOPIC_SPACE_OPEN_STATE, twitter.onOpenStateChange)
	evManager.On(events.TOPIC_RADSTELLE_OPEN_STATE, twitter.onOpenStateChange)
	evManager.On(events.TOPIC_LAB_3D_OPEN_STATE, twitter.onOpenStateChange)
	evManager.On(events.TOPIC_MACHINING_OPEN_STATE, twitter.onOpenStateChange)

	return &twitter
}

func (t *TwitterHandler) onOpenStateChange(topic events.EventName) {
	openValueTs, err := t.state.Open.OpenStateForEvent(topic)
	if err != nil {
		logger.WithField("topic", topic).WithError(err).Error("Invalid event or state")
		return
	}

	t.updateStateAndTweetDebounced(topic, openValueTs)
}

func (t *TwitterHandler) updateStateAndTweetDebounced(topic events.EventName, openValueTs *state.OpenValueTs) {
	makeMsgAndSend := func() {
		// get last state
		lastState, ok := t.lastStateSend[topic]
		// update last state
		t.lastStateSend[topic] = *openValueTs
		if !ok {
			logger.WithField("topic", topic).Warn("No last open state found for topic.")
			return
		} else {
			// any changes for the public?
			if isOpenToPublic(lastState.Value) == isOpenToPublic(openValueTs.Value) {
				logger.WithFields(logrus.Fields{
					"event": topic,
					"state": openValueTs.Value,
				}).Info("I don't tweet the same status twice.")
				return
			}
		}


		template := TWEET_TEMPLATE_CLOSED
		if isOpenToPublic(openValueTs.Value) {
			template = TWEET_TEMPLATE_OPEN
		}
		ts := time.Unix(openValueTs.Timestamp, 0)
		msg := fmt.Sprintf(template, getPlaceName(topic), ts.Format("15:04"))
		logger.WithField("msg", msg).Debug("Sending tweet.")
		err := t.api.Send(msg)
		if err != nil {
			logger.WithError(err).Error("Error sending tweet")
		}
	}

	_, ok := t.lastStateSend[topic]
	if !ok {
		// app start case, setting the first state and stop here
		t.lastStateSend[topic] = *openValueTs
		return;
	}


	if t.config.TwitterdelayInSec == 0 {
		// there is no delay configured, send immediately
		makeMsgAndSend()
		return
	}

	// get the cached debounced wrapper or create it
	debounceFunc, ok := t.debounceFuncs[topic]
	if !ok {
		debounceFunc, _, _ = debounce.New(time.Duration(t.config.TwitterdelayInSec) * time.Second)
		t.debounceFuncs[topic] = debounceFunc
	}

	debounceFunc(makeMsgAndSend)
}

// Converts our various states to a simple true/false. True if the current state should be shown as 'open' for the public.
func isOpenToPublic(openState state.OpenValue) bool {
	return openState == state.OPEN || openState == state.OPEN_PLUS
}
