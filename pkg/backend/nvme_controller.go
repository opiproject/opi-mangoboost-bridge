// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022-2024 Dell Inc, or its subsidiaries.
// Copyright (C) 2023 Intel Corporation
// Copyright (C) 2025 MangoBoost, Inc.

// Package backend implements the BackEnd APIs (network facing) of the storage Server
package backend

import (
	"context"
	"log"

	pb "github.com/opiproject/opi-api/storage/v1alpha1/gen/go"
	"github.com/opiproject/opi-spdk-bridge/pkg/utils"

	"go.einride.tech/aip/resourceid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) CreateNvmeRemoteController(_ context.Context, in *pb.CreateNvmeRemoteControllerRequest) (*pb.NvmeRemoteController, error) {
	// check input correctness
	if err := s.validateCreateNvmeRemoteControllerRequest(in); err != nil {
		return nil, err
	}

	// see https://google.aip.dev/133#user-specified-ids
	resourceID := resourceid.NewSystemGenerated()
	if in.NvmeRemoteControllerId != "" {
		log.Printf("client provided the ID of a resource %v, ignoring the name field %v", in.NvmeRemoteControllerId, in.NvmeRemoteController.Name)
		resourceID = in.NvmeRemoteControllerId
	}
	in.NvmeRemoteController.Name = utils.ResourceIDToRemoteControllerName(resourceID)
	// idempotent API when called with same key, should return same object
	volume, ok := s.Volumes.NvmeControllers[in.NvmeRemoteController.Name]
	if ok {
		log.Printf("Already existing NvmeRemoteController with id %v", in.NvmeRemoteController.Name)
		return volume, nil
	}
	// not found, so create a new one
	response := utils.ProtoClone(in.NvmeRemoteController)
	s.Volumes.NvmeControllers[in.NvmeRemoteController.Name] = response
	return response, nil
}

func (s *Server) DeleteNvmeRemoteController(_ context.Context, in *pb.DeleteNvmeRemoteControllerRequest) (*emptypb.Empty, error) {
	// check input correctness
	if err := s.validateDeleteNvmeRemoteControllerRequest(in); err != nil {
		return nil, err
	}

	volume, ok := s.Volumes.NvmeControllers[in.Name]
	if !ok {
		if in.AllowMissing {
			return &emptypb.Empty{}, nil
		}
		err := status.Errorf(codes.NotFound, "unable to find key %s", in.Name)
		return nil, err
	}
	if s.numberOfPathsForController(in.Name) > 0 {
		return nil, status.Error(codes.FailedPrecondition, "NvmePaths exist for controller")
	}
	delete(s.Volumes.NvmeControllers, volume.Name)
	return &emptypb.Empty{}, nil
}
