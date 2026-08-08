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
        brightness_schedule_enabled: false, brightness_schedule: [],
        start_with_cache_offline: true, wake_for_refresh: true, logging_level: "info"
    })
    property string dashboardSource: ""
    property bool dashboardAvailable: false
    property var batteryTest: ({active:false, samples:[], estimate_available:false})
    property string statusText: "Starting…"
    property bool offline: false
    property bool controlsVisible: false
    property bool settingsVisible: false
    property bool brightnessScheduleVisible: false
    property bool apiKeyConfigured: false
    property bool initialized: false
    property int cleaningPhase: 0
    property var scheduleSlots: []
    property var scheduleDayNames: ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"]
    property int scheduleStartDay: -1
    property int scheduleEndDay: -1
    property int scheduleStartSlot: -1
    property int scheduleEndSlot: -1
    property bool scheduleDirty: false

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
            use_system_brightness: settingsUseSystemBrightness.checked,
            restore_brightness_on_exit: restoreBrightness.checked,
            brightness_percent: Math.round(settingsBrightness.value),
            brightness_schedule_enabled: !!appConfig.brightness_schedule_enabled,
            brightness_schedule: appConfig.brightness_schedule || [],
            start_with_cache_offline: startCached.checked,
            always_on: false,
            wake_for_refresh: wakeForRefresh.checked,
            logging_level: logLevel.currentText.toLowerCase(),
            // Preserve settings that have no on-device control rather than
            // resetting them every time this page is saved.
            history_limit: appConfig.history_limit || 30,
            dither_palette: appConfig.dither_palette,
            quiet_hours_enabled: quietHours.checked,
            quiet_hours_start: quietStart.text,
            quiet_hours_end: quietEnd.text,
            battery_saver_percent: batterySaver.checked ? 20 : 0,
            dither: ditherMode.checked ? "auto" : "off",
            update_check: updateCheck.checked
        }
    }
    function applyState(s) {
        appState = s
        appConfig = s.config || appConfig
        apiKeyConfigured = !!s.api_key_configured
        if (s.brightness_percent !== undefined) {
            brightness.value = s.brightness_percent
            settingsBrightness.value = appConfig.brightness_percent === undefined ? s.brightness_percent : appConfig.brightness_percent
        }
        useSystemBrightness.checked = !!appConfig.use_system_brightness
        settingsUseSystemBrightness.checked = !!appConfig.use_system_brightness
        if (!apiKeyConfigured && !appConfig.device_id) settingsVisible = true
    }
    function configureFields() {
        var c = appConfig
        serverMode.currentIndex = (c.base_url === "https://trmnl.com" || c.base_url === "https://trmnl.app") ? 0 : 1
        customURL.text = serverMode.currentIndex === 1 ? c.base_url : ""
        deviceID.text = c.device_id || appState.detected_device_id || ""
        apiKey.text = ""
        minRefresh.text = String(c.minimum_refresh_seconds || 60)
        fitMode.currentIndex = c.fit_mode === "fill" ? 1 : (c.fit_mode === "stretch" ? 2 : 0)
        orientation.currentIndex = c.orientation === "portrait" ? 1 : (c.orientation === "landscape" ? 2 : 0)
        invertMode.checked = !!c.invert
        useSystemBrightness.checked = !!c.use_system_brightness
        settingsUseSystemBrightness.checked = !!c.use_system_brightness
        settingsBrightness.value = c.brightness_percent === undefined ? 50 : c.brightness_percent
        brightness.value = settingsBrightness.value
        restoreBrightness.checked = c.restore_brightness_on_exit !== false
        startCached.checked = c.start_with_cache_offline !== false
        wakeForRefresh.checked = c.wake_for_refresh !== false
        logLevel.currentIndex = c.logging_level === "debug" ? 1 : (c.logging_level === "warning" ? 2 : 0)
        quietHours.checked = !!c.quiet_hours_enabled
        quietStart.text = c.quiet_hours_start || "23:00"
        quietEnd.text = c.quiet_hours_end || "07:00"
        batterySaver.checked = (c.battery_saver_percent || 0) > 0
        ditherMode.checked = c.dither === "auto"
        updateCheck.checked = !!c.update_check
    }
    function requestNativeFullRefresh() {
        // AppLoad owns the framebuffer controller above this loaded component.
        // Its signal is the same full-panel refresh used by the five-finger
        // cleanup gesture. Locate it without depending on private object IDs.
        var cursor = root.parent
        for (var depth = 0; cursor && depth < 8; ++depth) {
            var children = cursor.children || []
            for (var i = 0; i < children.length; ++i) {
                var candidate = children[i]
                if (candidate && candidate.requestFullRefresh !== undefined
                        && candidate.refreshMode !== undefined) {
                    candidate.requestFullRefresh()
                    return true
                }
            }
            cursor = cursor.parent
        }
        return false
    }
    function cleanScreen() {
        panelRefreshDelay.restart()
    }
    function fallbackCleanScreen() {
        cleaningPhase = 1
        cleaningFlashTimer.restart()
    }
    function formatTestHours(hours) {
        if (hours === undefined || hours === null || !isFinite(hours)) return "calculating…"
        if (hours < 1) return Math.max(0, Math.round(hours * 60)) + " minutes"
        if (hours < 48) return hours.toFixed(1) + " hours"
        return (hours / 24).toFixed(1) + " days"
    }
    function loadBrightnessSchedule() {
        var source = appConfig.brightness_schedule || []
        var slots = []
        for (var i = 0; i < 336; ++i) {
            var value = i < source.length ? Number(source[i]) : -1
            slots.push(value >= 0 && value <= 100 ? Math.round(value) : -1)
        }
        scheduleSlots = slots
        scheduleEnabled.checked = !!appConfig.brightness_schedule_enabled
        scheduleDirty = false
        scheduleStartDay = scheduleEndDay = scheduleStartSlot = scheduleEndSlot = -1
    }
    function showBrightnessSchedule(show) {
        if (show) loadBrightnessSchedule()
        brightnessScheduleVisible = show
        controlsVisible = false
        settingsVisible = false
        cleanScreen()
    }
    function scheduleTime(slot) {
        if (slot >= 48) return "24:00"
        var hour = Math.floor(slot / 2)
        return (hour < 10 ? "0" : "") + hour + (slot % 2 ? ":30" : ":00")
    }
    function selectedFirstDay() { return Math.min(scheduleStartDay, scheduleEndDay) }
    function selectedLastDay() { return Math.max(scheduleStartDay, scheduleEndDay) }
    function selectedFirstSlot() { return Math.min(scheduleStartSlot, scheduleEndSlot) }
    function selectedLastSlot() { return Math.max(scheduleStartSlot, scheduleEndSlot) }
    function scheduleCellSelected(day, slot) {
        if (scheduleStartDay < 0 || scheduleStartSlot < 0) return false
        return day >= selectedFirstDay() && day <= selectedLastDay()
                && slot >= selectedFirstSlot() && slot <= selectedLastSlot()
    }
    function scheduleDayRange() {
        if (scheduleStartDay < 0) return "No days selected"
        var first = selectedFirstDay(), last = selectedLastDay()
        return first === last ? scheduleDayNames[first] : scheduleDayNames[first] + "–" + scheduleDayNames[last]
    }
    function scheduleSelectionSummary() {
        if (scheduleStartDay < 0 || scheduleStartSlot < 0) return "Drag across the grid to select days and times"
        return scheduleDayRange() + "  ·  " + scheduleTime(selectedFirstSlot()) + "–"
                + scheduleTime(selectedLastSlot() + 1) + "  ·  "
                + (scheduleSelectionUsesDefault() ? "using default brightness" : "setting " + Math.round(scheduleBrightness.value) + "%")
    }
    function scheduleSelectionUsesDefault() {
        if (scheduleStartDay < 0 || scheduleStartSlot < 0) return false
        for (var slot = selectedFirstSlot(); slot <= selectedLastSlot(); ++slot) {
            for (var day = selectedFirstDay(); day <= selectedLastDay(); ++day) {
                if (scheduleSlots[slot * 7 + day] >= 0) return false
            }
        }
        return true
    }
    function scheduleSlotColor(percent) {
        if (percent === undefined || percent < 0) return "#f7f5ef"
        if (percent <= 20) return "#3b3b3b"
        if (percent <= 40) return "#686868"
        if (percent <= 60) return "#929292"
        if (percent <= 80) return "#bcbcbc"
        return "#dedede"
    }
    function beginScheduleSelection(x, y) {
        var day = Math.max(0, Math.min(6, Math.floor(x / (scheduleGrid.width / 7))))
        var slot = Math.max(0, Math.min(47, Math.floor(y / (scheduleGrid.height / 48))))
        scheduleStartDay = scheduleEndDay = day
        scheduleStartSlot = scheduleEndSlot = slot
        var existing = scheduleSlots[slot * 7 + day]
        scheduleBrightness.value = existing >= 0 ? existing : (appConfig.brightness_percent === undefined ? 50 : appConfig.brightness_percent)
    }
    function extendScheduleSelection(x, y) {
        scheduleEndDay = Math.max(0, Math.min(6, Math.floor(x / (scheduleGrid.width / 7))))
        scheduleEndSlot = Math.max(0, Math.min(47, Math.floor(y / (scheduleGrid.height / 48))))
        scheduleRefreshDelay.restart()
    }
    function applyBrightnessToSelection(percent) {
        if (scheduleStartDay < 0 || scheduleStartSlot < 0) return
        var slots = scheduleSlots.slice(0)
        for (var slot = selectedFirstSlot(); slot <= selectedLastSlot(); ++slot) {
            for (var day = selectedFirstDay(); day <= selectedLastDay(); ++day) {
                slots[slot * 7 + day] = Math.max(0, Math.min(100, Math.round(percent)))
            }
        }
        scheduleSlots = slots
        scheduleDirty = true
        scheduleRefreshDelay.restart()
    }
    function clearBrightnessSelection() {
        if (scheduleStartDay < 0 || scheduleStartSlot < 0) return
        var slots = scheduleSlots.slice(0)
        for (var slot = selectedFirstSlot(); slot <= selectedLastSlot(); ++slot) {
            for (var day = selectedFirstDay(); day <= selectedLastDay(); ++day) slots[slot * 7 + day] = -1
        }
        scheduleSlots = slots
        scheduleDirty = true
        scheduleRefreshDelay.restart()
    }
    function showControls(show) {
        controlsVisible = show
        if (show) { settingsVisible = false; brightnessScheduleVisible = false }
        cleanScreen()
    }
    function showSettings(show) {
        if (show) configureFields()
        settingsVisible = show
        controlsVisible = false
        brightnessScheduleVisible = false
        cleanScreen()
    }
    Timer { id: scheduleRefreshDelay; interval: 320; onTriggered: root.cleanScreen() }

    Timer {
        id: panelRefreshDelay
        interval: 140
        onTriggered: {
            if (!root.requestNativeFullRefresh()) root.fallbackCleanScreen()
        }
    }
    Timer {
        id: cleaningFlashTimer
        interval: 220
        repeat: true
        onTriggered: {
            if (root.cleaningPhase === 1) {
                root.cleaningPhase = 2
            } else {
                root.cleaningPhase = 0
                stop()
            }
        }
    }

    AppLoad {
        id: endpoint
        applicationID: "trmnl.remarkable"
        onMessageReceived: function(type, contents) {
            var data = {}
            try { data = JSON.parse(contents || "{}") } catch (e) { root.statusText = "Invalid backend message"; return }
            if (type === 101) { root.applyState(data); if (!root.initialized) { root.configureFields(); root.initialized = true } }
            else if (type === 102) {
                root.dashboardSource = data.path
                root.dashboardAvailable = true
                panelRefreshDelay.restart()
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
            else if (type === 108) { root.batteryTest = data; batteryChart.requestPaint() }
        }
    }

    Component.onCompleted: endpoint.sendMessage(1, "")
    Connections {
        target: Qt.application
        function onStateChanged() {
            // qmllint disable missing-property
            if (Qt.application.state === Qt.ApplicationActive) {
                root.controlsVisible = false
                root.settingsVisible = false
                root.brightnessScheduleVisible = false
                diagnosticsPopup.close()
                root.cleanScreen()
                endpoint.sendMessage(10, "")
            }
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
        visible: !root.dashboardAvailable
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
        MouseArea { anchors.fill: parent; onClicked: root.showControls(!root.controlsVisible) }
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
                    Button { text: "×"; width: 84; height: 70; font.pixelSize: 34; onClicked: root.showControls(false) }
                }
                Text { text: root.statusText; width: parent.width; wrapMode: Text.Wrap; font.pixelSize: 23; color: root.offline ? "#7a1515" : "#222" }
                Row {
                    width: parent.width; spacing: 30
                    Text { text: "Battery  " + (root.appState.battery_percent >= 0 ? root.appState.battery_percent + "%" : "unavailable"); font.pixelSize: 23; font.bold: true }
                    Text { text: "Last refresh  " + (root.appState.last_refresh ? new Date(root.appState.last_refresh).toLocaleString(Qt.locale(), Locale.ShortFormat) : "never"); font.pixelSize: 23; font.bold: true }
                }
                Text {
                    visible: !!root.appState.battery_saving || !!root.appState.quiet_hours_active
                    width: parent.width; wrapMode: Text.Wrap; font.pixelSize: 21; color: "#7a4a12"
                    text: root.appState.battery_saving && root.appState.quiet_hours_active
                          ? "Battery saver and quiet hours are both active; scheduled refreshes resume in the morning."
                          : (root.appState.battery_saving
                             ? "Battery saver is active. Refreshes are less frequent until you charge the tablet."
                             : "Quiet hours are active. Scheduled refreshes resume when the window ends.")
                }
                Text { text: "Brightness  " + Math.round(brightness.value) + "%"; font.pixelSize: 28; font.bold: true }
                Slider {
                    id: brightness; width: parent.width; height: 78; from: 0; to: 100; stepSize: 1
                    enabled: root.appState.brightness !== undefined && !useSystemBrightness.checked
                    onMoved: { settingsBrightness.value = value; brightnessDebounce.restart() }
                }
                Timer { id: brightnessDebounce; interval: 250; onTriggered: endpoint.sendMessage(6, JSON.stringify({percent:Math.round(brightness.value)})) }
                CheckBox { id: useSystemBrightness; text: "Use system brightness"; font.pixelSize: 24; onClicked: { settingsUseSystemBrightness.checked = checked; endpoint.sendMessage(13, JSON.stringify({use_system_brightness:checked})) } }
                Button {
                    text: root.appState.brightness_schedule_active
                          ? "Brightness schedule  ·  now " + root.appState.scheduled_brightness_percent + "%"
                          : "Brightness schedule"
                    width: parent.width; height: 82; font.pixelSize: 25; font.bold: true
                    enabled: root.appState.brightness !== undefined
                    onClicked: root.showBrightnessSchedule(true)
                }
                Row {
                    spacing: 14
                    Button { text: "Refresh now"; width: 220; height: 78; font.pixelSize: 23; onClicked: endpoint.sendMessage(4, "") }
                    Button { text: "Next screen"; width: 220; height: 78; font.pixelSize: 23; onClicked: endpoint.sendMessage(5, "") }
                    Button { text: "Previous"; width: 180; height: 78; font.pixelSize: 23; onClicked: endpoint.sendMessage(12, "") }
                }
                CheckBox { text: "Dark / invert image"; checked: !!root.appConfig.invert; font.pixelSize: 24; onClicked: endpoint.sendMessage(13, JSON.stringify({invert:checked})) }
                Row {
                    spacing: 14
                    Button { text: "Settings"; width: 220; height: 78; font.pixelSize: 23; onClicked: root.showSettings(true) }
                    Button { text: "Diagnostics"; width: 220; height: 78; font.pixelSize: 23; onClicked: { root.cleanScreen(); endpoint.sendMessage(8, ""); diagnosticsPopup.open() } }
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
                Row { width: parent.width; Text { text: "TRMNL setup & settings"; font.pixelSize: 38; font.bold: true; width: parent.width - 120 } Button { text: "×"; width: 84; height: 70; font.pixelSize: 34; onClicked: root.showSettings(false) } }
                Text { text: root.apiKeyConfigured ? "API key configured. Leave blank to keep it, or enter a replacement." : "Enter your TRMNL Device API key."; font.pixelSize: 21; width: parent.width; wrapMode: Text.Wrap }
                TextField { id: apiKey; width: parent.width; height: 68; echoMode: TextInput.Password; placeholderText: "TRMNL Device API key"; font.pixelSize: 23 }
                ComboBox { id: serverMode; width: parent.width; height: 68; model: ["TRMNL cloud", "Custom BYOS server"]; font.pixelSize: 23 }
                TextField { id: customURL; visible: serverMode.currentIndex === 1; width: parent.width; height: 68; placeholderText: "https://your-server.example"; font.pixelSize: 23 }
                TextField { id: deviceID; width: parent.width; height: 68; placeholderText: "Optional device ID / MAC address"; font.pixelSize: 23 }
                Row { spacing: 18; width: parent.width; Text { text: "Minimum refresh seconds"; width: 330; font.pixelSize: 22; anchors.verticalCenter: parent.verticalCenter } TextField { id: minRefresh; width: 200; height: 64; inputMethodHints: Qt.ImhDigitsOnly; font.pixelSize: 22 } }
                Row { spacing: 18; Text { text: "Image fit"; width: 180; font.pixelSize: 22; anchors.verticalCenter: parent.verticalCenter } ComboBox { id: fitMode; width: 240; height: 64; model: ["Fit", "Fill", "Stretch"]; font.pixelSize: 22 } Text { text: "Orientation"; width: 180; font.pixelSize: 22; anchors.verticalCenter: parent.verticalCenter } ComboBox { id: orientation; width: 240; height: 64; model: ["Auto", "Portrait", "Landscape"]; font.pixelSize: 22 } }
                CheckBox { id: invertMode; text: "Dark / invert image"; font.pixelSize: 22 }
                Text { text: "Front light  " + Math.round(settingsBrightness.value) + "%"; font.pixelSize: 24; font.bold: true }
                Slider {
                    id: settingsBrightness; width: parent.width; height: 76; from: 0; to: 100; stepSize: 1
                    enabled: root.appState.brightness !== undefined && !settingsUseSystemBrightness.checked
                    onMoved: { brightness.value = value; settingsBrightnessDebounce.restart() }
                }
                Timer { id: settingsBrightnessDebounce; interval: 250; onTriggered: endpoint.sendMessage(6, JSON.stringify({percent:Math.round(settingsBrightness.value)})) }
                CheckBox {
                    id: settingsUseSystemBrightness; text: "Use the reMarkable system front-light level"; font.pixelSize: 22
                    onClicked: { useSystemBrightness.checked = checked; endpoint.sendMessage(13, JSON.stringify({use_system_brightness:checked})) }
                }
                CheckBox { id: ditherMode; text: "Smooth gradients for the colour panel"; font.pixelSize: 22 }
                Text {
                    width: parent.width; wrapMode: Text.Wrap; font.pixelSize: 19; color: "#444"
                    text: "Dashboards are drawn for bright screens, so gradients arrive as visible bands. This spreads the error across neighbouring pixels instead. Text and flat colour are left alone."
                }
                CheckBox { id: restoreBrightness; text: "Restore previous brightness when exiting"; font.pixelSize: 22 }
                CheckBox { id: startCached; text: "Start with cached screen when offline"; font.pixelSize: 22 }
                CheckBox { id: wakeForRefresh; text: "Sleep between updates and wake for refresh"; font.pixelSize: 22 }
                CheckBox { id: batterySaver; text: "Refresh less often below 20% battery"; font.pixelSize: 22 }
                CheckBox { id: quietHours; text: "Pause scheduled refreshes overnight"; font.pixelSize: 22 }
                Row {
                    spacing: 18; visible: quietHours.checked
                    Text { text: "From"; width: 80; font.pixelSize: 22; anchors.verticalCenter: parent.verticalCenter }
                    TextField { id: quietStart; width: 150; height: 64; placeholderText: "23:00"; font.pixelSize: 22 }
                    Text { text: "until"; width: 90; font.pixelSize: 22; anchors.verticalCenter: parent.verticalCenter }
                    TextField { id: quietEnd; width: 150; height: 64; placeholderText: "07:00"; font.pixelSize: 22 }
                }
                Text {
                    visible: quietHours.checked; width: parent.width; wrapMode: Text.Wrap; font.pixelSize: 19; color: "#444"
                    text: "24-hour times. The dashboard stays on screen; only scheduled wakeups pause. Refresh now still works."
                }
                Text {
                    width: parent.width; wrapMode: Text.Wrap; font.pixelSize: 19; color: root.appState.wake_alarm_error ? "#7a1515" : "#444"
                    text: !root.appState.wake_for_refresh_available
                          ? "Scheduled wake is unavailable on this firmware. TRMNL will still refresh whenever the tablet is awake."
                          : (root.appState.wake_alarm_error
                             ? "Scheduled wake error: " + root.appState.wake_alarm_error
                             : "Recommended: enable reMarkable Auto-sleep and Light sleep, but leave Auto power-off disabled. The dashboard remains visible while the tablet sleeps; TRMNL wakes at the next refresh interval, updates, and returns to sleep. A shutdown or empty battery still requires a normal power-on.")
                }
                Text {
                    visible: !!root.appState.next_refresh; width: parent.width; font.pixelSize: 19; color: "#444"
                    text: "Next refresh  " + (root.appState.next_refresh ? new Date(root.appState.next_refresh).toLocaleString(Qt.locale(), Locale.ShortFormat) : "after settings are saved")
                }
                Rectangle { width: parent.width; height: 3; color: "#b8b4aa" }
                Text { text: "Battery life test"; font.pixelSize: 28; font.bold: true }
                Text {
                    width: parent.width; wrapMode: Text.Wrap; font.pixelSize: 19; color: "#444"
                    text: "For the best estimate, charge to 100%, unplug the charger, tap Start new test, and leave TRMNL running until at least 10% is used. Plugging in during a test invalidates its projection. Samples survive sleep and app restarts."
                }
                Text {
                    width: parent.width; wrapMode: Text.Wrap; font.pixelSize: 22; font.bold: true
                    text: root.batteryTest.active
                          ? "Running · " + root.batteryTest.current_percent + "% battery"
                          : (root.batteryTest.started_at ? "Stopped · " + root.batteryTest.current_percent + "% battery" : "No test recorded")
                    color: root.batteryTest.active ? "#124e2c" : "#222"
                }
                Text {
                    visible: !!root.batteryTest.started_at; width: parent.width; wrapMode: Text.Wrap; font.pixelSize: 20
                    text: "Elapsed  " + root.formatTestHours(root.batteryTest.elapsed_hours)
                          + "    Used  " + root.batteryTest.percent_used + "%"
                          + "    Refreshes  " + root.batteryTest.refresh_count
                          + "    Wakes  " + root.batteryTest.wake_count
                }
                Text {
                    visible: !!root.batteryTest.started_at; width: parent.width; wrapMode: Text.Wrap; font.pixelSize: 22; font.bold: true
                    text: root.batteryTest.charging_observed
                          ? "Charging was detected during this test. Unplug, reset the data, and start a new test."
                          : root.batteryTest.estimate_available
                          ? "Projected total battery life  " + root.formatTestHours(root.batteryTest.projected_total_hours)
                            + "  ·  about " + root.formatTestHours(root.batteryTest.projected_remaining_hours) + " remaining"
                          : "Projection begins after the battery drops at least 10%."
                }
                Text {
                    visible: root.batteryTest.active && (root.batteryTest.battery_status === "Charging" || root.batteryTest.battery_status === "Full")
                    width: parent.width; wrapMode: Text.Wrap; font.pixelSize: 19; color: "#7a1515"
                    text: "Unplug the charger. Charging invalidates this test; reset it after disconnecting power."
                }
                Canvas {
                    id: batteryChart
                    visible: root.batteryTest.samples && root.batteryTest.samples.length > 1
                    width: parent.width; height: visible ? 230 : 0
                    onWidthChanged: requestPaint()
                    onPaint: {
                        var ctx = getContext("2d")
                        ctx.clearRect(0, 0, width, height)
                        ctx.fillStyle = "#eeece5"
                        ctx.fillRect(0, 0, width, height)
                        var samples = root.batteryTest.samples || []
                        if (samples.length < 2) return
                        var left = 54, right = width - 18, top = 18, bottom = height - 38
                        ctx.strokeStyle = "#77736b"; ctx.lineWidth = 2
                        ctx.beginPath(); ctx.moveTo(left, top); ctx.lineTo(left, bottom); ctx.lineTo(right, bottom); ctx.stroke()
                        var first = new Date(samples[0].at).getTime()
                        var last = new Date(samples[samples.length - 1].at).getTime()
                        if (last <= first) last = first + 1
                        ctx.strokeStyle = "#111111"; ctx.lineWidth = 4; ctx.beginPath()
                        for (var i = 0; i < samples.length; ++i) {
                            var x = left + (new Date(samples[i].at).getTime() - first) * (right - left) / (last - first)
                            var y = bottom - Math.max(0, Math.min(100, samples[i].percent)) * (bottom - top) / 100
                            if (i === 0) ctx.moveTo(x, y); else ctx.lineTo(x, y)
                        }
                        ctx.stroke()
                        ctx.fillStyle = "#222222"; ctx.font = "18px sans-serif"
                        ctx.fillText("100%", 2, top + 7); ctx.fillText("0%", 18, bottom + 7)
                        ctx.fillText("start", left, height - 10); ctx.fillText("now", right - 34, height - 10)
                    }
                }
                Row {
                    spacing: 14
                    Button { text: "Start new test"; width: 240; height: 76; font.pixelSize: 21; enabled: !root.batteryTest.active && root.appState.battery_percent >= 0; onClicked: endpoint.sendMessage(14, "") }
                    Button { text: "Stop"; width: 160; height: 76; font.pixelSize: 21; enabled: !!root.batteryTest.active; onClicked: endpoint.sendMessage(15, "") }
                    Button { text: "Reset data"; width: 190; height: 76; font.pixelSize: 21; enabled: !!root.batteryTest.started_at; onClicked: endpoint.sendMessage(16, "") }
                }
                Rectangle { width: parent.width; height: 3; color: "#b8b4aa" }
                CheckBox { id: updateCheck; text: "Check GitHub for a newer TRMNL version"; font.pixelSize: 22 }
                Text {
                    width: parent.width; wrapMode: Text.Wrap; font.pixelSize: 19
                    color: root.appState.update && root.appState.update.update_available ? "#7a4a12" : "#444"
                    text: !updateCheck.checked
                          ? "Off by default. This is the only request TRMNL makes to a server other than your dashboard, so it stays off until you turn it on."
                          : (root.appState.update
                             ? (root.appState.update.update_available
                                ? "Version " + root.appState.update.latest_version + " is available. Install it from a computer using the installer; the tablet does not update itself."
                                : "You are on the newest published version.")
                             : "Checked once a day.")
                }
                Row { spacing: 18; Text { text: "Logging"; width: 180; font.pixelSize: 22; anchors.verticalCenter: parent.verticalCenter } ComboBox { id: logLevel; width: 240; height: 64; model: ["Info", "Debug", "Warning"]; font.pixelSize: 22 } }
                Text { id: testResult; width: parent.width; font.pixelSize: 21; wrapMode: Text.Wrap }
                Row { spacing: 14; Button { text: "Test connection"; width: 240; height: 78; font.pixelSize: 22; onClicked: { testResult.text = "Testing…"; endpoint.sendMessage(3, JSON.stringify(root.configFromFields())) } } Button { text: "Save"; width: 190; height: 78; font.pixelSize: 22; onClicked: { endpoint.sendMessage(2, JSON.stringify(root.configFromFields())); root.showSettings(false) } } Button { text: "Clear cache"; width: 190; height: 78; font.pixelSize: 22; onClicked: endpoint.sendMessage(7, "") } Button { text: "Reset"; width: 150; height: 78; font.pixelSize: 22; onClicked: endpoint.sendMessage(11, "") } }
                Text { text: "Uninstall: /home/root/trmnl-remarkable/uninstall.sh\nRecovery: /home/root/trmnl-remarkable/recover-stock.sh\nRecovery tools: /home/root/trmnl-remarkable"; font.pixelSize: 19; width: parent.width; wrapMode: Text.Wrap; color: "#444" }
            }
        }
    }

    Popup {
        id: diagnosticsPopup; modal: true; focus: true; x: root.width*0.08; y: root.height*0.08; width: root.width*0.84; height: root.height*0.84
        onClosed: root.cleanScreen()
        background: Rectangle { color: "#f4f2eb"; border.width: 3 }
        contentItem: Column { spacing: 12; Text { text: "Diagnostics"; font.pixelSize: 30; font.bold: true } ScrollView { width: parent.width; height: parent.height - 100; TextArea { id: diagnosticsText; readOnly: true; wrapMode: Text.Wrap; font.pixelSize: 17 } } Button { text: "Close"; width: 180; height: 64; onClicked: diagnosticsPopup.close() } }
    }

    Rectangle {
        anchors.fill: parent
        z: 10000
        visible: root.cleaningPhase !== 0
        color: root.cleaningPhase === 1 ? "#000000" : "#ffffff"
        MouseArea { anchors.fill: parent }
    }

    ListModel { id: historyModel }
}
