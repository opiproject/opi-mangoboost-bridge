# OPI gRPC to Mangoboost Bridge

[![Linters](https://github.com/opiproject/opi-mangoboost-bridge/actions/workflows/linters.yml/badge.svg)](https://github.com/opiproject/mangoboost-bridge/actions/workflows/linters.yml)
[![CodeQL](https://github.com/opiproject/mangoboost-bridge/actions/workflows/codeql.yml/badge.svg)](https://github.com/opiproject/mangoboost-bridge/actions/workflows/codeql.yml)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/opiproject/mangoboost-bridge/badge)](https://securityscorecards.dev/viewer/?platform=github.com&org=opiproject&repo=opi-mangoboost-bridge)
[![tests](https://github.com/opiproject/opi-mangoboost-bridge/actions/workflows/go.yml/badge.svg)](https://github.com/opiproject/opi-mangoboost-bridge/actions/workflows/go.yml)
[![Docker](https://github.com/opiproject/opi-mangoboost-bridge/actions/workflows/docker-publish.yml/badge.svg)](https://github.com/opiproject/opi-mangoboost-bridge/actions/workflows/docker-publish.yml)
[![License](https://img.shields.io/github/license/opiproject/opi-mangoboost-bridge?style=flat-square&color=blue&label=License)](https://github.com/opiproject/opi-mangoboost-bridge/blob/master/LICENSE)
[![codecov](https://codecov.io/gh/opiproject/opi-mangoboost-bridge/branch/main/graph/badge.svg)](https://codecov.io/gh/opiproject/opi-mangoboost-bridge)
[![Go Report Card](https://goreportcard.com/badge/github.com/opiproject/opi-mangoboost-bridge)](https://goreportcard.com/report/github.com/opiproject/opi-mangoboost-bridge)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/opiproject/opi-mangoboost-bridge)
[![Pulls](https://img.shields.io/docker/pulls/opiproject/opi-mangoboost-bridge.svg?logo=docker&style=flat&label=Pulls)](https://hub.docker.com/r/opiproject/opi-mangoboost-bridge)
[![Last Release](https://img.shields.io/github/v/release/opiproject/opi-mangoboost-bridge?label=Latest&style=flat-square&logo=go)](https://github.com/opiproject/opi-mangoboost-bridge/releases)
[![GitHub stars](https://img.shields.io/github/stars/opiproject/opi-mangoboost-bridge.svg?style=flat-square&label=github%20stars)](https://github.com/opiproject/opi-mangoboost-bridge)
[![GitHub Contributors](https://img.shields.io/github/contributors/opiproject/opi-mangoboost-bridge.svg?style=flat-square)](https://github.com/opiproject/opi-mangoboost-bridge/graphs/contributors)

This repo includes OPI Mangoboost bridge API for DPUs, particularly OPI storage APIs for NVMe/TCP initiator and target offloads.
For additional information, please refer to the links below.

- [Mango StorageBoost - NTI](https://www.mangoboost.io/products/hardware/mango-storageboost-tm-nti)
- [Mango StorageBoost - NTT](https://www.mangoboost.io/products/hardware/mango-storageboost-tm-ntt)

The diagram below demonstrates an example of OPI-enabled workflows for xPU-based NVMe-oF initiator and target deployments.
Running on an xPU, the bridge translates and forwards OPI API commands to the MangoBoost SDK for service management and configuration.
On both the initiator and target, the NVMe/TCP data path is fully offloaded to the xPU hardware without any involvement of host and SoC CPU cores.

![opi-mangoboost-bridge system overview](doc/images/opi-mangoboost-bridge_system-overview.png "opi-mangoboost-bridge system overview")

## Getting started

:exclamation: `docker-compose` is deprecated. For details, see [Migrate to Compose V2](https://docs.docker.com/compose/migrate/).

Run `docker-compose up -d` or `docker compose up -d`

## Usage

TBU

## I Want To Contribute

This project welcomes contributions and suggestions.  We are happy to have the Community involved via submission of **Issues and Pull Requests** (with substantive content or even just fixes). We are hoping for the documents, test framework, etc. to become a community process with active engagement.  PRs can be reviewed by by any number of people, and a maintainer may accept.

See [CONTRIBUTING](https://github.com/opiproject/opi/blob/main/CONTRIBUTING.md) and [GitHub Basic Process](https://github.com/opiproject/opi/blob/main/doc-github-rules.md) for more details.
