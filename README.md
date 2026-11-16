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

This repo includes the OPI MangoBoost bridge, which serves the OPI storage APIs for the MangoBoost BoostX NVMe-oF/RDMA target (NRT).
The bridge translates OPI API calls into JSON-RPC calls to the MangoBoost SPDK target, which offloads the NVMe-oF/RDMA data path to the BoostX hardware.

At startup the bridge creates the SPDK RDMA transport with hardware offload, unless one already exists, and optionally adds an RDMA discovery listener.
Each OPI NVMe controller with the RDMA transport becomes an RDMA listener of its subsystem.
Volumes come from the AIO, Null and Malloc volume services or from remote NVMe controllers.

## Getting started

:exclamation: `docker-compose` is deprecated. For details, see [Migrate to Compose V2](https://docs.docker.com/compose/migrate/).

Run `docker-compose up -d` or `docker compose up -d`

## Usage

Start the MangoBoost SPDK target, then the bridge:

```bash
opi-mangoboost-bridge -spdk_addr=/var/tmp/spdk.sock -redis_addr=127.0.0.1:6379 \
  -discovery_addr=192.168.1.10:8009 -min_cntlid=1 -max_cntlid=999
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `-offload` | `true` | Enable BoostX hardware offload on the RDMA transport; `false` works with stock SPDK |
| `-discovery_addr` | empty | RDMA discovery listener in `ip:port` form; empty skips it |
| `-min_cntlid`, `-max_cntlid` | `0` | Controller ID range of new subsystems, so that targets exporting the same subsystem do not collide; `0` keeps the SPDK default |

Export a Malloc volume over NVMe-oF/RDMA:

```bash
grpcurl -plaintext -d '{"malloc_volume_id": "mal0", "malloc_volume": {"block_size": 512, "blocks_count": 262144}}' \
  localhost:50051 opi_api.storage.v1.MallocVolumeService.CreateMallocVolume
grpcurl -plaintext -d '{"nvme_subsystem_id": "subsys0", "nvme_subsystem": {"spec": {"nqn": "nqn.2022-09.io.spdk:opi0", "serial_number": "SN0001", "model_number": "BoostX NRT", "max_namespaces": 32}}}' \
  localhost:50051 opi_api.storage.v1.FrontendNvmeService.CreateNvmeSubsystem
grpcurl -plaintext -d '{"parent": "nvmeSubsystems/subsys0", "nvme_namespace_id": "ns0", "nvme_namespace": {"spec": {"volume_name_ref": "mal0", "host_nsid": 1}}}' \
  localhost:50051 opi_api.storage.v1.FrontendNvmeService.CreateNvmeNamespace
grpcurl -plaintext -d '{"parent": "nvmeSubsystems/subsys0", "nvme_controller_id": "ctrl0", "nvme_controller": {"spec": {"trtype": "NVME_TRANSPORT_TYPE_RDMA", "fabrics_id": {"traddr": "192.168.1.10", "trsvcid": "4420", "adrfam": "NVME_ADDRESS_FAMILY_IPV4"}}}}' \
  localhost:50051 opi_api.storage.v1.FrontendNvmeService.CreateNvmeController
```

Connect from a host with `nvme connect -t rdma -a 192.168.1.10 -s 4420 -n nqn.2022-09.io.spdk:opi0`.

## I Want To Contribute

This project welcomes contributions and suggestions.  We are happy to have the Community involved via submission of **Issues and Pull Requests** (with substantive content or even just fixes). We are hoping for the documents, test framework, etc. to become a community process with active engagement.  PRs can be reviewed by by any number of people, and a maintainer may accept.

See [CONTRIBUTING](https://github.com/opiproject/opi/blob/main/CONTRIBUTING.md) and [GitHub Basic Process](https://github.com/opiproject/opi/blob/main/doc-github-rules.md) for more details.
