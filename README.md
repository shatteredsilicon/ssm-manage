# SSM Manage

[![Go Report Card](https://goreportcard.com/badge/github.com/shatteredsilicon/ssm-manage)](https://goreportcard.com/report/github.com/shatteredsilicon/ssm-manage)

* Website: https://shatteredsilicon.net/software/ssm/documentation/latest/

SSM Manage is a tool for configuring options inside Shattered Silicon Monitoring (SSM) Server.

SSM Manage provides several key features:
* add/list/modify/remove web users
* add/list Pubic Key for SSH user access

## Building
```
export GOPATH=$(pwd)
go get -u github.com/shatteredsilicon/ssm-manage/cmd/ssm-configurator
ls -la bin/ssm-configurator
```
