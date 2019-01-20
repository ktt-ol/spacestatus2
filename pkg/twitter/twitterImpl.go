package twitter

import (
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/ktt-ol/status2/pkg/conf"
	"github.com/sirupsen/logrus"
)

type TwitterApi interface {
	Send(msg string) error
}

type twitterImpl struct {
	client *twitter.Client
}

func (t *twitterImpl) Send(msg string) error {
	_, _, err := t.client.Statuses.Update(msg, nil)
	return err
}

type MockImpl struct {
	lastMsg string
	tweetCount int
}

func (t *MockImpl) Send(msg string) error {
	t.lastMsg = msg
	t.tweetCount++
	logrus.WithField("where", "twitterMockImpl").Info("MOCK: ", msg)
	return nil
}

func NewTwitterImpl(c conf.TwitterConf) TwitterApi {
	config := oauth1.NewConfig(c.ConsumerKey, c.ConsumerSecret)
	token := oauth1.NewToken(c.AccessTokenKey, c.AccessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	return &twitterImpl{twitter.NewClient(httpClient)}
}

func NewMockingImpl() TwitterApi {
	return &MockImpl{}
}
