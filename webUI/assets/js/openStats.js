function openStatsInit() {
    var request = new XMLHttpRequest();
    request.open("GET", "/api/openStatistics", true);
    request.onreadystatechange = function () {
        if (request.readyState !== 4) {
            return;
        }
        if (request.status !== 200) {
            console.error('Error response for open stats: ', request.status);
            return;
        }
        var data = prepareData(JSON.parse(request.responseText));
        new Graph().makeGraph(data);
        new WeeklyStats().start('weeklyStatsSelect', 'weeklyStatsBody', data);
        new YearlyStats().start('yearlyStatsBody', data);

    };
    request.onerror = function (e) {
        // There was a connection error of some sort
        console.error('connection error', e);
    };
    request.send();
}


var MILLIS_PER_DAY = 1000 * 60 * 60 * 24;

var TimeUtils = new function () {
    this.toHours = function (sec) {
        return (sec / 60 / 60).toFixed(2);
    };

    this.formatAsTime = function (hourFormatBase60) {
        return hourFormatBase60.h + ':' + hourFormatBase60.m + ' Uhr';
    };

    this.formatAsDuration = function (hourFormatBase60) {
        return hourFormatBase60.h + ' Stunden, ' + hourFormatBase60.m + ' Minuten.';
    };

    this.getHourFormatBase60 = function (/* floating point */hour) {
        var hourValue = parseInt(hour, 10).toString();
        if (hourValue.length === 1) {
            hourValue = '0' + hourValue;
        }
        var minutePart = ((hour * 100) % 100);
        var minuteValue = parseInt(minutePart * 60 / 100, 10).toString();
        if (minuteValue.length === 1) {
            minuteValue = '0' + minuteValue;
        }
        return {
            h: hourValue,
            m: minuteValue
        };
    };
}();

function prepareData(data) {
    var chartData = [];
    data.forEach(function (yearData) {
        yearData.Entries.forEach(function (daySlots, slotIndex) {
            var ts = new Date(yearData.Year, 0, 1).getTime() + (MILLIS_PER_DAY * slotIndex);
            var slotDuration = 0;

            // if we don't have any entries for this day, we crate an empty entry
            if (daySlots.length === 0) {
                chartData.push({
                    date: ts,
                    open: 0,
                    close: 0,
                    duration: 0,
                    durationInSec: 0
                });
            } else {
                var dayEntries = [];

                // for every open/duration entry
                daySlots.forEach(function (entry, entryIndex) {
                    slotDuration += entry[1];
                    var close = entry[0] + entry[1];
                    dayEntries.push({
                        date: ts,
                        open: TimeUtils.toHours(entry[0]),
                        close: TimeUtils.toHours(close)
                    });
                });

                dayEntries.forEach(function (entry) {
                    entry.duration = TimeUtils.toHours(slotDuration);
                    entry.durationInSec = slotDuration;
                    chartData.push(entry);
                });

            }
        });
    });
    return chartData;
}

