package manager

import (
	"testing"

	"github.com/kube-vip/kube-vip/pkg/kubevip"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestParseBgpAnnotations(t *testing.T) {
	type args struct {
		node       *corev1.Node
		baseConfig kubevip.BGPConfig
		prefix     string
	}
	tests := []struct {
		name           string
		args           args
		wantConfig     kubevip.BGPConfig
		wantPeerConfig kubevip.BGPPeer
		wantErr        bool
	}{
		{
			name: "missing annotations",
			args: args{
				node:   &corev1.Node{},
				prefix: "bgp",
			},
			wantErr: true,
		},
		{
			name: "minimum required annotations",
			args: args{
				node: &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"bgp/node-asn": "65000",
							"bgp/peer-asn": "64000",
							"bgp/src-ip":   "10.0.0.254",
						}},
				},
				prefix: "bgp",
			},
			wantConfig: kubevip.BGPConfig{
				AS:       65000,
				RouterID: "10.0.0.254",
				SourceIP: "10.0.0.254",
				Peers: []kubevip.BGPPeer{},
			},
			wantPeerConfig: kubevip.BGPPeer{
				AS: 64000,
			},
			wantErr: false,
		},
		{
			name: "base config not overwritten",
			args: args{
				node: &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"bgp/node-asn": "65000",
							"bgp/peer-asn": "64000",
							"bgp/src-ip":   "10.0.0.254",
						}},
				},
				prefix: "bgp",
				baseConfig: kubevip.BGPConfig{
					HoldTime:          15,
					KeepaliveInterval: 5,
				},
			},
			wantConfig: kubevip.BGPConfig{
				AS:                65000,
				RouterID:          "10.0.0.254",
				SourceIP:          "10.0.0.254",
				HoldTime:          15,
				KeepaliveInterval: 5,
				Peers:             []kubevip.BGPPeer{},
			},
			wantPeerConfig: kubevip.BGPPeer{
				AS: 64000,
			},
			wantErr: false,
		},
		{
			name: "single peer ip",
			args: args{
				node: &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"bgp/node-asn": "65000",
							"bgp/peer-asn": "64000",
							"bgp/src-ip":   "10.0.0.254",
							"bgp/peer-ip":  "10.0.0.1",
							"bgp/bgp-pass": "cGFzc3dvcmQ=", // base64 encoded.
						}},
				},
				prefix: "bgp",
			},
			wantConfig: kubevip.BGPConfig{
				AS:       65000,
				RouterID: "10.0.0.254",
				SourceIP: "10.0.0.254",
				Peers: []kubevip.BGPPeer{
					{
						AS:       64000,
						Address:  "10.0.0.1",
						Password: "password",
					},
				},
			},
			wantPeerConfig: kubevip.BGPPeer{
				AS:       64000,
				Address:  "10.0.0.1",
				Password: "password",
			},
			wantErr: false,
		},
		{
			name: "multi peer ip",
			args: args{
				node: &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"bgp/node-asn": "65000",
							"bgp/peer-asn": "64000",
							"bgp/src-ip":   "10.0.0.254",
							"bgp/peer-ip":  "10.0.0.1,10.0.0.2,10.0.0.3",
							"bgp/bgp-pass": "cGFzc3dvcmQ=", // base64 encoded.
						}},
				},
				prefix: "bgp",
			},
			wantConfig: kubevip.BGPConfig{
				AS:       65000,
				RouterID: "10.0.0.254",
				SourceIP: "10.0.0.254",
				Peers: []kubevip.BGPPeer{
					{
						AS:       64000,
						Address:  "10.0.0.1",
						Password: "password",
					},
					{
						AS:       64000,
						Address:  "10.0.0.2",
						Password: "password",
					},
					{
						AS:       64000,
						Address:  "10.0.0.3",
						Password: "password",
					},
				},
			},
			wantPeerConfig: kubevip.BGPPeer{
				AS:       64000,
				Address:  "10.0.0.3",
				Password: "password",
			},
			wantErr: false,
		},
		{
			name: "multi peer ip with ordered annotations",
			args: args{
				node: &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"bgp/bgp-peers-0-node-asn": "65000",
							"bgp/bgp-peers-0-peer-asn": "64000",
							"bgp/bgp-peers-0-peer-ip":  "10.0.0.1,10.0.0.2,10.0.0.3",
							"bgp/bgp-peers-0-src-ip":   "10.0.0.254",
							"bgp/bgp-peers-0-bgp-pass": "cGFzc3dvcmQ=", // base64 encoded.
						}},
				},
				prefix: "bgp",
			},
			wantConfig: kubevip.BGPConfig{
				AS:       65000,
				RouterID: "10.0.0.254",
				SourceIP: "10.0.0.254",
				Peers: []kubevip.BGPPeer{
					{
						AS:       64000,
						Address:  "10.0.0.1",
						Password: "password",
					},
					{
						AS:       64000,
						Address:  "10.0.0.2",
						Password: "password",
					},
					{
						AS:       64000,
						Address:  "10.0.0.3",
						Password: "password",
					},
				},
			},
			wantPeerConfig: kubevip.BGPPeer{
				AS:       64000,
				Address:  "10.0.0.3",
				Password: "password",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bgpConfig, bgpPeerConfig, err := parseBgpAnnotations(tt.args.baseConfig, tt.args.node, tt.args.prefix)
			if !tt.wantErr {
				assert.Equal(t, tt.wantConfig, bgpConfig)
				assert.Equal(t, tt.wantPeerConfig, bgpPeerConfig)
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
