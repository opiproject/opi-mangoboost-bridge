// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2025 MangoBoost, Inc.

package frontend

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/opiproject/gospdk/spdk"
	pb "github.com/opiproject/opi-api/storage/v1alpha1/gen/go"
	"github.com/opiproject/opi-spdk-bridge/pkg/frontend"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	testTraddr = "192.168.1.10"
	testNqn    = "nqn.2022-09.io.spdk:opi0"
	trtypeRdma = "RDMA"

	rpcGetTransports   = "nvmf_get_transports"
	rpcCreateTransport = "nvmf_create_transport"
	rpcGetListeners    = "nvmf_subsystem_get_listeners"
	rpcAddListener     = "nvmf_subsystem_add_listener"

	keyTrtype  = "trtype"
	keyTraddr  = "traddr"
	keyTrsvcid = "trsvcid"
	keyAdrfam  = "adrfam"
)

type rpcCall struct {
	method string
	params map[string]any
}

// stubRPC records calls and answers each method with a canned result
type stubRPC struct {
	spdk.JSONRPC
	results map[string]any
	calls   []rpcCall
}

func (s *stubRPC) Call(_ context.Context, method string, params, result any) error {
	call := rpcCall{method: method}
	if params != nil {
		data, err := json.Marshal(params)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(data, &call.params); err != nil {
			return err
		}
	}
	s.calls = append(s.calls, call)

	data, err := json.Marshal(s.results[method])
	if err != nil {
		return err
	}
	return json.Unmarshal(data, result)
}

func (s *stubRPC) methods() []string {
	methods := make([]string, 0, len(s.calls))
	for _, c := range s.calls {
		methods = append(methods, c.method)
	}
	return methods
}

func rdmaController() *pb.NvmeController {
	return &pb.NvmeController{
		Name: "nvmeSubsystems/subsys0/nvmeControllers/ctrl0",
		Spec: &pb.NvmeControllerSpec{
			Trtype: pb.NvmeTransportType_NVME_TRANSPORT_TYPE_RDMA,
			Endpoint: &pb.NvmeControllerSpec_FabricsId{
				FabricsId: &pb.FabricsEndpoint{
					Traddr:  testTraddr,
					Trsvcid: "4420",
					Adrfam:  pb.NvmeAddressFamily_NVME_ADDRESS_FAMILY_IPV4,
				},
			},
		},
	}
}

func TestRdmaTransportCreateController(t *testing.T) {
	rpc := &stubRPC{results: map[string]any{rpcAddListener: true}}
	subsys := &pb.NvmeSubsystem{Spec: &pb.NvmeSubsystemSpec{Nqn: testNqn}}

	if err := NewNvmeRdmaTransport(rpc).CreateController(context.Background(), rdmaController(), subsys); err != nil {
		t.Fatalf("CreateController: %v", err)
	}

	want := []rpcCall{{
		method: rpcAddListener,
		params: map[string]any{
			"nqn": testNqn,
			"listen_address": map[string]any{
				keyTrtype:  "rdma",
				keyTraddr:  testTraddr,
				keyTrsvcid: "4420",
				keyAdrfam:  "IPV4",
			},
		},
	}}
	if !reflect.DeepEqual(rpc.calls, want) {
		t.Errorf("calls = %v, want %v", rpc.calls, want)
	}
}

func TestRdmaTransportDeleteController(t *testing.T) {
	rpc := &stubRPC{results: map[string]any{"nvmf_subsystem_remove_listener": true}}
	subsys := &pb.NvmeSubsystem{Spec: &pb.NvmeSubsystemSpec{Nqn: testNqn}}

	if err := NewNvmeRdmaTransport(rpc).DeleteController(context.Background(), rdmaController(), subsys); err != nil {
		t.Fatalf("DeleteController: %v", err)
	}
	if got := rpc.methods(); !reflect.DeepEqual(got, []string{"nvmf_subsystem_remove_listener"}) {
		t.Errorf("methods = %v", got)
	}
}

func TestRdmaTransportRejectsPsk(t *testing.T) {
	rpc := &stubRPC{}
	subsys := &pb.NvmeSubsystem{Spec: &pb.NvmeSubsystemSpec{
		Nqn: testNqn,
		Psk: []byte("NVMeTLSkey-1:01:MDAxMTIyMzM0NDU1NjY3Nzg4OTlhYWJiY2NkZGVlZmZwJEiQ:"),
	}}

	err := NewNvmeRdmaTransport(rpc).CreateController(context.Background(), rdmaController(), subsys)
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("err = %v, want InvalidArgument", err)
	}
	if len(rpc.calls) != 0 {
		t.Errorf("unexpected calls %v", rpc.methods())
	}
}

