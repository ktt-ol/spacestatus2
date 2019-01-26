package db

import (
	"github.com/ktt-ol/status2/internal/conf"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"fmt"
	"github.com/sirupsen/logrus"
	"time"
	"github.com/ktt-ol/status2/internal/state"
)

var logger = logrus.WithField("where", "db")

type DbManager interface {
	GetLastOpenStates() []LastOpenStates
	GetLastDevicesData() *LastDevices
	GetAllSpaceOpenStates() []OpenState
	UpdateOpenState(place Place, openValue state.OpenValueTs)
	UpdateDevicesAndPeople(devicesCount int64, peopleCount int64)
}

type dbManager struct {
	db *sql.DB
}

type LastOpenStates struct {
	Place Place
	State state.OpenValueTs
}

type LastDevices struct {
	Devices   int64
	People    int64
	Timestamp time.Time
}

type OpenState struct {
	Value state.OpenValue
	Time  time.Time
}

func NewManager(config conf.MySqlConf) DbManager {
	connectionString := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=true", config.User, config.Password, config.Host, config.Database)
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		logger.Fatal("Can't connect to db.", err)
	}

	return &dbManager{db: db}
}

func (db *dbManager) GetLastOpenStates() []LastOpenStates {
	const stmt = `SELECT place, state, timestamp FROM spacestate a
	inner join (SELECT max(id) as id FROM spacestate group by place) m
	on a.id = m.id`

	rows, err := db.db.Query(stmt)
	if err != nil {
		logger.Fatal("Error during select", err)
	}
	defer rows.Close()

	states := make([]LastOpenStates, 0, 4) // expecting 4 places
	for rows.Next() {
		var place Place
		var openValueStr string
		var ts time.Time

		err = rows.Scan(&place, &openValueStr, &ts)
		basicErrorCheck(err)
		openValue, err := state.ParseOpenValue(openValueStr)
		basicErrorCheck(err)

		los := LastOpenStates{Place: place, State: state.OpenValueTs{Value: openValue, Timestamp: ts.Unix()}}
		states = append(states, los)
	}

	return states
}

func (db *dbManager) GetLastDevicesData() *LastDevices {
	const stmt = `select devices, people, ts from devices order by ts desc limit 1`

	ld := LastDevices{}
	err := db.db.QueryRow(stmt).Scan(&ld.Devices, &ld.People, &ld.Timestamp)
	if err != nil {
		logger.Fatal(err)
	}

	return &ld
}

func (db *dbManager) GetAllSpaceOpenStates() []OpenState {
	//const stmt = "SELECT state, timestamp FROM spacestate where place = 'space' and `timestamp` > '2016-12-30' and `timestamp` < '2017-01-05' ORDER BY id asc"
	const stmt = "SELECT state, timestamp FROM spacestate where place = 'space' ORDER BY id asc"

	rows, err := db.db.Query(stmt)
	basicErrorCheck(err)
	defer rows.Close()

	result := make([]OpenState, 0, 100)
	for rows.Next() {
		var openValueStr string
		var ts time.Time

		err = rows.Scan(&openValueStr, &ts)
		basicErrorCheck(err)
		if openValueStr == "closing" {
			openValueStr = "open"
		}
		openValue, err := state.ParseOpenValue(openValueStr)
		if err != nil {
			//logger.WithField("value", openValueStr).Debug("Ignoring open value")
			continue
		}

		result = append(result, OpenState{openValue, ts})
	}

	return result
}

func (db *dbManager) UpdateOpenState(place Place, openValue state.OpenValueTs) {
	//if !IsValidPlace(place) {
	//	logger.WithField("place", place).Fatal("Invalid place.")
	//}

	stmt, err := db.db.Prepare("INSERT INTO spacestate (state, place, timestamp) VALUES (?, ?, ?)")
	basicErrorCheck(err)

	_, err = stmt.Exec(openValue.Value, place, time.Unix(openValue.Timestamp, 0))
	basicErrorCheck(err)
}

func (db *dbManager) UpdateDevicesAndPeople(devicesCount int64, peopleCount int64) {
	stmt, err := db.db.Prepare("INSERT INTO devices (devices, people, ts) VALUES (?, ?, NOW())")
	basicErrorCheck(err)

	_, err = stmt.Exec(devicesCount, peopleCount)
	basicErrorCheck(err)
}

func basicErrorCheck(err error) {
	if err != nil {
		logger.Fatal(err)
	}
}
