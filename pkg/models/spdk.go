// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2023 Intel Corporation
// Copyright (C) 2025 MangoBoost, Inc.

// Package models holds definitions for SPDK json RPC structs
package models

import "github.com/opiproject/gospdk/spdk"

// NhiNvmfSubsystemAddListenerParams holds the parameters required to Delete a NVMf subsystem
type NhiNvmfSubsystemAddListenerParams struct {
	spdk.NvmfSubsystemAddListenerParams
	NumQueues int `json:"nr_queues"`
}
