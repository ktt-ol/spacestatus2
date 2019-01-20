package web

import (
	"github.com/gin-gonic/gin"
	"github.com/ktt-ol/status2/pkg/db"
	"time"
	"github.com/sirupsen/logrus"
	"fmt"
)

type entry struct {
	begin *time.Time
	end   *time.Time
}

func (e *entry) String() string {
	return fmt.Sprintf("begin: %s, end: %s", e.begin, e.end)
}

type yearEntries struct {
	Year    int
	Entries [ /*days in year*/ ][ /*a day*/ ][ /*begin delta, duration*/ 2]int64
}

var openStatsLogger = logrus.WithField("where", "OpenStatistics")

const dayInSeconds int64 = 60 * 60 * 24;

func OpenStatistics(dbMgr db.DbManager, group *gin.RouterGroup) {
	group.GET("", func(c *gin.Context) {
		entries := normalizeResults(dbMgr.GetAllSpaceOpenStates())
		if len(entries) == 0 {
			c.JSON(200, nil)
			return
		}

		slots := buildSlots(entries, time.Now().Year())
		fillNilSlots(slots)
		c.JSON(200, slots)
	})
}

type tmp struct {
	// e.g. 2017, 2018
	currentYear int
	// unix timestamp in sec.
	currentYearTs int64
	// stores open times for the next year
	nextYearCarry [][2]int64

	// the final data structure for a single year
	entriesForCurrentYear [ /*days in year*/ ][ /*a day*/ ][ /*begin delta, duration*/ 2]int64

	statsResult []*yearEntries
}

func (note *tmp) addSlot(begin *time.Time, end *time.Time) {

	year, month, day := begin.Date()

	// the day starts with 1, the index with 0
	beginDayIndex := begin.YearDay() - 1
	endDayIndex := end.YearDay() - 1

	if year > note.currentYear {
		note.newYear(year)
	}

	beginDayTs := note.currentYearTs + int64(beginDayIndex)*dayInSeconds

	if beginDayIndex == endDayIndex {
		// the space was closed on the same day
		dayEntry := [2]int64{
			begin.Unix() - beginDayTs, // offset from the day slot
			end.Unix() - begin.Unix(), // duration
		}

		if year > note.currentYear {
			note.nextYearCarry = append(note.nextYearCarry, dayEntry)
		} else {
			note.entriesForCurrentYear[beginDayIndex] = append(note.entriesForCurrentYear[beginDayIndex], dayEntry)
		}

		return
	}

	// the end is at the following day...
	// fill up the current day
	currentDayEnd := time.Date(year, month, day, 23, 59, 59, 0, begin.Location())
	dayEntry := [2]int64{
		begin.Unix() - beginDayTs,           // offset from the day slot
		currentDayEnd.Unix() - begin.Unix(), // duration
	}
	if year > note.currentYear {
		note.entriesForCurrentYear[beginDayIndex] = append(note.entriesForCurrentYear[beginDayIndex], dayEntry)
	} else {
		note.entriesForCurrentYear[beginDayIndex] = append(note.entriesForCurrentYear[beginDayIndex], dayEntry)
	}

	nextDayStart := currentDayEnd.Add(time.Duration(1) * time.Second)
	note.addSlot(&nextDayStart, end)
}

func (note *tmp) newYear(beginYear int) {
	if note.entriesForCurrentYear != nil {
		note.statsResult = append(note.statsResult, &yearEntries{note.currentYear, note.entriesForCurrentYear})
	}

	note.currentYear = beginYear
	note.entriesForCurrentYear = makeYearStructure(beginYear)
	note.currentYearTs = time.Date(note.currentYear, 1, 1, 0, 0, 0, 0, time.UTC).Unix()

	// add the last year carry
	for i := range note.nextYearCarry {
		note.entriesForCurrentYear[0] = append(note.entriesForCurrentYear[0], note.nextYearCarry[i])
	}
}


func buildSlots(entries []*entry, yearNow int) []*yearEntries {
	note := &tmp{nextYearCarry: make([][2]int64, 0), statsResult: make([]*yearEntries, 0, 10)}

	for i := range entries {
		entry := entries[i]
		beginYear := entry.begin.Year()
		if beginYear != note.currentYear {
			note.newYear(beginYear)
		}

		if entry.end == nil {
			openStatsLogger.WithField("entry", entry).Warn("end is nil")
			continue
		}
		if entry.begin.After(*entry.end) {
			openStatsLogger.WithField("entry", entry).Warn("begin is after end")
			continue
		}

		note.addSlot(entry.begin, entry.end)

	} // end for

	// add the last year entries
	note.statsResult = append(note.statsResult, &yearEntries{note.currentYear, note.entriesForCurrentYear})

	trimCurrentYear(note.statsResult, yearNow)

	return note.statsResult
}

func trimCurrentYear(slots []*yearEntries, yearNow int) {
	for i := range slots {
		if slots[i].Year == yearNow {
			lastIndex := len(slots[i].Entries) - 1

			// if the last entry is not len 0, there is nothing to trim
			if len(slots[i].Entries[lastIndex]) > 0 {
				return
			}
			// find backwards the next len > 0
			for ; lastIndex >= 0; lastIndex-- {
				if len(slots[i].Entries[lastIndex]) > 0 {
					// and remove the tail
					slots[i].Entries = slots[i].Entries[0:lastIndex + 1]
					return
				}
			}
		}
	}
}

func makeYearStructure(forYear int) [][][2]int64 {
	days := 365
	if isLeap(forYear) {
		days = 366
	}

	return make([][][2]int64, days, days)
}

// copy from time package
func isLeap(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

/**
 * rules: - use the first 'off' and ignore any following ones - use the first
 * 'on' and ignore any following ones
 *
 * @param sqlResults
 * @returns {Array}
 */
func normalizeResults(openValues []db.OpenState) []*entry {
	// this is an rough estimate
	normalizedEntries := make([]*entry, 0, len(openValues)/2)

	lastIsOpen := false;
	//lastEntry := entry{}
	var lastEntry *entry

	for i, _ := range openValues {
		openValue := &openValues[i]
		isOpen := openValue.Value.IsPublicOpen()
		if lastIsOpen == isOpen {
			// ignore double state (e.g. open -> open+)
			continue;
		}

		if isOpen {
			lastEntry = &entry{
				begin: &openValue.Time,
				end:   nil,
			}
			normalizedEntries = append(normalizedEntries, lastEntry)
		} else {
			lastEntry.end = &openValue.Time
		}

		lastIsOpen = isOpen
	}

	return normalizedEntries;
}

func fillNilSlots(entries []*yearEntries) {
	for i := range entries {
		year := entries[i]
		for x := range year.Entries {
			if year.Entries[x] == nil {
				year.Entries[x] = make([][2]int64, 0)
			}
		}
	}
}
