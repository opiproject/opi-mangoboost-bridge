// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2023 Intel Corporation
// Copyright (C) 2025 MangoBoost, Inc.

// Package models holds definitions for SPDK json RPC structs
package models

import "github.com/opiproject/gospdk/spdk"

// NvmfSubsystemAddNsParams holds parameters for adding a namespace to an NVMf subsystem
type NvmfSubsystemAddNsParams struct {
	Nqn       string `json:"nqn"`
	Namespace struct {
		// We can leave the Nsid empty, as it will be assigned by SPDK
		Nsid     int    `json:"nsid,omitempty"`
		BdevName string `json:"bdev_name"`
	} `json:"namespace"`
}

// BdevNvmeAttachControllerParams holds parameters for attaching an NVMe controller
type BdevNvmeAttachControllerParams struct {
	Name    string `json:"name"`
	Trtype  string `json:"trtype"`
	Traddr  string `json:"traddr"`
	Hostnqn string `json:"hostnqn,omitempty"`
	Adrfam  string `json:"adrfam,omitempty"`
	Trsvcid string `json:"trsvcid,omitempty"`
	Subnqn  string `json:"subnqn,omitempty"`

	// We don't use "spdk.BdevNvmeAttachControllerParams" because of these fields
	// Use pointers to explicitly pass "false" for optional fields
	Hdgst *bool `json:"hdgst,omitempty"`
	Ddgst *bool `json:"ddgst,omitempty"`

	Psk       string `json:"psk,omitempty"`
	Multipath string `json:"multipath,omitempty"`

	// Custom params
	AdminQueueSize int  `json:"admin_queue_size,omitempty"`
	IoQueueSize    int  `json:"io_queue_size,omitempty"`
	NumIoQueues    int  `json:"num_io_queues,omitempty"`
	SkipExamine    bool `json:"skip_examine,omitempty"`
}

// NvmfCreateSubsystemParams holds parameters for creating an NVMf subsystem
type NvmfCreateSubsystemParams struct {
	spdk.NvmfCreateSubsystemParams

	// Custom params
	MultiPath bool `json:"multipath"`
	MinCntlid int  `json:"min_cntlid,omitempty"`
	MaxCntlid int  `json:"max_cntlid,omitempty"`
}

// NvmfCreateTransportParams holds parameters for creating an NVMf transport
type NvmfCreateTransportParams struct {
	Trtype string `json:"trtype"`

	// Custom params: enables BoostX NRT hardware offload on the RDMA transport
	EnableOffload bool `json:"enable_offload,omitempty"`
}

// NvmfTransport is an entry of the nvmf_get_transports result
type NvmfTransport struct {
	Trtype        string `json:"trtype"`
	EnableOffload bool   `json:"enable_offload"`
}

// NvmfListener is an entry of the nvmf_subsystem_get_listeners result
type NvmfListener struct {
	Address struct {
		Trtype  string `json:"trtype"`
		Adrfam  string `json:"adrfam"`
		Traddr  string `json:"traddr"`
		Trsvcid string `json:"trsvcid"`
	} `json:"address"`
}
