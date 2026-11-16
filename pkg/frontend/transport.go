// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2023 Intel Corporation
// Copyright (C) 2025 MangoBoost, Inc.

// Package frontend implements the FrontEnd APIs (host facing) of the storage Server
package frontend

import (
	"context"
	"log"

	"github.com/opiproject/gospdk/spdk"
	pb "github.com/opiproject/opi-api/storage/v1alpha1/gen/go"
	"github.com/opiproject/opi-spdk-bridge/pkg/frontend"
	"github.com/opiproject/opi-spdk-bridge/pkg/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NVMe-oF/RDMA listener, served by the BoostX NRT target
type nvmeRdmaTransport struct {
	rpc spdk.JSONRPC
}

// build time check that struct implements interface
var _ frontend.NvmeTransport = (*nvmeRdmaTransport)(nil)

// NewNvmeRdmaTransport creates a new instance of a NvmeTransport for rdma
func NewNvmeRdmaTransport(rpc spdk.JSONRPC) frontend.NvmeTransport {
	if rpc == nil {
		log.Panicf("rpc cannot be nil")
	}

	return &nvmeRdmaTransport{
		rpc: rpc,
	}
}

func (c *nvmeRdmaTransport) CreateController(
	ctx context.Context,
	ctrlr *pb.NvmeController,
	subsys *pb.NvmeSubsystem,
) error {
	if len(subsys.GetSpec().GetPsk()) > 0 {
		return status.Error(codes.InvalidArgument,
			"psk is not supported for rdma")
	}

	params := c.params(ctrlr, subsys)
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

func (c *nvmeRdmaTransport) DeleteController(
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

func (c *nvmeRdmaTransport) params(
	ctrlr *pb.NvmeController,
	subsys *pb.NvmeSubsystem,
) spdk.NvmfSubsystemAddListenerParams {
	fabricsID := ctrlr.GetSpec().GetFabricsId()

	result := spdk.NvmfSubsystemAddListenerParams{}
	result.Nqn = subsys.GetSpec().GetNqn()
	result.ListenAddress.Trtype = "rdma"
	result.ListenAddress.Traddr = fabricsID.GetTraddr()
	result.ListenAddress.Trsvcid = fabricsID.GetTrsvcid()
	result.ListenAddress.Adrfam = utils.OpiAdressFamilyToSpdk(fabricsID.GetAdrfam())

	return result
}
