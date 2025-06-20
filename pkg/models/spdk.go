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
