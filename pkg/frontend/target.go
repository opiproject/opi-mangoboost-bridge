// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2025 MangoBoost, Inc.

// Package frontend implements the FrontEnd APIs (host facing) of the storage Server
package frontend

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/opiproject/gospdk/spdk"
	"github.com/opiproject/opi-mangoboost-bridge/pkg/models"
)

const discoveryNqn = "nqn.2014-08.org.nvmexpress.discovery"

// RdmaTargetConfig describes the target-wide RDMA setup of the NRT target
type RdmaTargetConfig struct {
	// Offload enables BoostX NRT hardware offload on the RDMA transport.
	Offload bool
	// DiscoveryAddr is an ip:port discovery listener; empty skips it.
	DiscoveryAddr string
}

// SetupRdmaTarget creates the RDMA transport and the discovery listener.
// Both steps are skipped when SPDK already has them, so a restarted bridge
// reattaches to a running target.
func SetupRdmaTarget(ctx context.Context, rpc spdk.JSONRPC, cfg RdmaTargetConfig) error {
	if err := ensureRdmaTransport(ctx, rpc, cfg.Offload); err != nil {
		return err
	}
	if cfg.DiscoveryAddr == "" {
		return nil
	}
	return ensureDiscoveryListener(ctx, rpc, cfg.DiscoveryAddr)
}

func ensureRdmaTransport(ctx context.Context, rpc spdk.JSONRPC, offload bool) error {
	var transports []models.NvmfTransport
	if err := rpc.Call(ctx, "nvmf_get_transports", nil, &transports); err != nil {
		return fmt.Errorf("nvmf_get_transports: %w", err)
	}
	for _, t := range transports {
		if strings.EqualFold(t.Trtype, "rdma") {
			if t.EnableOffload != offload {
				log.Printf("RDMA transport already exists with enable_offload=%v, requested %v",
					t.EnableOffload, offload)
			}
			return nil
		}
	}

	params := models.NvmfCreateTransportParams{Trtype: "RDMA", EnableOffload: offload}
	var result bool
	if err := rpc.Call(ctx, "nvmf_create_transport", &params, &result); err != nil {
		return fmt.Errorf("nvmf_create_transport: %w", err)
	}
	if !result {
		return fmt.Errorf("could not create RDMA transport")
	}
	log.Printf("Created RDMA transport with enable_offload=%v", offload)
	return nil
}

func ensureDiscoveryListener(ctx context.Context, rpc spdk.JSONRPC, addr string) error {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("invalid discovery address %q: %w", addr, err)
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return fmt.Errorf("invalid discovery address %q: not an IP address", addr)
	}
	adrfam := "ipv4"
	if ip.To4() == nil {
		adrfam = "ipv6"
	}

	var listeners []models.NvmfListener
	getParams := map[string]string{"nqn": discoveryNqn}
	if err := rpc.Call(ctx, "nvmf_subsystem_get_listeners", &getParams, &listeners); err != nil {
		return fmt.Errorf("nvmf_subsystem_get_listeners: %w", err)
	}
	for _, l := range listeners {
		if strings.EqualFold(l.Address.Trtype, "rdma") &&
			net.ParseIP(l.Address.Traddr).Equal(ip) && l.Address.Trsvcid == port {
			return nil
		}
	}

	params := spdk.NvmfSubsystemAddListenerParams{}
	params.Nqn = discoveryNqn
	params.ListenAddress.Trtype = "rdma"
	params.ListenAddress.Traddr = host
	params.ListenAddress.Trsvcid = port
	params.ListenAddress.Adrfam = adrfam
	var result spdk.NvmfSubsystemAddListenerResult
	if err := rpc.Call(ctx, "nvmf_subsystem_add_listener", &params, &result); err != nil {
		return fmt.Errorf("nvmf_subsystem_add_listener: %w", err)
	}
	if !result {
		return fmt.Errorf("could not add discovery listener %s", addr)
	}
	log.Printf("Added RDMA discovery listener %s", addr)
	return nil
}