func TestCreateNvmeSubsystemCntlidRange(t *testing.T) {
	tests := map[string]struct {
		min, max int
		want     map[string]any
	}{
		"spdk default": {
			want: map[string]any{},
		},
		"custom range": {
			min: 1, max: 999,
			want: map[string]any{"min_cntlid": float64(1), "max_cntlid": float64(999)},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			rpc := &stubRPC{results: map[string]any{
				"nvmf_create_subsystem": true,
				"spdk_get_version":      map[string]any{"version": "SPDK v24.01"},
			}}
			s := &Server{
				Nvme:      &frontend.NvmeParameters{Subsystems: map[string]*pb.NvmeSubsystem{}},
				rpc:       rpc,
				MinCntlid: tt.min,
				MaxCntlid: tt.max,
			}
			_, err := s.CreateNvmeSubsystem(context.Background(), &pb.CreateNvmeSubsystemRequest{
				NvmeSubsystemId: "subsys0",
				NvmeSubsystem: &pb.NvmeSubsystem{Spec: &pb.NvmeSubsystemSpec{
					Nqn:          testNqn,
					SerialNumber: "SN0001",
					ModelNumber:  "BoostX NRT",
				}},
			})
			if err != nil {
				t.Fatalf("CreateNvmeSubsystem: %v", err)
			}

			params := rpc.calls[0].params
			got := map[string]any{}
			for _, key := range []string{"min_cntlid", "max_cntlid"} {
				if v, ok := params[key]; ok {
					got[key] = v
				}
			}
			if rpc.calls[0].method != "nvmf_create_subsystem" || !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s params cntlid = %v, want %v", rpc.calls[0].method, got, tt.want)
			}
		})
	}
}

func TestSetupRdmaTarget(t *testing.T) {
	rdmaListener := map[string]any{"address": map[string]any{
		keyTrtype: trtypeRdma, keyAdrfam: "IPv4", keyTraddr: testTraddr, keyTrsvcid: "8009",
	}}
	tests := map[string]struct {
		cfg     RdmaTargetConfig
		results map[string]any
		want    []string
	}{
		"fresh target": {
			cfg: RdmaTargetConfig{Offload: true, DiscoveryAddr: "192.168.1.10:8009"},
			results: map[string]any{
				rpcGetTransports:   []any{},
				rpcCreateTransport: true,
				rpcGetListeners:    []any{},
				rpcAddListener:     true,
			},
			want: []string{rpcGetTransports, rpcCreateTransport,
				rpcGetListeners, rpcAddListener},
		},
		"configured target": {
			cfg: RdmaTargetConfig{Offload: true, DiscoveryAddr: "192.168.1.10:8009"},
			results: map[string]any{
				rpcGetTransports: []any{map[string]any{keyTrtype: trtypeRdma, "enable_offload": true}},
				rpcGetListeners:  []any{rdmaListener},
			},
			want: []string{rpcGetTransports, rpcGetListeners},
		},
		"no discovery": {
			cfg: RdmaTargetConfig{Offload: true},
			results: map[string]any{
				rpcGetTransports:   []any{map[string]any{keyTrtype: "TCP"}},
				rpcCreateTransport: true,
			},
			want: []string{rpcGetTransports, rpcCreateTransport},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			rpc := &stubRPC{results: tt.results}
			if err := SetupRdmaTarget(context.Background(), rpc, tt.cfg); err != nil {
				t.Fatalf("SetupRdmaTarget: %v", err)
			}
			if got := rpc.methods(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("methods = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetupRdmaTargetParams(t *testing.T) {
	rpc := &stubRPC{results: map[string]any{
		rpcGetTransports:   []any{},
		rpcCreateTransport: true,
		rpcGetListeners:    []any{},
		rpcAddListener:     true,
	}}
	cfg := RdmaTargetConfig{Offload: true, DiscoveryAddr: "[fd00::10]:8009"}
	if err := SetupRdmaTarget(context.Background(), rpc, cfg); err != nil {
		t.Fatalf("SetupRdmaTarget: %v", err)
	}

	wantTransport := map[string]any{keyTrtype: trtypeRdma, "enable_offload": true}
	if got := rpc.calls[1].params; !reflect.DeepEqual(got, wantTransport) {
		t.Errorf("nvmf_create_transport params = %v, want %v", got, wantTransport)
	}
	wantListener := map[string]any{
		"nqn": discoveryNqn,
		"listen_address": map[string]any{
			keyTrtype: "rdma", keyTraddr: "fd00::10", keyTrsvcid: "8009", keyAdrfam: "ipv6",
		},
	}
	if got := rpc.calls[3].params; !reflect.DeepEqual(got, wantListener) {
		t.Errorf("nvmf_subsystem_add_listener params = %v, want %v", got, wantListener)
	}
}

func TestSetupRdmaTargetInvalidDiscoveryAddr(t *testing.T) {
	rpc := &stubRPC{results: map[string]any{
		rpcGetTransports: []any{map[string]any{keyTrtype: trtypeRdma}},
	}}
	if err := SetupRdmaTarget(context.Background(), rpc, RdmaTargetConfig{DiscoveryAddr: "target:8009"}); err == nil {
		t.Error("expected an error for a hostname")
	}
}
