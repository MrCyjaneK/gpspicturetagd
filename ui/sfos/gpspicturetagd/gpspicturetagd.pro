TARGET = gpspicturetagd

CONFIG += sailfishapp c++11

QT += core qml quick

SOURCES += \
    src/main.cpp \
    src/daemoncontroller.cpp

HEADERS += \
    src/daemoncontroller.h

OTHER_FILES += \
    qml/*.qml \
    qml/pages/*.qml \
    qml/cover/*.qml \
    gpspicturetagd.desktop \
    icons/86x86/gpspicturetagd.png \
    icons/108x108/gpspicturetagd.png \
    icons/128x128/gpspicturetagd.png \
    icons/172x172/gpspicturetagd.png \
    icons/256x256/gpspicturetagd.png \
    icons/512x512/gpspicturetagd.png

desktop.files = gpspicturetagd.desktop

SAILFISHAPP_ICONS = 86x86 108x108 128x128 172x172 256x256 512x512
