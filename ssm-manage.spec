%define debug_package %{nil}

%global _dwz_low_mem_die_limit 0

%global provider        github
%global provider_tld	com
%global project         shatteredsilicon
%global repo            ssm-manage
%global provider_prefix	%{provider}.%{provider_tld}/%{project}/%{repo}

Name:		%{repo}
Version:	%{_version}
Release:	1%{?dist}
Summary:	SSM configuration managament tool

License:	AGPLv3
URL:		https://%{provider_prefix}
Source0:	%{name}-%{version}.tar.gz

BuildRequires:	golang

%if 0%{?fedora} || 0%{?rhel} == 7
BuildRequires: systemd
Requires(post): systemd
Requires(preun): systemd
Requires(postun): systemd
%endif

%description
SSM Manage tool is part of Shattered Silicon Monitoring and Management.
See the SSM docs for more information.


%prep
%setup -q -n %{repo}
mkdir -p src/%{provider}.%{provider_tld}/%{project}
ln -s $(pwd) src/%{provider_prefix}


%build
export GOPATH=$(pwd)
GO111MODULE=off go build -ldflags "${LDFLAGS:-} -s -w -B 0x$(head -c20 /dev/urandom|od -An -tx1|tr -d ' \n')" -a -v -x %{provider_prefix}/cmd/ssm-configure
GO111MODULE=off go build -ldflags "${LDFLAGS:-} -s -w -B 0x$(head -c20 /dev/urandom|od -An -tx1|tr -d ' \n')" -a -v -x %{provider_prefix}/cmd/ssm-configurator


%install
install -d -p %{buildroot}%{_bindir}
install -d -p %{buildroot}%{_sbindir}
install -p -m 0755 ssm-configure    %{buildroot}%{_bindir}/ssm-configure
install -p -m 0755 ssm-configurator %{buildroot}%{_sbindir}/ssm-configurator

install -d %{buildroot}/usr/lib/systemd/system
install -p -m 0644 packaging/ssm-manage.service %{buildroot}/usr/lib/systemd/system/%{name}.service


%post
%systemd_post %{name}.service

%preun
%systemd_preun %{name}.service

%postun
%systemd_postun %{name}.service


%files
%license src/%{provider_prefix}/LICENSE
%doc src/%{provider_prefix}/README.md
%{_bindir}/ssm-configure
%{_sbindir}/ssm-configurator
/usr/lib/systemd/system/%{name}.service


%changelog
* Fri Jun 30 2017 Mykola Marzhan <mykola.marzhan@percona.com> - 1.1.6-1
- move repository from Percona-Lab to percona organization

* Fri Mar  3 2017 Mykola Marzhan <mykola.marzhan@percona.com> - 1.1.1-1
- add pmm-configure

* Fri Feb  3 2017 Mykola Marzhan <mykola.marzhan@percona.com> - 1.1.0-2
- add build_timestamp to Release value

* Wed Feb  1 2017 Mykola Marzhan <mykola.marzhan@percona.com> - 1.1.0-1
- init version
