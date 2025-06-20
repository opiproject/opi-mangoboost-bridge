// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022-2023 Dell Inc, or its subsidiaries.
// Copyright (C) 2023 Intel Corporation
// Copyright (C) 2025 MangoBoost, Inc.

// Package backend implements the BackEnd APIs (network facing) of the storage Server
package backend

import (
	"log"

	"github.com/philippgille/gokv"

	"github.com/opiproject/gospdk/spdk"
	pb "github.com/opiproject/opi-api/storage/v1alpha1/gen/go"
	"github.com/opiproject/opi-spdk-bridge/pkg/backend"
	"github.com/opiproject/opi-spdk-bridge/pkg/utils"
)

// Server represents the backend server implementation that handles NVMe remote controller operations
type Server struct {
	pb.NvmeRemoteControllerServiceServer

	rpc                spdk.JSONRPC
	store              gokv.Store
	Volumes            backend.VolumeParameters
	Pagination         map[string]int
	keyToTemporaryFile func(pskKey []byte) (string, error)
}

// NewServer creates and initializes a new Server instance with the provided JSON-RPC client and storage
func NewServer(jsonRPC spdk.JSONRPC, store gokv.Store) *Server {
	if jsonRPC == nil {
		log.Panic("nil for JSONRPC is not allowed")
	}
	if store == nil {
		log.Panic("nil for Store is not allowed")
	}

	volumes := backend.VolumeParameters{
		AioVolumes:      make(map[string]*pb.AioVolume),
		NullVolumes:     make(map[string]*pb.NullVolume),
		MallocVolumes:   make(map[string]*pb.MallocVolume),
		NvmeControllers: make(map[string]*pb.NvmeRemoteController),
		NvmePaths:       make(map[string]*pb.NvmePath),
	}
	pagination := make(map[string]int)

	opiSpdkServer := backend.NewServer(jsonRPC, store)
	opiSpdkServer.Volumes = volumes
	opiSpdkServer.Pagination = pagination

	return &Server{
		NvmeRemoteControllerServiceServer: opiSpdkServer,
		Volumes:                           volumes,
		Pagination:                        pagination,
		rpc:                               jsonRPC,
		store:                             store,
		keyToTemporaryFile:                utils.KeyToTemporaryFile,
	}
}
