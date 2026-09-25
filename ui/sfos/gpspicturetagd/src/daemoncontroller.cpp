#include "daemoncontroller.h"

#include <QDir>
#include <QFile>
#include <QFileInfo>
#include <QProcess>
#include <QRegExp>
#include <QStringList>
#include <QTextStream>

namespace {
const char *kPackagedBin = "/usr/share/gpspicturetagd/bin/gpspicturetagd";
const char *kUserTimer = "gpspicturetagd.timer";
}

DaemonController::DaemonController(QObject *parent)
    : QObject(parent)
    , m_busy(false)
    , m_running(false)
    , m_watchDir(QStringLiteral("~/Pictures"))
    , m_skipDirs(QStringLiteral("Default, Screenshots"))
    , m_maxAge(QStringLiteral("2h"))
    , m_scanEvery(QStringLiteral("5m"))
    , m_gpsTimeout(QStringLiteral("15m"))
    , m_notify(QStringLiteral("change"))
{
    m_logPath = QDir::homePath() + QStringLiteral("/.cache/gpspicturetagd.log");
    loadConfig();
    connect(&m_poll, &QTimer::timeout, this, &DaemonController::refresh);
    m_poll.start(2000);
    refresh();
    refreshLog();
}

QString DaemonController::expandHome(const QString &path) const
{
    if (path == QLatin1String("~"))
        return QDir::homePath();
    if (path.startsWith(QLatin1String("~/")))
        return QDir::homePath() + path.mid(1);
    return path;
}

QString DaemonController::configPath() const
{
    return QDir::homePath() + QStringLiteral("/.config/gpspicturetagd/config.yaml");
}

QString DaemonController::daemonBin() const
{
    return QString::fromUtf8(kPackagedBin);
}

void DaemonController::loadConfig()
{
    QFile f(configPath());
    if (!f.open(QIODevice::ReadOnly | QIODevice::Text))
        return;
    const QString text = QString::fromUtf8(f.readAll());
    f.close();

    QStringList skip;
    bool inSkip = false;
    const QStringList lines = text.split(QLatin1Char('\n'));
    for (int i = 0; i < lines.size(); ++i) {
        QString line = lines.at(i);
        const int hash = line.indexOf(QLatin1Char('#'));
        if (hash >= 0)
            line = line.left(hash);
        if (line.trimmed().isEmpty())
            continue;
        if (!line.startsWith(QLatin1Char(' ')) && !line.startsWith(QLatin1Char('\t'))
            && !line.startsWith(QLatin1String("- ")))
            inSkip = false;
        if (line.startsWith(QLatin1String("dir:"))) {
            m_watchDir = line.mid(4).trimmed();
            if (m_watchDir.startsWith(QLatin1Char('"')) && m_watchDir.endsWith(QLatin1Char('"')))
                m_watchDir = m_watchDir.mid(1, m_watchDir.size() - 2);
        } else if (line.trimmed().startsWith(QLatin1String("skip_dirs:"))) {
            inSkip = true;
        } else if (line.trimmed().startsWith(QLatin1String("max_age:"))) {
            QString v = line.mid(line.indexOf(QLatin1Char(':')) + 1).trimmed();
            if (!v.isEmpty())
                m_maxAge = v;
        } else if (line.trimmed().startsWith(QLatin1String("scan_every:"))) {
            QString v = line.mid(line.indexOf(QLatin1Char(':')) + 1).trimmed();
            if (!v.isEmpty())
                m_scanEvery = v;
        } else if (line.trimmed().startsWith(QLatin1String("gps_timeout:"))) {
            QString v = line.mid(line.indexOf(QLatin1Char(':')) + 1).trimmed();
            if (!v.isEmpty())
                m_gpsTimeout = v;
        } else if (line.trimmed().startsWith(QLatin1String("notify:"))) {
            QString v = line.mid(line.indexOf(QLatin1Char(':')) + 1).trimmed();
            if (!v.isEmpty())
                m_notify = v;
        } else if (inSkip) {
            QString item = line.trimmed();
            if (item.startsWith(QLatin1String("- ")))
                skip.append(item.mid(2).trimmed());
        }
    }
    if (!skip.isEmpty())
        m_skipDirs = skip.join(QStringLiteral(", "));
}

