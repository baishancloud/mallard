#
# spec file for mallard2 apps
# Build 2018-07-10
#

%define appName [appName]
%define package %{appName}-%{version}-%{release}
%define srcDir [srcDir]
%define destDir [destDir]
%define _rpmdir ./rpms

Summary: (%{appName})
Name: %{appName}
Version: [version]
Release: [fixversion].[arch]
License: private
Group: Applications/Server
Source: http://www.baishancloud.com/
URL: http://www.baishancloud.com/
Distribution: Centos/Redhat
Packager: fuxiaohei

%description
%{appName}

# preprocessing cmd
#%prep

# build package
%build
# go build -o %{package} %{src}

# copy file
%install
cp %{srcDir}/%{appName} %{_builddir}/%{package}
install -d %{buildroot}/usr/local/mallard/%{destDir}/var
install -d %{buildroot}/usr/local/mallard/%{destDir}/datalogs
install -d %{buildroot}/etc/logrotate.d
install -m 755 %{_builddir}/%{package} %{buildroot}/usr/local/mallard/%{destDir}/%{appName}
install -m 644 %{srcDir}/%{appName}-config.json %{buildroot}/usr/local/mallard/%{destDir}/config.example.json
install -m 644 %{srcDir}/build.logrotate %{buildroot}/usr/local/mallard/%{destDir}/mallard.logrotate
install -m 644 %{srcDir}/config.json %{buildroot}/usr/local/mallard/%{destDir}/config.json

install -d %{buildroot}/etc/supervisor/conf.d/
install -m 644 %{srcDir}/%{appName}.conf %{buildroot}/etc/supervisor/conf.d/%{appName}.conf
# install -m 644 %{addondir}/config.json %{buildroot}/usr/local/mallard/mallard-agent/config.json

# cmd before install
#%pre

# cmd after install
%post
cp usr/local/mallard/%{destDir}/mallard.logrotate /etc/logrotate.d/mallard
if [ "`ps aux|grep supervisord|grep -v grep`" != "" ]; then
	supervisorctl update
	supervisorctl restart %{appName}
else
	service supervisord start
fi

# cmd before uninstall
#%preun

# cmd after uninstall
%postun
if [ "$1" = "0" ]; then
	if [ "`ps aux|grep supervisord|grep -v grep`" != "" ]; then
                 supervisorctl stop %{appName}
                 supervisorctl update
        fi
fi

# define files in rpm
%files
/usr/local/mallard/%{destDir}/*
/etc/supervisor/conf.d/%{appName}.conf

# clean temporary files
%clean
rm -rf %{buildroot}
rm -f %_topdir/BUILD/*