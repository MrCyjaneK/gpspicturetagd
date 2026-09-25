import QtQuick 2.0
import Sailfish.Silica 1.0

CoverBackground {
    Column {
        anchors.centerIn: parent
        width: parent.width - 2 * Theme.paddingMedium
        spacing: Theme.paddingSmall

        Label {
            width: parent.width
            horizontalAlignment: Text.AlignHCenter
            color: Theme.primaryColor
            font.pixelSize: Theme.fontSizeMedium
            wrapMode: Text.Wrap
            text: qsTr("GPS Picture Tag")
        }
        Label {
            width: parent.width
            horizontalAlignment: Text.AlignHCenter
            wrapMode: Text.Wrap
            color: Daemon.running ? Theme.highlightColor : Theme.secondaryColor
            text: Daemon.running ? qsTr("Timer on") : qsTr("Timer off")
        }
        Label {
            width: parent.width
            horizontalAlignment: Text.AlignHCenter
            wrapMode: Text.Wrap
            color: Theme.secondaryColor
            font.pixelSize: Theme.fontSizeExtraSmall
            visible: Daemon.lastFix.length > 0
            text: Daemon.lastFix
        }
    }

    CoverActionList {
        CoverAction {
            iconSource: Daemon.running
                        ? "image://theme/icon-cover-cancel"
                        : "image://theme/icon-cover-play"
            onTriggered: {
                if (Daemon.busy)
                    return
                Daemon.running ? Daemon.stopDaemon() : Daemon.startDaemon()
            }
        }
    }
}