function Graph() {
    // the last item to hover on (performance optimization for the balloon)
    var lastHoverItem = null;
    var ballonTextCache = '';

    function onRollOver(event) {
        var balloon = event.chart.balloon;
        var item = event.item;
        if (lastHoverItem !== event.index) {
            // data update is needed
            lastHoverItem = event.index;
            balloon.setPosition(event.item.x + 60, event.item.y);
            var total = TimeUtils.getHourFormatBase60(item.values.value);
            ballonTextCache =
                'Geöffnet: ' + TimeUtils.formatAsTime(TimeUtils.getHourFormatBase60(item.values.open)) +
                '<br>Geschlossen: ' + TimeUtils.formatAsTime(TimeUtils.getHourFormatBase60(item.values.close)) +
                '<br>Insgesamt offen: \n' + TimeUtils.formatAsDuration(total);
        }
        balloon.showBalloon(ballonTextCache);
    }

    this.makeGraph = function (data) {
        var chart = AmCharts.makeChart("chartdiv", {
            "type": "serial",
            "listeners": [{
                "event": "dataUpdated",
                "method": function (eventData) {
                    // different zoom methods can be used - zoomToIndexes, zoomToDates,
                    // zoomToCategoryValues
                    eventData.chart.zoomToIndexes(data.length - 40, data.length - 1);
                }
            }, {
                "event": "rollOverGraphItem",
                "method": onRollOver
            }],
            "pathToImages": 'assets/images/amchart/',
            // "theme": "light",
            "marginRight": 80,
            "marginTop": 24,
            "dataProvider": data,
            "valueAxes": [{
                "maximum": 24,
                "minimum": 0,
            }],
            "graphs": [
                {
                    "showBalloon": false,
                    "showBalloonAt": "open",
                    // "balloonText": "Open:<b>[[open]]</b><br>Low:<b>[[low]]</b><br>High:<b>[[high]]</b><br>Close:<b>[[close]]</b><br>",

                    "fillColors": "#7f8da9",
                    "lineColor": "#7f8da9",
                    "lineAlpha": 1,
                    "fillAlphas": 0.9,
                    "negativeFillColors": "#db4c3c",
                    "negativeLineColor": "#db4c3c",
                    "type": "candlestick",

                    "highField": "open",
                    "lowField": "close",
                    "closeField": "close",
                    "openField": "open",

                    "valueField": "duration",
                }
            ],
            "chartCursor": {
                "cursorPosition": 'mouse',
                "categoryBalloonDateFormat": 'DD MMMM YYYY'
            },
            "categoryField": "date",
            "categoryAxis": {
                "parseDates": true,
                "minPeriod": 'DD',
            },
            "chartScrollbar": {
                "backgroundAlpha": 0.1,
                "backgroundColor": "#868686",
                "selectedBackgroundColor": "#67b7dc",
                "selectedBackgroundAlpha": 1,
            },
        });
    };
}

function WeeklyStats() {
    var WEEK_DAYS = ['Montag', 'Dienstag', 'Mittwoch', 'Donnerstag', 'Freitag', 'Samstag', 'Sonntag'];

// find the last date except today
    function findStartPosition(chartData) {
        var dayBackCounter = 1;
        var date = 0;
        for (var i = chartData.length - 1; i >= 0; i--) {
            if (date !== chartData[i].date) {
                date = chartData[i].date;
                if (dayBackCounter <= 0) {
                    return i;
                }
                dayBackCounter--;
            }
        }
    }

    function decrementDoWP(pointer) {
        // add extra 7 to avoid negative results
        return (pointer - 1 + 7) % 7;
    }

// use monday instead of sunday as start of week
    function startOfWeekCorrection(dayOfWeek) {
        return (dayOfWeek - 1 + 7) % 7;
    }


    function createWeekStats(chartData, weeks) {
        var i;
        // working on the already prepared chartData

        // go back until we have enough days, specified by "weeks"

        var startPosition = findStartPosition(chartData);
        var lastDate = new Date(chartData[startPosition].date);
        // from 0 to 6; the index of the current ay in dayOfWeeks
        var dayOfWeeksPointer = startOfWeekCorrection(lastDate.getDay());

        // monday == 0
        var dayOfWeeks = [0, 0, 0, 0, 0, 0, 0];
        var lastTs = 0;
        var daysCount = 0;
        for (i = startPosition; i >= 0 && daysCount < weeks * 7; i--) {
            // did we had this day already?
            if (chartData[i].date === lastTs) {
                continue;
            }
            lastTs = chartData[i].date;
            // check for correct day
            if (dayOfWeeksPointer !== startOfWeekCorrection(new Date(chartData[i].date).getDay())) {
                console.log(dayOfWeeksPointer);
                console.log(new Date(chartData[i].date));
                throw new Error('Invalid state!');
            }
            // add total open time for this day
            dayOfWeeks[dayOfWeeksPointer] += chartData[i].durationInSec;

            // move the pointer to the previous day
            dayOfWeeksPointer = decrementDoWP(dayOfWeeksPointer);
            daysCount++;
        }

        // create the average by dividing through the amount of weeks
        for (i = 0; i < dayOfWeeks.length; i++) {
            dayOfWeeks[i] /= weeks;
        }

        var totalSum = 0;
        var result = [];
        dayOfWeeks.forEach(function (value, index) {
            totalSum += value;
            var openTime = TimeUtils.getHourFormatBase60(TimeUtils.toHours(value));
            result.push({
                name: WEEK_DAYS[index],
                stat: TimeUtils.formatAsDuration(openTime)
            });
        });
        var totalSumFormatted = TimeUtils.formatAsDuration(TimeUtils.getHourFormatBase60(TimeUtils.toHours(totalSum / 7)));
        result.push({
            name: 'Durchschnitt über alle Tage',
            stat: totalSumFormatted
        });

        return result;
    }

    this.start = function (selectId, tBodyId, chartData) {
        var select = document.getElementById(selectId);
        var tBody = document.getElementById(tBodyId);

        function update() {
            var weeks = select.options[select.selectedIndex].value;
            var data = createWeekStats(chartData, weeks);

            var html = '';
            data.forEach(function (day) {
                html += '<tr><td>' + day.name + '</td><td>' + day.stat + '</td></tr>';
            });
            tBody.innerHTML = html;
        }

        select.addEventListener('change', update);

        update();
    }
}

