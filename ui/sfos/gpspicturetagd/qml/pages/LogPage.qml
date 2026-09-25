import QtQuick 2.0
import Sailfish.Silica 1.0

Page {
    id: page

    onStatusChanged: {
        if (status === PageStatus.Active) {
            Daemon.refreshLog()
            list.positionViewAtEnd()
        }
    }

    SilicaListView {
        id: list
        anchors.fill: parent
        clip: true
        model: Daemon.logLines
        spacing: Theme.paddingSmall

        header: PageHeader { title: qsTr("Log") }

        PullDownMenu {
            MenuItem {
                text: qsTr("Clear log")
                onClicked: Daemon.clearLog()
            }
            MenuItem {
                text: qsTr("Refresh")
                onClicked: {
                    Daemon.refreshLog()
                    list.positionViewAtEnd()
                }
            }
        }

        ViewPlaceholder {
            enabled: list.count === 0
            text: qsTr("No log yet.")
        }

        delegate: Item {
            width: list.width
            height: lineLabel.height

            Label {
                id: lineLabel
                x: Theme.horizontalPageMargin
                width: parent.width - 2 * Theme.horizontalPageMargin
                wrapMode: Text.WrapAnywhere
                font.pixelSize: Theme.fontSizeExtraSmall
                font.family: "monospace"
                color: Theme.primaryColor
                text: modelData
            }
        }

        VerticalScrollDecorator {}

        onCountChanged: {
            if (page.status === PageStatus.Active)
                positionViewAtEnd()
        }
    }

    Timer {
        interval: 3000
        running: page.status === PageStatus.Active
        repeat: true
        onTriggered: Daemon.refreshLog()
    }
}
