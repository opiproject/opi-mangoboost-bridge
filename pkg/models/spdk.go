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

type NvmfSubsystemAddNsParams struct {
	spdk.NvmfSubsystemAddNsParams

	// Custom params
	Pending bool `json:"pending"`
}

type BdevNvmeAttachControllerParams struct {
	spdk.BdevNvmeAttachControllerParams

	// Custom params
	AdminQueueSize int  `json:"admin_queue_size"`
	IoQueueSize    int  `json:"io_queue_size"`
	NumIoQueues    int  `json:"num_io_queues"`
	SkipExamine    bool `json:"skip_examine"`
}

type NvmfCreateSubsystemParams struct {
	spdk.NvmfCreateSubsystemParams

	// Custom params
	MultiPath bool `json:"multipath"`
}