function YearlyStats() {
    // we started on 23.02.2012
    var DAYS_FIRST_YEAR = 365 - 54;
    var START_YEAR = 2012;
    var YEAR_TODAY = new Date().getFullYear();

    /**
     * Creates the stats per the given year
     * @param openTimeData
     * @param year
     * @param daysToUse
     * @returns {{year: *, hoursTotal: number, hoursPerDay: number}}
     */
    function createStats(openTimeData, year, daysToUse) {
        var yearBegin = new Date(year, 0).getTime();
        var yearEnd = new Date(year + 1, 0).getTime();

        var yearSum = 0;
        var lastTs = 0;
        for (var i = 0; i < openTimeData.length; i++) {
            if (openTimeData[i].date < yearBegin) {
                continue;
            }
            if (openTimeData[i].date >= yearEnd) {
                break;
            }
            if (openTimeData[i].date === lastTs) {
                continue;
            }
            // console.log(new Date(data[i].date));
            lastTs = openTimeData[i].date;
            yearSum += openTimeData[i].durationInSec;
        }

        var hoursTotal = yearSum / 60 / 60;
        return {
            year: year,
            hoursTotal: Math.round(hoursTotal),
            hoursPerDay: parseFloat(hoursTotal / daysToUse).toFixed(2)
        };
    }

    // gets the amount of days in this year until now
    function daysUntilNow() {
        var begin = new Date(YEAR_TODAY, 0).getTime();
        var end = Date.now();

        return Math.floor((end - begin) / MILLIS_PER_DAY);
    }

    this.start = function (tBodyId, chartData) {
        var yearStats = [];
        for (var y = YEAR_TODAY; y >= START_YEAR; y--) {
            var daysToUse = 365;
            if (y === START_YEAR) {
                daysToUse = DAYS_FIRST_YEAR;
            } else if (y === YEAR_TODAY) {
                daysToUse = daysUntilNow();
            }
            yearStats.push(createStats(chartData, y, daysToUse));
        }

        var html = '';
        console.log('x', yearStats)
        yearStats.forEach(function (stat) {
            html += '<tr><td>' + stat.year + '</td><td>' + stat.hoursTotal + '</td><td>' + stat.hoursPerDay + '</td></tr>';
        });

        document.getElementById(tBodyId).innerHTML = html;
    };
}