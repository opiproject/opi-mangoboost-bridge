// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022-2023 Dell Inc, or its subsidiaries.
// Copyright (C) 2023 Intel Corporation
// Copyright (C) 2025 MangoBoost, Inc.

// Package frontend implements the FrontEnd APIs (host facing) of the storage Server
package frontend

import (
	"context"
	"fmt"
	"log"

	"github.com/opiproject/gospdk/spdk"
	pb "github.com/opiproject/opi-api/storage/v1alpha1/gen/go"
	"github.com/opiproject/opi-mangoboost-bridge/pkg/models"
	"github.com/opiproject/opi-spdk-bridge/pkg/utils"

	"go.einride.tech/aip/resourceid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateNvmeNamespace creates a new NVMe namespace with the specified parameters
func (s *Server) CreateNvmeNamespace(ctx context.Context, in *pb.CreateNvmeNamespaceRequest) (*pb.NvmeNamespace, error) {
	// check input correctness
	if err := s.validateCreateNvmeNamespaceRequest(in); err != nil {
		return nil, err
	}

	resourceID := resourceid.NewSystemGenerated()
	if in.NvmeNamespaceId != "" {
		log.Printf("client provided the ID of a resource %v, ignoring the name field %v", in.NvmeNamespaceId, in.NvmeNamespace.Name)
		resourceID = in.NvmeNamespaceId
	}
	in.NvmeNamespace.Name = utils.ResourceIDToNamespaceName(utils.GetSubsystemIDFromNvmeName(in.Parent), resourceID)
	namespace, ok := s.Nvme.Namespaces[in.NvmeNamespace.Name]
	if ok {
		log.Printf("Already existing NvmeNamespace with id %v", in.NvmeNamespace.Name)
		return namespace, nil
	}
	subsys, ok := s.Nvme.Subsystems[in.Parent]
	if !ok {
		err := fmt.Errorf("unable to find subsystem %s", in.Parent)
		return nil, err
	}

	// Check if the subsystem has NvmeController, that is, if the subsystem is for NTI
	// if so, Pending should be set to false
	var pending = true

	Blobarray := []*pb.NvmeController{}
	for _, controller := range s.Nvme.Controllers {
		Blobarray = append(Blobarray, controller)
	}
	if len(Blobarray) > 0 {
		pending = false
		log.Printf("Found an NvmeController for namespace %s, setting Pending to false", in.NvmeNamespace.Name)
	} else {
		log.Printf("No NvmeController found for namespace %s, setting Pending to true", in.NvmeNamespace.Name)
	}

	params := models.NvmfSubsystemAddNsParams{
		Nqn: subsys.Spec.Nqn,
		Namespace: struct {
			Nsid     int    `json:"nsid,omitempty"`
			BdevName string `json:"bdev_name"`
		}{
			Nsid:     int(in.NvmeNamespace.Spec.HostNsid),
			BdevName: in.NvmeNamespace.Spec.VolumeNameRef,
		},
		Pending: pending,
	}

	var result spdk.NvmfSubsystemAddNsResult
	err := s.rpc.Call(ctx, "nvmf_subsystem_add_ns", &params, &result)
	if err != nil {
		return nil, err
	}
	log.Printf("Received from SPDK: %v", result)
	if result < 0 {
		msg := fmt.Sprintf("Could not create NS: %s", in.NvmeNamespace.Name)
		return nil, status.Errorf(codes.InvalidArgument, "%s", msg)
	}

	response := utils.ProtoClone(in.NvmeNamespace)
	response.Status = &pb.NvmeNamespaceStatus{
		State:     pb.NvmeNamespaceStatus_STATE_ENABLED,
		OperState: pb.NvmeNamespaceStatus_OPER_STATE_ONLINE,
	}
	response.Spec.HostNsid = int32(result)
	s.Nvme.Namespaces[in.NvmeNamespace.Name] = response
	return response, nil
}