bool DaemonController::writeConfig() const
{
    const QString path = configPath();
    QDir().mkpath(QFileInfo(path).absolutePath());
    QFile f(path);
    if (!f.open(QIODevice::WriteOnly | QIODevice::Truncate | QIODevice::Text))
        return false;
    QTextStream out(&f);
    out << "dir: " << m_watchDir << "\n";
    out << "skip_dirs:\n";
    const QStringList parts = m_skipDirs.split(QLatin1Char(','), QString::SkipEmptyParts);
    for (int i = 0; i < parts.size(); ++i)
        out << "  - " << parts.at(i).trimmed() << "\n";
    out << "max_age: " << m_maxAge << "\n";
    out << "scan_every: " << m_scanEvery << "\n";
    out << "gps_timeout: " << m_gpsTimeout << "\n";
    out << "notify: " << m_notify << "\n";
    return true;
}

void DaemonController::setWatchDir(const QString &v)
{
    if (m_watchDir == v)
        return;
    m_watchDir = v;
    writeConfig();
    emit settingsChanged();
    reloadDaemon();
}

void DaemonController::setSkipDirs(const QString &v)
{
    if (m_skipDirs == v)
        return;
    m_skipDirs = v;
    writeConfig();
    emit settingsChanged();
    reloadDaemon();
}

void DaemonController::setMaxAge(const QString &v)
{
    if (m_maxAge == v)
        return;
    m_maxAge = v;
    writeConfig();
    emit settingsChanged();
    reloadDaemon();
}

void DaemonController::setScanEvery(const QString &v)
{
    if (m_scanEvery == v)
        return;
    m_scanEvery = v;
    writeConfig();
    emit settingsChanged();
    reloadDaemon();
}

void DaemonController::setGpsTimeout(const QString &v)
{
    if (m_gpsTimeout == v)
        return;
    m_gpsTimeout = v;
    writeConfig();
    emit settingsChanged();
    reloadDaemon();
}

void DaemonController::setNotify(const QString &v)
{
    if (m_notify == v)
        return;
    m_notify = v;
    writeConfig();
    emit settingsChanged();
    reloadDaemon();
}

void DaemonController::reloadDaemon()
{
    if (m_busy || !processRunning())
        return;
    startDaemon();
}

bool DaemonController::processRunning() const
{
    QProcess p;
    p.start(QStringLiteral("systemctl"),
            QStringList() << QStringLiteral("--user") << QStringLiteral("is-active")
                          << QString::fromUtf8(kUserTimer));
    if (!p.waitForFinished(2000)) {
        p.kill();
        return false;
    }
    return p.exitCode() == 0;
}

void DaemonController::refresh()
{
    const bool run = processRunning();
    setRunning(run);
    if (run)
        setStatusText(m_lastFix.isEmpty() ? QStringLiteral("Timer enabled") : m_lastFix);
    else
        setStatusText(QStringLiteral("Stopped"));
}

void DaemonController::parseLog(const QString &log)
{
    QRegExp gpsRe(QStringLiteral("lat=([\\-0-9.]+) lon=([\\-0-9.]+) acc=([0-9.]+)m.*src=(\\S+)"));
    int pos = 0;
    int lastGps = -1;
    while ((pos = gpsRe.indexIn(log, pos)) >= 0) {
        lastGps = pos;
        pos += gpsRe.matchedLength();
    }
    if (lastGps >= 0) {
        gpsRe.indexIn(log, lastGps);
        setLastFix(QStringLiteral("%1, %2 · %3 m · %4")
                       .arg(gpsRe.cap(1), gpsRe.cap(2), gpsRe.cap(3), gpsRe.cap(4)));
    }

    QRegExp tagRe(QStringLiteral("tagged (.+) lat=([\\-0-9.]+) lon=([\\-0-9.]+)"));
    pos = 0;
    int lastTag = -1;
    while ((pos = tagRe.indexIn(log, pos)) >= 0) {
        lastTag = pos;
        pos += tagRe.matchedLength();
    }
    if (lastTag >= 0) {
        tagRe.indexIn(log, lastTag);
        QString path = tagRe.cap(1);
        const int slash = path.lastIndexOf(QLatin1Char('/'));
        if (slash >= 0)
            path = path.mid(slash + 1);
        setLastTagged(QStringLiteral("%1 · %2, %3").arg(path, tagRe.cap(2), tagRe.cap(3)));
    }
}

