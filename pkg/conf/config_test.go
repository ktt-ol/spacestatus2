package conf

import (
	"testing"
	"github.com/stretchr/testify/require"
)

func Test_Config(t *testing.T) {
	config := LoadConfig("../config.example.toml")

	// test some values

	require.Equal(t, false, config.Misc.DebugLogging)

	require.Equal(t, "tls://server:8883", config.Mqtt.Url)
	require.Equal(t, "/net/devices", config.Mqtt.Topics.Devices)
	require.Equal(t, "/access-control-system/space-state", config.Mqtt.Topics.StateSpace)

	require.Equal(t, "localhost", config.MySql.Host)
	require.Equal(t, 900, config.MySql.SaveDevicesIntervalInSec)

	require.Equal(t, false, config.Twitter.Enabled)
	require.Equal(t, 180, config.Twitter.TwitterdelayInSec)
	require.Equal(t, "?", config.Twitter.AccessTokenKey)

	require.Equal(t, "localhost", config.Web.Host)
}
