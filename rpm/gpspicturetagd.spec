Name:       gpspicturetagd

%global _missing_build_ids_terminate_build 0

%{!?qtc_qmake5:%define qtc_qmake5 %qmake5}
%{!?qtc_make:%define qtc_make make}

Summary:    GPS Picture Tag
Version:    0.1
Release:    2
Group:      Applications/System
License:    GPLv3
Source0:    %{name}-%{version}.tar.bz2
Requires:   mapplauncherd-booster-silica-qt5
Requires:   sailfishsilica-qt5
BuildRequires:  pkgconfig(sailfishapp)
BuildRequires:  pkgconfig(Qt5Quick)
BuildRequires:  pkgconfig(Qt5Qml)
BuildRequires:  pkgconfig(Qt5Core)
BuildRequires:  pkgconfig(qdeclarative5-boostable)
BuildRequires:  desktop-file-utils

%description
Tag pictures you took later, after GPS locks

%prep
%setup -q -n %{name}-%{version}

%build
export CGO_ENABLED=0 GOTOOLCHAIN=local
export GOCACHE=%{_builddir}/.gocache GOPATH=%{_builddir}/.gopath
export GOPROXY=https://proxy.golang.org,direct
go build -ldflags "-s -w -X main.version=%{version}" \
	-o gpspicturetagd-daemon ./cmd/gpspicturetagd

cd ui/sfos/gpspicturetagd
%qtc_qmake5
%qtc_make %{?_smp_mflags}

%install
rm -rf %{buildroot}
make -C ui/sfos/gpspicturetagd install INSTALL_ROOT=%{buildroot}
install -D -m 755 gpspicturetagd-daemon %{buildroot}%{_datadir}/%{name}/bin/gpspicturetagd
install -D -m 644 config/gpspicturetagd.yaml %{buildroot}%{_datadir}/%{name}/config.yaml
for sz in 86x86 108x108 128x128 172x172 256x256 512x512; do
	install -D -m 644 icons/${sz}/%{name}.png \
		%{buildroot}%{_datadir}/icons/hicolor/${sz}/apps/%{name}.png
done

desktop-file-install --delete-original \
	--dir %{buildroot}%{_datadir}/applications \
	--set-key=Icon --set-value=%{name} \
	%{buildroot}%{_datadir}/applications/*.desktop

%files
%defattr(-,root,root,-)
%{_datadir}/applications/%{name}.desktop
%{_datadir}/icons/hicolor/*/apps/%{name}.png
%{_datadir}/%{name}/qml
%dir %{_datadir}/%{name}/bin
%{_datadir}/%{name}/bin/gpspicturetagd
%{_bindir}/%{name}
%{_datadir}/%{name}/config.yaml
