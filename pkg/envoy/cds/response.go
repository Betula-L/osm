package cds

import (
	"fmt"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"

	v2_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"

	api "github.com/envoyproxy/go-control-plane/envoy/api/v2"

	"time"

	xds "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/gogo/protobuf/types"
	"github.com/golang/glog"

	"github.com/deislabs/smc/pkg/envoy"
)

const (
	sleepTime = 5
)

func (s *Server) newDiscoveryResponse(proxy envoy.Proxyer) (*xds.DiscoveryResponse, error) {
	glog.Infof("[%s] Composing SDS Discovery Response for proxy: %s", serverName, proxy.GetCommonName())
	resp := &xds.DiscoveryResponse{
		TypeUrl: discoveryResponseTypeUrl,
	}

	// The name must match the domain being cURLed in the demo
	clusterName := "bookstore.mesh"
	connTimeout := 5 * time.Second

	cluster := &xds.Cluster{
		// The name must match the domain being cURLed in the demo
		Name:           clusterName,
		AltStatName:    clusterName,
		ConnectTimeout: &connTimeout,
		LbPolicy:       xds.Cluster_ROUND_ROBIN,
		ClusterDiscoveryType: &xds.Cluster_Type{
			Type: xds.Cluster_EDS,
		},
		EdsClusterConfig: &api.Cluster_EdsClusterConfig{
			EdsConfig: &v2_core.ConfigSource{
				ConfigSourceSpecifier: &v2_core.ConfigSource_ApiConfigSource{
					ApiConfigSource: &v2_core.ApiConfigSource{
						ApiType: v2_core.ApiConfigSource_GRPC,
						GrpcServices: []*v2_core.GrpcService{
							{
								TargetSpecifier: &v2_core.GrpcService_EnvoyGrpc_{
									EnvoyGrpc: &v2_core.GrpcService_EnvoyGrpc{
										// This must match the hard-coded EDS cluster name in the bootstrap config
										ClusterName: "eds",
									},
								},
							},
						},
					},
				},
			},
		},

		TlsContext: &auth.UpstreamTlsContext{
			CommonTlsContext: &auth.CommonTlsContext{
				TlsParams:       nil,
				TlsCertificates: nil,
				TlsCertificateSdsSecretConfigs: []*auth.SdsSecretConfig{
					{
						// The Name field must match the auth.Secret.Name from the SDS response
						Name: "server_cert",
						SdsConfig: &v2_core.ConfigSource{
							ConfigSourceSpecifier: &v2_core.ConfigSource_ApiConfigSource{
								ApiConfigSource: &v2_core.ApiConfigSource{
									ApiType: v2_core.ApiConfigSource_GRPC,
									GrpcServices: []*v2_core.GrpcService{
										{
											TargetSpecifier: &v2_core.GrpcService_EnvoyGrpc_{
												EnvoyGrpc: &v2_core.GrpcService_EnvoyGrpc{
													// This must match the hard-coded SDS cluster name in the bootstrap config
													ClusterName: "sds",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	marshalledClusters, err := types.MarshalAny(cluster)
	if err != nil {
		glog.Errorf("[%s] Failed to marshal cluster for proxy %s: %v", serverName, proxy.GetCommonName(), err)
		return nil, err
	}
	resp.Resources = append(resp.Resources, marshalledClusters)

	s.lastVersion = s.lastVersion + 1
	s.lastNonce = string(time.Now().Nanosecond())
	resp.Nonce = s.lastNonce
	resp.VersionInfo = fmt.Sprintf("v%d", s.lastVersion)

	glog.V(7).Infof("[%s] Constructed response: %+v", serverName, resp)

	return resp, nil
}
