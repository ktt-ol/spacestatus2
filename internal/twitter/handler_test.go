package twitter

import (
	"testing"
	"github.com/ktt-ol/status2/internal/test"
	"github.com/ktt-ol/status2/internal/state"
	"github.com/ktt-ol/status2/internal/conf"
	"github.com/stretchr/testify/require"
	"github.com/ktt-ol/status2/internal/events"
	"time"
)

func setupObjects(t *testing.T, twitterdelayInSec int) (*state.State, *TwitterHandler, *MockImpl) {
	eventsMock := new(test.EventManagerMock)
	appState := state.NewDefaultState()
	twitterConf := conf.TwitterConf{Enabled: true, Mocking: true, TwitterdelayInSec: twitterdelayInSec}

	twitt := NewTwitterHandler(twitterConf, eventsMock, appState)
	mockImpl, ok := twitt.api.(*MockImpl)
	if !ok {
		t.Fatal("Not the mock impl")
	}
	require.Equal(t, 0, mockImpl.tweetCount)

	return appState, twitt, mockImpl
}

func Test_disabledByConfig(t *testing.T) {
	eventsMock := new(test.EventManagerMock)
	appState := state.NewDefaultState()
	twitterConf := conf.TwitterConf{Enabled: false, Mocking: true}

	NewTwitterHandler(twitterConf, eventsMock, appState)
	require.Equal(t, 0, eventsMock.OnCount)

	//mMock := new(test.MessageMock)
}

func Test_sendStatusOnlyOnce(t *testing.T) {
	appState, twitt, mockImpl := setupObjects(t, 0)

	// simulate the first retained states
	appState.Open.Space.Value = state.NONE
	appState.Open.Lab3d.Value = state.OPEN
	twitt.onOpenStateChange(events.TOPIC_SPACE_OPEN_STATE)
	twitt.onOpenStateChange(events.TOPIC_LAB_3D_OPEN_STATE)
	require.Equal(t, 0, mockImpl.tweetCount)

	appState.Open.Space.Value = state.OPEN
	twitt.onOpenStateChange(events.TOPIC_SPACE_OPEN_STATE)
	require.Equal(t, 1, mockImpl.tweetCount)

	twitt.onOpenStateChange(events.TOPIC_SPACE_OPEN_STATE)
	require.Equal(t, 1, mockImpl.tweetCount)

	appState.Open.Space.Value = state.OPEN_PLUS
	twitt.onOpenStateChange(events.TOPIC_SPACE_OPEN_STATE)
	require.Equal(t, 1, mockImpl.tweetCount)

	appState.Open.Space.Value = state.NONE
	twitt.onOpenStateChange(events.TOPIC_SPACE_OPEN_STATE)
	require.Equal(t, 2, mockImpl.tweetCount)

	appState.Open.Space.Value = state.MEMBER
	twitt.onOpenStateChange(events.TOPIC_SPACE_OPEN_STATE)
	require.Equal(t, 2, mockImpl.tweetCount)

	// different topic
	appState.Open.Lab3d.Value = state.MEMBER
	twitt.onOpenStateChange(events.TOPIC_LAB_3D_OPEN_STATE)
	require.Equal(t, 3, mockImpl.tweetCount)
}

