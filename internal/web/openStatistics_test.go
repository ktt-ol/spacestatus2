package web

import (
	"testing"
	"github.com/ktt-ol/status2/internal/db"
	"github.com/ktt-ol/status2/internal/state"
	"time"
	"github.com/stretchr/testify/require"
)

func Test_normalizeResults(t *testing.T) {

	start := mkTestTime(2012, 1, 4, 1, 10)
	end := mkTestTime(2012, 1, 4, 2, 40)

	start2 := mkTestTime(2012, 1, 5, 3, 11)
	testInput := []db.OpenState{
		db.OpenState{state.OPEN, *start},
		db.OpenState{state.OPEN, *mkTestTime(2012, 1, 4, 1, 20)},
		db.OpenState{state.OPEN_PLUS, *mkTestTime(2012, 1, 4, 1, 30)},

		db.OpenState{state.NONE, *end},
		db.OpenState{state.NONE, *mkTestTime(2012, 1, 4, 2, 50)},

		db.OpenState{state.OPEN, *start2},
	}

	result := normalizeResults(testInput)

	require.Equal(t, 2, len(result))
	require.Equal(t, start, result[0].begin)
	require.Equal(t, end, result[0].end)
	require.Equal(t, start2, result[1].begin)
	require.Nil(t, result[1].end)
}

func Test_buildSlots(t *testing.T) {
	testEntries := []*entry{
		&entry{mkTestTime(2016, 12, 30, 15, 17), mkTestTime(2016, 12, 31, 1, 43)},
		&entry{mkTestTime(2016, 12, 31, 13, 05), mkTestTime(2017, 01, 01, 04, 14)},
		&entry{mkTestTime(2017, 01, 01, 16, 44), mkTestTime(2017, 01, 02, 04, 42)},
		&entry{mkTestTime(2017, 01, 02, 15, 27), mkTestTime(2017, 01, 02, 23, 22)},
		// test for an opening time longer than 24 hours
		&entry{mkTestTime(2017, 01, 05, 18, 03), mkTestTime(2017, 01, 07, 10, 42)},
	}

	slots := buildSlots(testEntries, 2018)

	require.Equal(t, 2, len(slots))
	y2016 := slots[0]
	y2017 := slots[1]
	require.Equal(t, 2016, y2016.Year)
	require.Equal(t, 2017, y2017.Year)

	require.Equal(t, 366, len(y2016.Entries)) // is a leap year
	require.Equal(t, 365, len(y2017.Entries))

	// test for correct entries per  day
	require.Equal(t, 1, len(y2016.Entries[364])) // 1 entry at 30.12.2016
	require.Equal(t, 2, len(y2016.Entries[365])) // 2 entries at 31.12.2016
	require.Equal(t, 2, len(y2017.Entries[0]))   // 2 entries at 01.01.2017
	require.Equal(t, 2, len(y2017.Entries[1]))   // 2 entries at 02.01.2017
	require.Equal(t, 0, len(y2017.Entries[2]))   // 0 entries at 03.01.2017

	require.Equal(t, 0, len(y2017.Entries[3]))
	require.Equal(t, 1, len(y2017.Entries[4]))
	require.Equal(t, 1, len(y2017.Entries[5]))
	require.Equal(t, 1, len(y2017.Entries[6]))

	// some validations about the values
	validateEntries(t, y2016)
	validateEntries(t, y2017)
}

func Test_buildSlots_no_last_empty_slots(t *testing.T) {
	lastDate := mkTestTime(2018, 11, 03, 22, 42)
	testEntries := []*entry{
		&entry{mkTestTime(2018, 11, 02, 10, 27), mkTestTime(2018, 11, 02, 23, 22)},
		&entry{mkTestTime(2018, 11, 03, 18, 03), lastDate},
	}

	slots := buildSlots(testEntries, 2018)
	require.Equal(t, 1, len(slots))

	require.Equal(t, lastDate.YearDay(), len(slots[0].Entries))
}

func validateEntries(t *testing.T, yearData *yearEntries) {
	secondsInDay := int64(60 * 60 * 24)
	for a := range yearData.Entries {
		for b := range yearData.Entries[a] {
			require.True(t, yearData.Entries[a][b][0] < secondsInDay, "Index %d,%d, value %d", a, b, yearData.Entries[a][b][0])
			require.True(t, yearData.Entries[a][b][1] <= secondsInDay, "Index %d,%d, value %d", a, b, yearData.Entries[a][b][0])
		}
	}
}

func mkTestTime(year int, month time.Month, day, hour, min int) *time.Time {
	date := time.Date(year, month, day, hour, min, 0, 0, time.UTC)
	return &date
}
