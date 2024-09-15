package conf

import (
	"github.com/BurntSushi/toml"
	"github.com/sirupsen/logrus"
)

func LoadConfig(configFile string) TomlConfig {
	logrus.WithField("configFile", configFile).Info("Loading config.")
	config := &TomlConfig{}
	if _, err := toml.DecodeFile(configFile, config); err != nil {
		logrus.WithError(err).Fatal("Could not read config file.")
	}

	return *config
}

type TomlConfig struct {
	Mqtt    MqttConf
	MySql   MySqlConf
	Twitter TwitterConf
	Web     WebServiceConf
	Misc    MiscConf
}

type MqttConf struct {
	Url      string
	Username string
	Password string
	// if empty, the system certificates are used
	CertFile string

	Topics MqttTopicsConf
}

type MqttTopicsConf struct {
	SpaceInternalBrokerTopic string
	Devices                  string

	StateSpace       string
	StateSpaceNext   string
	StateRadstelle   string
	StateLab3d       string
	StateMachining   string
	StateWoodworking string

	EnergyFront     string
	EnergyBack      string
	EnergyMachining string

	KeyholderId              string
	KeyholderName            string
	KeyholderNameMachining   string
	KeyholderNameWoodworking string

	BackdoorBoltContact string
}

type MySqlConf struct {
	Host                     string
	User                     string
	Password                 string
	Database                 string
	SaveDevicesIntervalInSec int
}

type TwitterConf struct {
	Mocking           bool // # if true, it does everthing except the actual tweet. Useful for developing.
	Enabled           bool
	TwitterdelayInSec int // delay tweeting after space state change for this long; it's also the minimum time between two tweets
	// auth
	ConsumerKey       string
	ConsumerSecret    string
	AccessTokenKey    string
	AccessTokenSecret string
}

type WebServiceConf struct {
	Host           string
	Port           int
	SwitchPassword string // to change a status on the /switch page
}

type MiscConf struct {
	DebugLogging bool
	Logfile      string
}
