var CHECK_INTERVAL = 5 * 60 * 1000;
var START_FAIL_AFTER = 3 * 1000;

var lastkeepalive, source, startUpTimer;

var timestamps = {
    spaceOpen: 0,
    spaceDevices: 0,
    lab3dOpen: 0,
    machining: 0,
    woodworking: 0,
    // freifunk: 0,
    // weather: 0,
    energyFront: 0,
    energyBack: 0,
    energyMachining: 0
};

function statusInit() {
    initEventSource();
    updateLastUpdates();
}

function initEventSource() {
    clearTimeout(startUpTimer);

    startUpTimer = setTimeout(function () {
        addBodyClass('startup-error');
    }, START_FAIL_AFTER);

    source = new EventSource('/api/statusStream?spaceOpen=1&radstelleOpen=1&machining=1&woodworking=1&spaceDevices=1&powerUsage=1&lab3dOpen=1&mqtt=1&keyholder=1&keyholder_machining=1&keyholder_woodworking=1');
    source.onopen = function () {
        console.log('EventSource is open');
        removeBodyClass('connection-error')
        lastkeepalive = Date.now();
    };

    source.onerror = function (err) {
        console.error('EventSource error.', err);
        clearTimeout(startUpTimer);
        addBodyClass('connection-error');
        source.close();
        setTimeout(initEventSource, 5000);
    };

    source.addEventListener('mqtt', function (e) {
        var data = JSON.parse(e.data);

        // we got real data...
        clearTimeout(startUpTimer);
        removeBodyClass('startup-error');

        if (data.spaceBrokerOnline) {
            setText('spaceBrokerOnline_status', 'Läuft');
            setOnlyClass('spaceBrokerOnline_style', '');
        } else {
            setText('spaceBrokerOnline_status', 'Offline!');
            setOnlyClass('spaceBrokerOnline_style', 'danger');
        }
    });

    addKeyHolderListener(source, 'keyholder');
    addKeyHolderListener(source, 'keyholder_machining');
    addKeyHolderListener(source, 'keyholder_woodworking');

    addOpenListener(source, 'spaceOpen');
    addOpenListener(source, 'lab3dOpen');
    addOpenListener(source, 'radstelleOpen');
    addOpenListener(source, 'machining');
    addOpenListener(source, 'woodworking');


    source.addEventListener('spaceDevices', function (e) {
        var data = JSON.parse(e.data);

        timestamps.spaceDevices = data.timestamp;
        var personList = 'Keiner sichtbar';
        if (data.people && data.people.length > 0) {
            personList = data.people.map(makePersonHtml).join('');
        }
        setText('personList', personList);
        setText('anonPeopleCount', data.peopleCount - data.people.length);
        setText('devicesCount', data.unknownDevicesCount);
    });

    source.addEventListener('powerUsage', function (e) {
        var data = JSON.parse(e.data);

        timestamps.energyFront = data.front.timestamp;
        timestamps.energyBack = data.back.timestamp;
        timestamps.energyMachining = data.machining.timestamp;
        setText('energyFront', data.front.value);
        setText('energyBack', data.back.value)
        setText('energyMachining', data.machining.value)
    });


    source.addEventListener('keepalive', function (e) {
        lastkeepalive = Date.now();
    }, false);

    setTimeout(checkConnection, CHECK_INTERVAL);
}

// check whether we have seen a keepalive event within the last 20 minutes or are disconnected; reconnect if necessary
function checkConnection() {
    console.log('Checking connection...');
    if ((Date.now() - lastkeepalive > 20 * 60 * 1000) || source.readyState === 2) {
        source.close();
        console.warn('Restarting event source.', Date.now() - lastkeepalive, source.readyState);
        setTimeout(initEventSource, 3000);
        return;
    }
    setTimeout(checkConnection, CHECK_INTERVAL);
}

function addKeyHolderListener(source, topic) {
    source.addEventListener(topic, function (e) {
        var keyholder = e.data;
        setText(topic + '_name', keyholder);
    });
}

function addOpenListener(source, topic) {
    source.addEventListener(topic, function (e) {

        var data = JSON.parse(e.data);
        timestamps[topic] = data.timestamp;
        var status = '?';
        var style = '';
        switch (data.state) {
            case 'none':
                status = 'ZU!';
                style = 'danger';
                break;
            case 'open':
                status = 'AUF!';
                style = 'success';
                break;
            case 'open+':
                status = 'AUF+!';
                style = 'success';
                break;
            case 'keyholder':
                status = 'ZU (Keyholder only!)';
                style = 'danger';
                break;
            case 'member':
                status = 'ZU (Member only!)';
                style = 'danger';
                break;
            case 'closing':
                status = 'GLEICH ZU!';
                style = 'warning';
                break;
        }
        setText(topic + '_status', status);
        setOnlyClass(topic + '_style', style);
    }, false);
}

function makePersonHtml(personData) {
    let safePersonName = personData.name.replace(/[&<>"']/g, c => `&#${c.charCodeAt(0)};`)
    var html = '<span class="person"><span class="name">' + safePersonName + '</span>';
    if (personData.devices && personData.devices.length > 0) {
        html += '<span class="devices text-muted"> [';
        personData.devices.forEach(function (device) {
            if (device.name !== '') {
                let safeName = device.name.replace(/[&<>"']/g, c => `&#${c.charCodeAt(0)};`)
                html += '<small class="location ' + device.location + '">' + safeName + '</small>';
            }
        });
        html += ']</span>';
    }
    html += '</span>';

    return html;
}

function elapsedTime(t) {
    var result = '', diff;
    diff = Math.round(Date.now() / 1000) - t;
    if (diff / 86400 >= 1) {
        result += Math.floor(diff / 86400) + 'd';
    }
    diff %= 86400;
    if (diff / 3600 >= 1) {
        result += Math.floor(diff / 3600) + 'h';
    }
    diff %= 3600;
    if (diff / 60 >= 1) {
        result += Math.floor(diff / 60) + 'm';
    }
    diff %= 60;
    result += Math.floor(diff) + 's';
    return result;
}

function updateLastUpdates() {
    for (var field in timestamps) {
        if (timestamps[field] > 0) {
            setText(field + '_lu', elapsedTime(timestamps[field]));
        }
    }

    setTimeout(updateLastUpdates, 1000);
}

function setText(domId, newText) {
    document.getElementById(domId).innerHTML = newText;
}

function setOnlyClass(domId, className) {
    document.getElementById(domId).className = className;
}

function addBodyClass(className) {
    document.getElementsByTagName('body')[0].classList.add(className);
}

function removeBodyClass(className) {
    document.getElementsByTagName('body')[0].classList.remove(className);
}
