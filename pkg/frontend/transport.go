// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2023 Intel Corporation
// Copyright (C) 2025 MangoBoost, Inc.

// Package frontend implements the FrontEnd APIs (host facing) of the storage Server
package frontend

import (
	"context"
	"log"
	"strconv"

	"github.com/opiproject/gospdk/spdk"
	pb "github.com/opiproject/opi-api/storage/v1alpha1/gen/go"
	"github.com/opiproject/opi-mangoboost-bridge/pkg/models"
	"github.com/opiproject/opi-spdk-bridge/pkg/frontend"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Nvme Host Interface (NHI)
type nvmeNhiTransport struct {
	rpc spdk.JSONRPC
}

// build time check that struct implements interface
var _ frontend.NvmeTransport = (*nvmeNhiTransport)(nil)

// NewNvmeNhiTransport creates a new instance of a NvmeTransport for nhi
func NewNvmeNhiTransport(rpc spdk.JSONRPC) frontend.NvmeTransport {
	if rpc == nil {
		log.Panicf("rpc cannot be nil")
	}

	return &nvmeNhiTransport{
		rpc: rpc,
	}
}

func (c *nvmeNhiTransport) CreateController(
	ctx context.Context,
	ctrlr *pb.NvmeController,
	subsys *pb.NvmeSubsystem,
) error {
	if ctrlr.GetSpec().GetPcieId().GetPortId().GetValue() != 0 {
		return status.Error(codes.InvalidArgument, "only port 0 is supported")
	}

	if ctrlr.GetSpec().GetPcieId().GetPhysicalFunction().GetValue() != 0 {
		return status.Error(codes.InvalidArgument,
			"only physical_function 0 is supported")
	}

	if subsys.GetSpec().GetHostnqn() != "" {
		return status.Error(codes.InvalidArgument,
			"hostnqn for subsystem is not supported for nhi")
	}

	maxNsq := ctrlr.GetSpec().GetMaxNsq()
	maxNcq := ctrlr.GetSpec().GetMaxNcq()
	if maxNsq != maxNcq {
		return status.Error(codes.InvalidArgument,
			"max_nsq and max_ncq must be equal")
	}

	params := c.params(ctrlr, subsys)
	if maxNsq > 0 {
		params.NumQueues = int(maxNsq) + 1 // + 1 admin queue
	}
	var result spdk.NvmfSubsystemAddListenerResult
	err := c.rpc.Call(ctx, "nvmf_subsystem_add_listener", &params, &result)
	if err != nil {
		return status.Error(codes.Unknown, err.Error())
	}
	log.Printf("Received from SPDK: %v", result)
	if !result {
		return status.Errorf(codes.InvalidArgument,
			"Could not create CTRL: %s", ctrlr.Name)
	}

	return nil
}

func (c *nvmeNhiTransport) DeleteController(
	ctx context.Context,
	ctrlr *pb.NvmeController,
	subsys *pb.NvmeSubsystem,
) error {
	params := c.params(ctrlr, subsys)
	var result spdk.NvmfSubsystemAddListenerResult
	err := c.rpc.Call(ctx, "nvmf_subsystem_remove_listener", &params, &result)
	if err != nil {
		return err
	}
	log.Printf("Received from SPDK: %v", result)
	if !result {
		return status.Errorf(codes.InvalidArgument,
			"Could not delete CTRL: %s", ctrlr.Name)
	}

	return nil
}

func (c *nvmeNhiTransport) params(
	ctrlr *pb.NvmeController,
	subsys *pb.NvmeSubsystem,
) models.NhiNvmfSubsystemAddListenerParams {
	result := models.NhiNvmfSubsystemAddListenerParams{}
	result.Nqn = subsys.GetSpec().GetNqn()
	result.ListenAddress.Trtype = "NHI"
	result.ListenAddress.Traddr = strconv.Itoa(int(ctrlr.GetSpec().GetPcieId().GetPhysicalFunction().GetValue()))
	result.ListenAddress.Trsvcid = strconv.Itoa(int(ctrlr.GetSpec().GetPcieId().GetPhysicalFunction().GetValue()))

	return result
}
