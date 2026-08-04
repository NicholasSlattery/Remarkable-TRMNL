pragma ComponentBehavior: Bound
import QtQuick 2.5
import QtQuick.Controls 2.5
import net.asivery.AppLoad 1.0

Rectangle {
    id: root
    anchors.fill: parent
    color: "#e8e6df"
    signal close

    property var appState: ({})
    property var appConfig: ({
        base_url: "https://trmnl.com", fit_mode: "fit", orientation: "auto",
        minimum_refresh_seconds: 60, restore_brightness_on_exit: true,
        use_system_brightness: true, brightness_percent: 50,
        start_with_cache_offline: true, logging_level: "info"
    })
    property string dashboardSource: ""
    property string statusText: "Starting…"
    property bool offline: false
    property bool controlsVisible: false
    property bool settingsVisible: false
    property bool apiKeyConfigured: false
    property bool initialized: false

    function unloading() { endpoint.terminate() }
    function fitValue() {
        if (appConfig.fit_mode === "fill") return Image.PreserveAspectCrop
        if (appConfig.fit_mode === "stretch") return Image.Stretch
        return Image.PreserveAspectFit
    }
    function configFromFields() {
        return {
            api_key: apiKey.text,
            base_url: serverMode.currentIndex === 0 ? "https://trmnl.com" : customURL.text,
            device_id: deviceID.text,
            refresh_mode: "server",
            minimum_refresh_seconds: parseInt(minRefresh.text || "60"),
            fit_mode: fitMode.currentText.toLowerCase(),
            orientation: orientation.currentText.toLowerCase(),
            invert: invertMode.checked,
            use_system_brightness: useSystemBrightness.checked,
            restore_brightness_on_exit: restoreBrightness.checked,
            brightness_percent: Math.round(brightness.value),
            start_with_cache_offline: startCached.checked,
            always_on: false,
            logging_level: logLevel.currentText.toLowerCase(),
            history_limit: 30
        }
    }
    function applyState(s) {
        appState = s
        appConfig = s.config || appConfig
        apiKeyConfigured = !!s.api_key_configured
        if (s.brightness_percent !== undefined) brightness.value = s.brightness_percent
        if (!apiKeyConfigured && !appConfig.device_id) settingsVisible = true
    }
    function configureFields() {
        var c = appConfig
        serverMode.currentIndex = (c.base_url === "https://trmnl.com" || c.base_url === "https://trmnl.app") ? 0 : 1
        customURL.text = serverMode.currentIndex === 1 ? c.base_url : ""
        deviceID.text = c.device_id || ""
        apiKey.text = ""
        minRefresh.text = String(c.minimum_refresh_seconds || 60)
        fitMode.currentIndex = c.fit_mode === "fill" ? 1 : (c.fit_mode === "stretch" ? 2 : 0)
        orientation.currentIndex = c.orientation === "portrait" ? 1 : (c.orientation === "landscape" ? 2 : 0)
        invertMode.checked = !!c.invert
        useSystemBrightness.checked = !!c.use_system_brightness
        restoreBrightness.checked = c.restore_brightness_on_exit !== false
        startCached.checked = c.start_with_cache_offline !== false
        logLevel.currentIndex = c.logging_level === "debug" ? 1 : (c.logging_level === "warning" ? 2 : 0)
    }

    AppLoad {
        id: endpoint
        applicationID: "trmnl.remarkable"
        onMessageReceived: function(type, contents) {
            var data = {}
            try { data = JSON.parse(contents || "{}") } catch (e) { root.statusText = "Invalid backend message"; return }
            if (type === 101) { root.applyState(data); if (!root.initialized) { root.configureFields(); root.initialized = true } }
            else if (type === 102) {
                root.dashboardSource = ""
                root.dashboardSource = data.path
                root.offline = !!data.cached
            } else if (type === 103) { root.statusText = data.message || ""; root.offline = false }
            else if (type === 104) { root.statusText = data.message || "Error"; root.offline = true }
            else if (type === 105) {
                historyModel.clear()
                if (Array.isArray(data)) {
                    for (var i=0; i<data.length; ++i) historyModel.append(data[i])
                }
            }
            else if (type === 106) { testResult.text = data.message || ""; testResult.color = data.ok ? "#124e2c" : "#7a1515" }
            else if (type === 107) { diagnosticsText.text = JSON.stringify(data, null, 2) }
        }
    }

    Component.onCompleted: endpoint.sendMessage(1, "")
    Connections {
        target: Qt.application
        function onStateChanged() {
            // qmllint disable missing-property
            if (Qt.application.state === Qt.ApplicationActive) endpoint.sendMessage(10, "")
            // qmllint enable missing-property
        }
    }

    Item {
        id: displayArea
        anchors.centerIn: parent
        width: root.appConfig.orientation === "landscape" ? root.height : root.width
        height: root.appConfig.orientation === "landscape" ? root.width : root.height
        rotation: root.appConfig.orientation === "landscape" ? 90 : 0

        Image {
            id: dashboard
            anchors.fill: parent
            source: root.dashboardSource
            fillMode: root.fitValue()
            smooth: true
            mipmap: true
            cache: false
            asynchronous: true
        }
    }

    Rectangle {
        visible: root.dashboardSource === ""
        anchors.centerIn: parent
        width: Math.min(parent.width * 0.8, 900)
        height: 320
        color: "#f5f3ec"
        border.width: 3
        radius: 12
        Column {
            anchors.centerIn: parent
            spacing: 28
            Text { anchors.horizontalCenter: parent.horizontalCenter; text: "TRMNL"; font.pixelSize: 80; font.bold: true; color: "#111" }
            Text { anchors.horizontalCenter: parent.horizontalCenter; text: root.apiKeyConfigured ? "Waiting for dashboard…" : "Tap the upper-right corner to set up"; font.pixelSize: 28; color: "#333" }
        }
    }

    Rectangle {
        anchors.right: parent.right; anchors.top: parent.top; width: 112; height: 112
        color: root.controlsVisible ? "#dddddd" : "transparent"; opacity: root.controlsVisible ? 1 : 0.15
        Text { anchors.centerIn: parent; text: "Menu"; font.pixelSize: 18; font.bold: true; color: "#111" }
        MouseArea { anchors.fill: parent; onClicked: { root.controlsVisible = !root.controlsVisible; if (root.controlsVisible) root.settingsVisible = false } }
    }

    Item {
        anchors.left: parent.left; anchors.top: parent.top; width: 120; height: 120
        MouseArea {
            anchors.fill: parent
            onPressed: fallbackExit.start()
            onReleased: fallbackExit.stop()
            onCanceled: fallbackExit.stop()
        }
        Timer { id: fallbackExit; interval: 2000; onTriggered: root.close() }
    }

    Rectangle {
        id: statusBadge
        visible: root.statusText !== "" && !root.controlsVisible && !root.settingsVisible
        anchors.left: parent.left; anchors.bottom: parent.bottom; anchors.margins: 18
        width: Math.min(statusLabel.implicitWidth + 34, parent.width * 0.72); height: 54
        color: root.offline ? "#f1d8d0" : "#eceae3"; border.width: 2; radius: 8; opacity: 0.93
        Text { id: statusLabel; anchors.centerIn: parent; text: root.statusText; elide: Text.ElideRight; width: parent.width - 24; font.pixelSize: 20; color: "#222" }
    }

    Rectangle {
        id: controls
        visible: root.controlsVisible
        anchors.right: parent.right; anchors.top: parent.top; anchors.bottom: parent.bottom
        width: Math.min(parent.width * 0.78, 840)
        color: "#f4f2eb"; border.width: 3

        Flickable {
            anchors.fill: parent; anchors.margins: 28; contentHeight: controlColumn.implicitHeight; clip: true
            Column {
                id: controlColumn; width: parent.width; spacing: 22
                Row {
                    width: parent.width
                    Text { text: "TRMNL controls"; font.pixelSize: 38; font.bold: true; width: parent.width - 100 }
                    Button { text: "×"; width: 84; height: 70; font.pixelSize: 34; onClicked: root.controlsVisible = false }
                }
                Text { text: root.statusText; width: parent.width; wrapMode: Text.Wrap; font.pixelSize: 23; color: root.offline ? "#7a1515" : "#222" }
                Row {
                    width: parent.width; spacing: 30
                    Text { text: "Battery  " + (root.appState.battery_percent >= 0 ? root.appState.battery_percent + "%" : "unavailable"); font.pixelSize: 23; font.bold: true }
                    Text { text: "Last refresh  " + (root.appState.last_refresh ? new Date(root.appState.last_refresh).toLocaleString(Qt.locale(), Locale.ShortFormat) : "never"); font.pixelSize: 23; font.bold: true }
                }
                Text { text: "Brightness  " + Math.round(brightness.value) + "%"; font.pixelSize: 28; font.bold: true }
                Slider {
                    id: brightness; width: parent.width; height: 78; from: 0; to: 100; stepSize: 1
                    enabled: root.appState.brightness !== undefined && !useSystemBrightness.checked
                    onMoved: brightnessDebounce.restart()
                }
                Timer { id: brightnessDebounce; interval: 250; onTriggered: endpoint.sendMessage(6, JSON.stringify({percent:Math.round(brightness.value)})) }
                CheckBox { id: useSystemBrightness; text: "Use system brightness"; font.pixelSize: 24; onClicked: endpoint.sendMessage(13, JSON.stringify({use_system_brightness:checked})) }
                Row {
                    spacing: 14
                    Button { text: "Refresh now"; width: 220; height: 78; font.pixelSize: 23; onClicked: endpoint.sendMessage(4, "") }
                    Button { text: "Next screen"; width: 220; height: 78; font.pixelSize: 23; onClicked: endpoint.sendMessage(5, "") }
                    Button { text: "Previous"; width: 180; height: 78; font.pixelSize: 23; onClicked: endpoint.sendMessage(12, "") }
                }
                CheckBox { text: "Dark / invert image"; checked: !!root.appConfig.invert; font.pixelSize: 24; onClicked: endpoint.sendMessage(13, JSON.stringify({invert:checked})) }
                Row {
                    spacing: 14
                    Button { text: "Settings"; width: 220; height: 78; font.pixelSize: 23; onClicked: { root.configureFields(); root.settingsVisible = true; root.controlsVisible = false } }
                    Button { text: "Diagnostics"; width: 220; height: 78; font.pixelSize: 23; onClicked: { endpoint.sendMessage(8, ""); diagnosticsPopup.open() } }
                }
                Button { text: "Return to reMarkable"; width: parent.width; height: 92; font.pixelSize: 27; font.bold: true; onClicked: root.close() }
                Text { text: "Refresh history"; font.pixelSize: 28; font.bold: true }
                Repeater {
                    model: historyModel
                    Rectangle {
                        id: historyEntry
                        required property int index
                        required property var model
                        width: controlColumn.width; height: 66; color: historyEntry.index % 2 ? "#e5e2da" : "#efede6"
                        Text { anchors.fill: parent; anchors.margins: 10; verticalAlignment: Text.AlignVCenter; elide: Text.ElideRight; font.pixelSize: 18; text: (historyEntry.model.ok ? "OK  " : "FAIL  ") + new Date(historyEntry.model.at).toLocaleString(Qt.locale(), Locale.ShortFormat) + " | " + historyEntry.model.action + " | " + historyEntry.model.detail }
                    }
                }
                Text { text: "Fallback exit: hold the upper-left corner for 2 seconds."; font.pixelSize: 20; color: "#444"; wrapMode: Text.Wrap; width: parent.width }
            }
        }
    }

    Rectangle {
        id: settings
        visible: root.settingsVisible
        anchors.fill: parent; color: "#f4f2eb"; border.width: 3
        Flickable {
            anchors.fill: parent; anchors.margins: 34; contentHeight: settingsColumn.implicitHeight + 40; clip: true
            Column {
                id: settingsColumn; width: parent.width; spacing: 18
                Row { width: parent.width; Text { text: "TRMNL setup & settings"; font.pixelSize: 38; font.bold: true; width: parent.width - 120 } Button { text: "×"; width: 84; height: 70; font.pixelSize: 34; onClicked: root.settingsVisible = false } }
                Text { text: root.apiKeyConfigured ? "API key configured. Leave blank to keep it, or enter a replacement." : "Enter your TRMNL Device API key."; font.pixelSize: 21; width: parent.width; wrapMode: Text.Wrap }
                TextField { id: apiKey; width: parent.width; height: 68; echoMode: TextInput.Password; placeholderText: "TRMNL Device API key"; font.pixelSize: 23 }
                ComboBox { id: serverMode; width: parent.width; height: 68; model: ["TRMNL cloud", "Custom BYOS server"]; font.pixelSize: 23 }
                TextField { id: customURL; visible: serverMode.currentIndex === 1; width: parent.width; height: 68; placeholderText: "https://your-server.example"; font.pixelSize: 23 }
                TextField { id: deviceID; width: parent.width; height: 68; placeholderText: "Optional device ID / MAC address"; font.pixelSize: 23 }
                Row { spacing: 18; width: parent.width; Text { text: "Minimum refresh seconds"; width: 330; font.pixelSize: 22; anchors.verticalCenter: parent.verticalCenter } TextField { id: minRefresh; width: 200; height: 64; inputMethodHints: Qt.ImhDigitsOnly; font.pixelSize: 22 } }
                Row { spacing: 18; Text { text: "Image fit"; width: 180; font.pixelSize: 22; anchors.verticalCenter: parent.verticalCenter } ComboBox { id: fitMode; width: 240; height: 64; model: ["Fit", "Fill", "Stretch"]; font.pixelSize: 22 } Text { text: "Orientation"; width: 180; font.pixelSize: 22; anchors.verticalCenter: parent.verticalCenter } ComboBox { id: orientation; width: 240; height: 64; model: ["Auto", "Portrait", "Landscape"]; font.pixelSize: 22 } }
                CheckBox { id: invertMode; text: "Dark / invert image"; font.pixelSize: 22 }
                CheckBox { id: restoreBrightness; text: "Restore previous brightness when exiting"; font.pixelSize: 22 }
                CheckBox { id: startCached; text: "Start with cached screen when offline"; font.pixelSize: 22 }
                Row { spacing: 18; Text { text: "Logging"; width: 180; font.pixelSize: 22; anchors.verticalCenter: parent.verticalCenter } ComboBox { id: logLevel; width: 240; height: 64; model: ["Info", "Debug", "Warning"]; font.pixelSize: 22 } }
                Text { id: testResult; width: parent.width; font.pixelSize: 21; wrapMode: Text.Wrap }
                Row { spacing: 14; Button { text: "Test connection"; width: 240; height: 78; font.pixelSize: 22; onClicked: { testResult.text = "Testing…"; endpoint.sendMessage(3, JSON.stringify(root.configFromFields())) } } Button { text: "Save"; width: 190; height: 78; font.pixelSize: 22; onClicked: { endpoint.sendMessage(2, JSON.stringify(root.configFromFields())); root.settingsVisible = false } } Button { text: "Clear cache"; width: 190; height: 78; font.pixelSize: 22; onClicked: endpoint.sendMessage(7, "") } Button { text: "Reset"; width: 150; height: 78; font.pixelSize: 22; onClicked: endpoint.sendMessage(11, "") } }
                Text { text: "Uninstall: /home/root/trmnl-remarkable/uninstall.sh\nRecovery: /home/root/trmnl-remarkable/recover-stock.sh\nRecovery tools: /home/root/trmnl-remarkable"; font.pixelSize: 19; width: parent.width; wrapMode: Text.Wrap; color: "#444" }
            }
        }
    }

    Popup {
        id: diagnosticsPopup; modal: true; focus: true; x: root.width*0.08; y: root.height*0.08; width: root.width*0.84; height: root.height*0.84
        background: Rectangle { color: "#f4f2eb"; border.width: 3 }
        contentItem: Column { spacing: 12; Text { text: "Diagnostics"; font.pixelSize: 30; font.bold: true } ScrollView { width: parent.width; height: parent.height - 100; TextArea { id: diagnosticsText; readOnly: true; wrapMode: Text.Wrap; font.pixelSize: 17 } } Button { text: "Close"; width: 180; height: 64; onClicked: diagnosticsPopup.close() } }
    }

    ListModel { id: historyModel }
}