void DaemonController::startDaemon()
{
    if (m_busy)
        return;
    setBusy(true);
    writeConfig();
    QDir().mkpath(QFileInfo(m_logPath).absolutePath());
    QProcess::execute(daemonBin(),
                      QStringList() << QStringLiteral("-install") << QStringLiteral("-log")
                                    << m_logPath);
    refresh();
    refreshLog();
    setBusy(false);
}

void DaemonController::stopDaemon()
{
    if (m_busy)
        return;
    setBusy(true);
    QProcess::execute(daemonBin(),
                      QStringList() << QStringLiteral("-uninstall") << QStringLiteral("-log")
                                    << m_logPath);
    QProcess::execute(QStringLiteral("systemctl"),
                      QStringList() << QStringLiteral("--user") << QStringLiteral("disable")
                                    << QStringLiteral("--now") << QString::fromUtf8(kUserTimer));
    QProcess::execute(QStringLiteral("systemctl"),
                      QStringList() << QStringLiteral("--user") << QStringLiteral("disable")
                                    << QStringLiteral("--now")
                                    << QStringLiteral("gpspicturetagd.service"));
    QProcess::execute(QStringLiteral("pkill"),
                      QStringList() << QStringLiteral("-f")
                                    << QString::fromUtf8(kPackagedBin));
    refresh();
    setBusy(false);
}

void DaemonController::refreshLog()
{
    QFile f(m_logPath);
    if (!f.open(QIODevice::ReadOnly | QIODevice::Text)) {
        if (m_logText != QString() || !m_logLines.isEmpty()) {
            m_logText.clear();
            m_logLines.clear();
            emit logTextChanged();
        }
        return;
    }
    QString text = QString::fromUtf8(f.readAll());
    f.close();
    const int maxChars = 64000;
    if (text.size() > maxChars)
        text = text.right(maxChars);
    if (text == m_logText)
        return;
    m_logText = text;
    QStringList lines = text.split(QLatin1Char('\n'), QString::SkipEmptyParts);
    const int maxLines = 400;
    if (lines.size() > maxLines)
        lines = lines.mid(lines.size() - maxLines);
    m_logLines = lines;
    emit logTextChanged();
    parseLog(text);
}

void DaemonController::clearLog()
{
    QFile f(m_logPath);
    f.open(QIODevice::WriteOnly | QIODevice::Truncate);
    f.close();
    m_logText.clear();
    m_logLines.clear();
    emit logTextChanged();
}

void DaemonController::setBusy(bool v)
{
    if (m_busy == v)
        return;
    m_busy = v;
    emit busyChanged();
}

void DaemonController::setRunning(bool v)
{
    if (m_running == v)
        return;
    m_running = v;
    emit runningChanged();
}

void DaemonController::setStatusText(const QString &v)
{
    if (m_statusText == v)
        return;
    m_statusText = v;
    emit statusTextChanged();
}

void DaemonController::setLastFix(const QString &v)
{
    if (m_lastFix == v)
        return;
    m_lastFix = v;
    emit lastFixChanged();
}

void DaemonController::setLastTagged(const QString &v)
{
    if (m_lastTagged == v)
        return;
    m_lastTagged = v;
    emit lastTaggedChanged();
}
