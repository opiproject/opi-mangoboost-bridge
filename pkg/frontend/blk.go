// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2023 Intel Corporation
// Copyright (C) 2025 MangoBoost, Inc.

// Package frontend implements the FrontEnd APIs (host facing) of the storage Server
package frontend

import (
	pb "github.com/opiproject/opi-api/storage/v1alpha1/gen/go"
	"github.com/opiproject/opi-spdk-bridge/pkg/frontend"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Virtio Host Interface (VHI): NOT YET IMPLEMENTED
type virtioVhiTransport struct{}

func (v virtioVhiTransport) CreateParams(_ *pb.VirtioBlk) (any, error) {
	return nil, status.Errorf(codes.Unimplemented, "CreateParams method is not implemented")
}

func (v virtioVhiTransport) DeleteParams(_ *pb.VirtioBlk) (any, error) {
	return nil, status.Errorf(codes.Unimplemented, "DeleteParams method is not implemented")
}

// NewVirtioVhiTransport creates an isntance of virtioVhiTransport
func NewVirtioVhiTransport() frontend.VirtioBlkTransport {
	return &virtioVhiTransport{}
}
