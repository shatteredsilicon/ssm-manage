# PMM Manage

[![Build Status](https://travis-ci.org/shatteredsilicon/ssm-manage.svg?branch=master)](https://travis-ci.org/shatteredsilicon/ssm-manage)
[![Go Report Card](https://goreportcard.com/badge/github.com/shatteredsilicon/ssm-manage)](https://goreportcard.com/report/github.com/shatteredsilicon/ssm-manage)
[![CLA assistant](https://cla-assistant.percona.com/readme/badge/shatteredsilicon/ssm-manage)](https://cla-assistant.percona.com/shatteredsilicon/ssm-manage)

* Website: https://www.percona.com/doc/percona-monitoring-and-management/index.html
* Forum: https://www.percona.com/forums/questions-discussions/percona-monitoring-and-management/

PMM Manage is a tool for configuring options inside Percona Monitoring and Management (PMM) Server.

PMM Manage provides several key features:
* add/list/modify/remove web users
* add/list Pubic Key for SSH user access

## Building
```
export GOPATH=$(pwd)
go get -u github.com/shatteredsilicon/ssm-manage/cmd/ssm-configurator
ls -la bin/ssm-configurator
```
