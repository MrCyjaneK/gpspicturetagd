#include <sailfishapp.h>

#include <QGuiApplication>
#include <QQuickView>
#include <QQmlContext>

#include "daemoncontroller.h"

Q_DECL_EXPORT int main(int argc, char *argv[])
{
    QScopedPointer<QGuiApplication> app(SailfishApp::application(argc, argv));
    QScopedPointer<QQuickView> view(SailfishApp::createView());

    DaemonController daemon;
    view->rootContext()->setContextProperty(QStringLiteral("Daemon"), &daemon);

    view->setSource(SailfishApp::pathToMainQml());
    view->show();
    return app->exec();
}
