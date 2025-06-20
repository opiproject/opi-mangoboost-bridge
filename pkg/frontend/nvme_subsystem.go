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
	"os"

	"github.com/opiproject/gospdk/spdk"
	pb "github.com/opiproject/opi-api/storage/v1alpha1/gen/go"
	"github.com/opiproject/opi-mangoboost-bridge/pkg/models"
	"github.com/opiproject/opi-spdk-bridge/pkg/utils"

	"go.einride.tech/aip/resourceid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateNvmeSubsystem creates a new NVMe subsystem with the specified parameters
func (s *Server) CreateNvmeSubsystem(ctx context.Context, in *pb.CreateNvmeSubsystemRequest) (*pb.NvmeSubsystem, error) {
	// check input correctness
	if err := s.validateCreateNvmeSubsystemRequest(in); err != nil {
		return nil, err
	}
	// see https://google.aip.dev/133#user-specified-ids
	resourceID := resourceid.NewSystemGenerated()
	if in.NvmeSubsystemId != "" {
		log.Printf("client provided the ID of a resource %v, ignoring the name field %v", in.NvmeSubsystemId, in.NvmeSubsystem.Name)
		resourceID = in.NvmeSubsystemId
	}
	in.NvmeSubsystem.Name = utils.ResourceIDToSubsystemName(resourceID)
	// idempotent API when called with same key, should return same object
	subsys, ok := s.Nvme.Subsystems[in.NvmeSubsystem.Name]
	if ok {
		log.Printf("Already existing NvmeSubsystem with id %v", in.NvmeSubsystem.Name)
		return subsys, nil
	}
	// check if another object exists with same NQN, it is not allowed
	for _, item := range s.Nvme.Subsystems {
		if in.NvmeSubsystem.Spec.Nqn == item.Spec.Nqn {
			msg := fmt.Sprintf("Could not create NQN: %s since object %s with same NQN already exists", in.NvmeSubsystem.Spec.Nqn, item.Name)
			return nil, status.Errorf(codes.AlreadyExists, "%s", msg)
		}
	}
	// not found, so create a new one
	params := models.NvmfCreateSubsystemParams{
		NvmfCreateSubsystemParams: spdk.NvmfCreateSubsystemParams{

			Nqn:           in.NvmeSubsystem.Spec.Nqn,
			SerialNumber:  in.NvmeSubsystem.Spec.SerialNumber,
			ModelNumber:   in.NvmeSubsystem.Spec.ModelNumber,
			AllowAnyHost:  (in.NvmeSubsystem.Spec.Hostnqn == ""),
			MaxNamespaces: int(in.NvmeSubsystem.Spec.MaxNamespaces),
		},
		// TODO: "multipath" parameter is currently hardcoded to false. It should be able to take it as an input.
		MultiPath: false,
	}

	var result spdk.NvmfCreateSubsystemResult
	err := s.rpc.Call(ctx, "nvmf_create_subsystem", &params, &result)
	if err != nil {
		return nil, err
	}
	log.Printf("Received from SPDK: %v", result)
	if !result {
		msg := fmt.Sprintf("Could not create NQN: %s", in.NvmeSubsystem.Spec.Nqn)
		return nil, status.Errorf(codes.InvalidArgument, "%s", msg)
	}
	// if Hostnqn is not empty, add it to subsystem
	if in.NvmeSubsystem.Spec.Hostnqn != "" {
		psk := ""
		if len(in.NvmeSubsystem.Spec.Psk) > 0 {
			log.Printf("Notice, TLS is used for subsystem %v", in.NvmeSubsystem.Name)
			keyFile, err := s.keyToTemporaryFile(in.NvmeSubsystem.Spec.Psk)
			if err != nil {
				return nil, err
			}
			defer func() {
				err := os.Remove(keyFile)
				log.Printf("Cleanup key file %v: %v", keyFile, err)
			}()

			psk = keyFile
		}
		params := spdk.NvmfSubsystemAddHostParams{
			Nqn:  in.NvmeSubsystem.Spec.Nqn,
			Host: in.NvmeSubsystem.Spec.Hostnqn,
			Psk:  psk,
		}
		var result spdk.NvmfSubsystemAddHostResult
		err = s.rpc.Call(ctx, "nvmf_subsystem_add_host", &params, &result)
		if err != nil {
			return nil, err
		}
		log.Printf("Received from SPDK: %v", result)
		if !result {
			msg := fmt.Sprintf("Could not add Hostnqn %s to NQN: %s", in.NvmeSubsystem.Spec.Hostnqn, in.NvmeSubsystem.Spec.Nqn)
			return nil, status.Errorf(codes.InvalidArgument, "%s", msg)
		}
	}
	// get SPDK version
	var ver spdk.GetVersionResult
	err = s.rpc.Call(ctx, "spdk_get_version", nil, &ver)
	if err != nil {
		return nil, err
	}
	log.Printf("Received from SPDK: %v", ver)
	response := utils.ProtoClone(in.NvmeSubsystem)
	response.Status = &pb.NvmeSubsystemStatus{FirmwareRevision: ver.Version}
	s.Nvme.Subsystems[in.NvmeSubsystem.Name] = response
	return response, nil
}
