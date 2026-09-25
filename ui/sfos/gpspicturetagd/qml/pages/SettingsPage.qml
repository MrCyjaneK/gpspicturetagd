import QtQuick 2.0
import Sailfish.Silica 1.0

Page {
    id: page
    property bool _ready: false
    Component.onCompleted: _ready = true

    SilicaFlickable {
        anchors.fill: parent
        contentHeight: column.height + Theme.paddingLarge

        Column {
            id: column
            width: parent.width

            PageHeader { title: qsTr("Settings") }

            TextField {
                width: parent.width
                label: qsTr("Pictures directory")
                text: Daemon.watchDir
                onFocusChanged: if (!focus && _ready) Daemon.watchDir = text
                EnterKey.onClicked: { Daemon.watchDir = text; focus = false }
            }

            TextField {
                width: parent.width
                label: qsTr("Skip directories")
                text: Daemon.skipDirs
                placeholderText: qsTr("Default, Screenshots")
                onFocusChanged: if (!focus && _ready) Daemon.skipDirs = text
                EnterKey.onClicked: { Daemon.skipDirs = text; focus = false }
            }

            TextField {
                width: parent.width
                label: qsTr("Max age")
                text: Daemon.maxAge
                placeholderText: qsTr("2h")
                onFocusChanged: if (!focus && _ready) Daemon.maxAge = text
                EnterKey.onClicked: { Daemon.maxAge = text; focus = false }
            }
            Label {
                x: Theme.horizontalPageMargin
                width: parent.width - 2 * Theme.horizontalPageMargin
                wrapMode: Text.Wrap
                color: Theme.secondaryColor
                font.pixelSize: Theme.fontSizeExtraSmall
                text: qsTr("Only tag photos newer than this (Go duration, e.g. 2h, 30m). Older files are never rewritten.")
            }

            TextField {
                width: parent.width
                label: qsTr("Scan interval")
                text: Daemon.scanEvery
                placeholderText: qsTr("5m")
                onFocusChanged: if (!focus && _ready) Daemon.scanEvery = text
                EnterKey.onClicked: { Daemon.scanEvery = text; focus = false }
            }
            Label {
                x: Theme.horizontalPageMargin
                width: parent.width - 2 * Theme.horizontalPageMargin
                wrapMode: Text.Wrap
                color: Theme.secondaryColor
                font.pixelSize: Theme.fontSizeExtraSmall
                text: qsTr("How often to start the process")
            }

            TextField {
                width: parent.width
                label: qsTr("GPS timeout")
                text: Daemon.gpsTimeout
                placeholderText: qsTr("15m")
                onFocusChanged: if (!focus && _ready) Daemon.gpsTimeout = text
                EnterKey.onClicked: { Daemon.gpsTimeout = text; focus = false }
            }
            Label {
                x: Theme.horizontalPageMargin
                width: parent.width - 2 * Theme.horizontalPageMargin
                wrapMode: Text.Wrap
                color: Theme.secondaryColor
                font.pixelSize: Theme.fontSizeExtraSmall
                text: qsTr("How long GPS may stay on while waiting for a lock. Prefers GPS over cell/Wi-Fi if it arrives in time.")
            }

            ComboBox {
                id: notifyBox
                width: parent.width
                label: qsTr("Notifications")
                currentIndex: {
                    switch (Daemon.notify) {
                    case "always": return 1
                    case "error": return 2
                    case "never": return 3
                    default: return 0
                    }
                }
                menu: ContextMenu {
                    MenuItem { text: qsTr("When something changed") }
                    MenuItem { text: qsTr("Always") }
                    MenuItem { text: qsTr("On error") }
                    MenuItem { text: qsTr("Never") }
                }
                onCurrentIndexChanged: {
                    if (!_ready)
                        return
                    var values = ["change", "always", "error", "never"]
                    Daemon.notify = values[currentIndex]
                }
            }
            Label {
                x: Theme.horizontalPageMargin
                width: parent.width - 2 * Theme.horizontalPageMargin
                wrapMode: Text.Wrap
                color: Theme.secondaryColor
                font.pixelSize: Theme.fontSizeExtraSmall
                text: qsTr("After each scan: banner when photos were tagged or something failed, every time, only on failure, or never.")
            }

            SectionHeader { text: qsTr("Paths") }
            DetailItem { label: qsTr("Daemon"); value: Daemon.binaryPath }
            DetailItem { label: qsTr("Log"); value: Daemon.logPath }
            DetailItem { label: qsTr("Config"); value: Daemon.configPath }
        }

        VerticalScrollDecorator {}
    }
}
