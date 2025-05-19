// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022-2024 Dell Inc, or its subsidiaries.
// Copyright (C) 2023 Intel Corporation
// Copyright (C) 2025 MangoBoost, Inc.

// Package backend implements the BackEnd APIs (network facing) of the storage Server
package backend

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/opiproject/gospdk/spdk"
	pb "github.com/opiproject/opi-api/storage/v1alpha1/gen/go"
	"github.com/opiproject/opi-mangoboost-bridge/pkg/models"
	"github.com/opiproject/opi-spdk-bridge/pkg/utils"

	"go.einride.tech/aip/resourceid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	defaultAdminQueueSize = 32
	defaultIoQueueSize    = 128
	defaultNumIoQueues    = 128
)

func (s *Server) CreateNvmePath(ctx context.Context, in *pb.CreateNvmePathRequest) (*pb.NvmePath, error) {
	// check input correctness
	if err := s.validateCreateNvmePathRequest(in); err != nil {
		return nil, err
	}

	// see https://google.aip.dev/133#user-specified-ids
	resourceID := resourceid.NewSystemGenerated()
	if in.NvmePathId != "" {
		log.Printf("client provided the ID of a resource %v, ignoring the name field %v", in.NvmePathId, in.NvmePath.Name)
		resourceID = in.NvmePathId
	}
	in.NvmePath.Name = utils.ResourceIDToNvmePathName(
		utils.GetRemoteControllerIDFromNvmeRemoteName(in.Parent),
		resourceID,
	)

	nvmePath, ok := s.Volumes.NvmePaths[in.NvmePath.Name]
	if ok {
		log.Printf("Already existing NvmePath with id %v", in.NvmePath.Name)
		return nvmePath, nil
	}

	controller, ok := s.Volumes.NvmeControllers[in.Parent]
	if !ok {
		err := status.Errorf(codes.NotFound, "unable to find NvmeRemoteController by key %s", in.Parent)
		return nil, err
	}

	if in.NvmePath.Trtype == pb.NvmeTransportType_NVME_TRANSPORT_TYPE_PCIE && controller.Tcp != nil {
		err := status.Errorf(codes.FailedPrecondition, "pcie transport on tcp controller is not allowed")
		return nil, err
	}

	multipath := ""
	if numberOfPaths := s.numberOfPathsForController(controller.Name); numberOfPaths > 0 {
		// set multipath parameter only when at least one path already exists
		multipath = s.opiMultipathToSpdk(controller.Multipath)
	}
	psk := ""
	if len(controller.GetTcp().GetPsk()) > 0 {
		log.Printf("Notice, TLS is used to establish connection: to %v", in.NvmePath)
		keyFile, err := s.keyToTemporaryFile(controller.Tcp.Psk)
		if err != nil {
			return nil, err
		}
		defer func() {
			err := os.Remove(keyFile)
			log.Printf("Cleanup key file %v: %v", keyFile, err)
		}()

		psk = keyFile
	}

	// Set skip_examine to false if trttype is PCIe,
	// otherwise set it to true
	skipExamine := true
	if in.NvmePath.Trtype == pb.NvmeTransportType_NVME_TRANSPORT_TYPE_PCIE {
		skipExamine = false
	}

	params := models.BdevNvmeAttachControllerParams{
		BdevNvmeAttachControllerParams: spdk.BdevNvmeAttachControllerParams{
			Name:      utils.GetRemoteControllerIDFromNvmeRemoteName(controller.Name),
			Trtype:    s.opiTransportToSpdk(in.NvmePath.GetTrtype()),
			Traddr:    in.NvmePath.GetTraddr(),
			Adrfam:    utils.OpiAdressFamilyToSpdk(in.NvmePath.GetFabrics().GetAdrfam()),
			Trsvcid:   fmt.Sprint(in.NvmePath.GetFabrics().GetTrsvcid()),
			Subnqn:    in.NvmePath.GetFabrics().GetSubnqn(),
			Hostnqn:   in.NvmePath.GetFabrics().GetHostnqn(),
			Multipath: multipath,
			Hdgst:     controller.GetTcp().GetHdgst(),
			Ddgst:     controller.GetTcp().GetDdgst(),
			Psk:       psk,
		},
		AdminQueueSize: defaultAdminQueueSize,
		IoQueueSize:    defaultIoQueueSize,
		NumIoQueues:    defaultNumIoQueues,
		SkipExamine:    skipExamine,
	}

	var result []spdk.BdevNvmeAttachControllerResult
	err := s.rpc.Call(ctx, "bdev_nvme_attach_controller", &params, &result)
	if err != nil {
		return nil, err
	}
	log.Printf("Received from SPDK: %v", result)

	response := utils.ProtoClone(in.NvmePath)
	s.Volumes.NvmePaths[in.NvmePath.Name] = response
	return response, nil
}

func (s *Server) numberOfPathsForController(controllerName string) int {
	numberOfPaths := 0
	prefix := controllerName + "/"
	for _, path := range s.Volumes.NvmePaths {
		if strings.HasPrefix(path.Name, prefix) {
			numberOfPaths++
		}
	}
	return numberOfPaths
}

func (s *Server) opiTransportToSpdk(transport pb.NvmeTransportType) string {
	return strings.ReplaceAll(transport.String(), "NVME_TRANSPORT_TYPE_", "")
}

func (s *Server) opiMultipathToSpdk(multipath pb.NvmeMultipath) string {
	return strings.ToLower(
		strings.ReplaceAll(multipath.String(), "NVME_MULTIPATH_", ""),
	)
}
