import QtQuick 2.0
import Sailfish.Silica 1.0

Page {
    id: page

    SilicaFlickable {
        anchors.fill: parent
        contentHeight: column.height

        PullDownMenu {
            MenuItem {
                text: qsTr("Settings")
                onClicked: pageStack.push(Qt.resolvedUrl("SettingsPage.qml"))
            }
            MenuItem {
                text: qsTr("Log")
                onClicked: pageStack.push(Qt.resolvedUrl("LogPage.qml"))
            }
            MenuItem {
                text: Daemon.running ? qsTr("Disable timer") : qsTr("Enable timer")
                enabled: !Daemon.busy
                onClicked: Daemon.running ? Daemon.stopDaemon() : Daemon.startDaemon()
            }
        }

        Column {
            id: column
            width: parent.width
            spacing: Theme.paddingMedium

            PageHeader { title: qsTr("GPS Picture Tag") }

            Label {
                x: Theme.horizontalPageMargin
                width: parent.width - 2 * Theme.horizontalPageMargin
                wrapMode: Text.Wrap
                color: Theme.highlightColor
                font.pixelSize: Theme.fontSizeLarge
                text: Daemon.running ? qsTr("Timer enabled") : qsTr("Timer off")
            }

            Label {
                x: Theme.horizontalPageMargin
                width: parent.width - 2 * Theme.horizontalPageMargin
                wrapMode: Text.Wrap
                color: Theme.secondaryColor
                font.pixelSize: Theme.fontSizeSmall
                text: Daemon.statusText
            }

            DetailItem {
                label: qsTr("Last lock")
                value: Daemon.lastFix.length ? Daemon.lastFix : qsTr("-")
            }
            DetailItem {
                label: qsTr("Last tagged")
                value: Daemon.lastTagged.length ? Daemon.lastTagged : qsTr("-")
            }
            DetailItem {
                label: qsTr("Watch")
                value: Daemon.watchDir
            }
            DetailItem {
                label: qsTr("Daemon")
                value: Daemon.binaryPath
            }
            DetailItem {
                label: qsTr("Log")
                value: Daemon.logPath
            }
            DetailItem {
                label: qsTr("Config")
                value: Daemon.configPath
            }

            Item { width: 1; height: Theme.paddingLarge }

            Button {
                anchors.horizontalCenter: parent.horizontalCenter
                preferredWidth: Theme.buttonWidthLarge
                enabled: !Daemon.busy
                text: Daemon.running ? qsTr("Disable timer") : qsTr("Enable timer")
                onClicked: Daemon.running ? Daemon.stopDaemon() : Daemon.startDaemon()
            }
            Button {
                anchors.horizontalCenter: parent.horizontalCenter
                preferredWidth: Theme.buttonWidthLarge
                text: qsTr("View log")
                onClicked: pageStack.push(Qt.resolvedUrl("LogPage.qml"))
            }
        }

        VerticalScrollDecorator {}
    }
}