func Test_debounce(t *testing.T) {
	appState, twitt, mockImpl := setupObjects(t, 1)

	// simulate the first retained states
	appState.Open.Space.Value = state.NONE
	twitt.onOpenStateChange(events.TOPIC_SPACE_OPEN_STATE)
	// need to sleep for the debounce
	time.Sleep(time.Duration(1100 * time.Millisecond))
	require.Equal(t, 0, mockImpl.tweetCount)


	appState.Open.Space.Value = state.OPEN
	twitt.onOpenStateChange(events.TOPIC_SPACE_OPEN_STATE)
	// should be zero, because of the debounce
	require.Equal(t, 0, mockImpl.tweetCount)
	time.Sleep(time.Duration(1100 * time.Millisecond))
	// timeout
	require.Equal(t, 1, mockImpl.tweetCount)

	// changing the topic fast, the debounce should avoid tweeting
	appState.Open.Space.Value = state.NONE
	twitt.onOpenStateChange(events.TOPIC_SPACE_OPEN_STATE)
	require.Equal(t, 1, mockImpl.tweetCount)
	appState.Open.Space.Value = state.OPEN
	twitt.onOpenStateChange(events.TOPIC_SPACE_OPEN_STATE)
	require.Equal(t, 1, mockImpl.tweetCount)
	appState.Open.Space.Value = state.NONE
	twitt.onOpenStateChange(events.TOPIC_SPACE_OPEN_STATE)
	require.Equal(t, 1, mockImpl.tweetCount)
	appState.Open.Space.Value = state.OPEN
	twitt.onOpenStateChange(events.TOPIC_SPACE_OPEN_STATE)
	require.Equal(t, 1, mockImpl.tweetCount)
	time.Sleep(time.Duration(1100 * time.Millisecond))
	// no tweet, because of the same end state
	require.Equal(t, 1, mockImpl.tweetCount)

	// fast change with different end state
	appState.Open.Space.Value = state.NONE
	twitt.onOpenStateChange(events.TOPIC_SPACE_OPEN_STATE)
	require.Equal(t, 1, mockImpl.tweetCount)
	appState.Open.Space.Value = state.OPEN
	twitt.onOpenStateChange(events.TOPIC_SPACE_OPEN_STATE)
	require.Equal(t, 1, mockImpl.tweetCount)
	appState.Open.Space.Value = state.NONE
	twitt.onOpenStateChange(events.TOPIC_SPACE_OPEN_STATE)
	require.Equal(t, 1, mockImpl.tweetCount)
	time.Sleep(time.Duration(1100 * time.Millisecond))
	// no tweet, because of the same end state
	require.Equal(t, 2, mockImpl.tweetCount)
}

func Test_skipFirstStatus(t *testing.T) {
	// test that the first status won't be tweeted
	appState, twitt, mockImpl := setupObjects(t, 0)

	appState.Open.Space.Value = state.OPEN
	twitt.onOpenStateChange(events.TOPIC_SPACE_OPEN_STATE)
	require.Equal(t, 0, mockImpl.tweetCount)

	appState.Open.Lab3d.Value = state.OPEN
	twitt.onOpenStateChange(events.TOPIC_LAB_3D_OPEN_STATE)
	require.Equal(t, 0, mockImpl.tweetCount)

	appState.Open.Machining.Value = state.OPEN
	twitt.onOpenStateChange(events.TOPIC_MACHINING_OPEN_STATE)
	require.Equal(t, 0, mockImpl.tweetCount)


	// but: if we have a status change within the first twitter delay time, that change must be tweeted
	appState, twitt, mockImpl = setupObjects(t, 1)

	// app starts and get directly the first (retained) status (OPEN)
	appState.Open.Space.Value = state.OPEN
	twitt.onOpenStateChange(events.TOPIC_SPACE_OPEN_STATE)
	// should not be tweeted (because of the first change AND the delay)
	time.Sleep(time.Duration(100 * time.Millisecond))
	require.Equal(t, 0, mockImpl.tweetCount)
	time.Sleep(time.Duration(100 * time.Millisecond))
	// some closes the space for the public
	appState.Open.Space.Value = state.MEMBER
	twitt.onOpenStateChange(events.TOPIC_SPACE_OPEN_STATE)
	time.Sleep(time.Duration(100 * time.Millisecond))
	// no change, because of the delay
	require.Equal(t, 0, mockImpl.tweetCount)
	time.Sleep(time.Duration(1000 * time.Millisecond))
	// tweet delay is finished
	require.Equal(t, 1, mockImpl.tweetCount)
}

