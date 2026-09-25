#ifndef DAEMONCONTROLLER_H
#define DAEMONCONTROLLER_H

#include <QObject>
#include <QString>
#include <QStringList>
#include <QTimer>

class DaemonController : public QObject
{
    Q_OBJECT
    Q_PROPERTY(bool busy READ busy NOTIFY busyChanged)
    Q_PROPERTY(bool running READ running NOTIFY runningChanged)
    Q_PROPERTY(QString statusText READ statusText NOTIFY statusTextChanged)
    Q_PROPERTY(QString lastFix READ lastFix NOTIFY lastFixChanged)
    Q_PROPERTY(QString lastTagged READ lastTagged NOTIFY lastTaggedChanged)
    Q_PROPERTY(QString logText READ logText NOTIFY logTextChanged)
    Q_PROPERTY(QStringList logLines READ logLines NOTIFY logTextChanged)
    Q_PROPERTY(QString logPath READ logPath CONSTANT)
    Q_PROPERTY(QString configPath READ configPath CONSTANT)
    Q_PROPERTY(QString binaryPath READ daemonBin CONSTANT)
    Q_PROPERTY(QString watchDir READ watchDir WRITE setWatchDir NOTIFY settingsChanged)
    Q_PROPERTY(QString skipDirs READ skipDirs WRITE setSkipDirs NOTIFY settingsChanged)
    Q_PROPERTY(QString maxAge READ maxAge WRITE setMaxAge NOTIFY settingsChanged)
    Q_PROPERTY(QString scanEvery READ scanEvery WRITE setScanEvery NOTIFY settingsChanged)
    Q_PROPERTY(QString gpsTimeout READ gpsTimeout WRITE setGpsTimeout NOTIFY settingsChanged)
    Q_PROPERTY(QString notify READ notify WRITE setNotify NOTIFY settingsChanged)

public:
    explicit DaemonController(QObject *parent = 0);

    bool busy() const { return m_busy; }
    bool running() const { return m_running; }
    QString statusText() const { return m_statusText; }
    QString lastFix() const { return m_lastFix; }
    QString lastTagged() const { return m_lastTagged; }
    QString logText() const { return m_logText; }
    QStringList logLines() const { return m_logLines; }
    QString logPath() const { return m_logPath; }
    QString configPath() const;
    QString daemonBin() const;

    QString watchDir() const { return m_watchDir; }
    void setWatchDir(const QString &v);

    QString skipDirs() const { return m_skipDirs; }
    void setSkipDirs(const QString &v);

    QString maxAge() const { return m_maxAge; }
    void setMaxAge(const QString &v);

    QString scanEvery() const { return m_scanEvery; }
    void setScanEvery(const QString &v);

    QString gpsTimeout() const { return m_gpsTimeout; }
    void setGpsTimeout(const QString &v);

    QString notify() const { return m_notify; }
    void setNotify(const QString &v);

    Q_INVOKABLE void refresh();
    Q_INVOKABLE void startDaemon();
    Q_INVOKABLE void stopDaemon();
    Q_INVOKABLE void refreshLog();
    Q_INVOKABLE void clearLog();

signals:
    void busyChanged();
    void runningChanged();
    void statusTextChanged();
    void lastFixChanged();
    void lastTaggedChanged();
    void logTextChanged();
    void settingsChanged();

private:
    void loadConfig();
    bool writeConfig() const;
    void reloadDaemon();
    QString expandHome(const QString &path) const;
    bool processRunning() const;
    void parseLog(const QString &log);
    void setBusy(bool v);
    void setRunning(bool v);
    void setStatusText(const QString &v);
    void setLastFix(const QString &v);
    void setLastTagged(const QString &v);

    QTimer m_poll;
    bool m_busy;
    bool m_running;
    QString m_statusText;
    QString m_lastFix;
    QString m_lastTagged;
    QString m_logText;
    QStringList m_logLines;
    QString m_logPath;
    QString m_watchDir;
    QString m_skipDirs;
    QString m_maxAge;
    QString m_scanEvery;
    QString m_gpsTimeout;
    QString m_notify;
};

#endif
