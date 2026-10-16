//go:build m1ss_minimal

package core

import (
	"context"

	suiVless "github.com/alireza0/s-ui/core/protocol/vless"

	"github.com/sagernet/sing-box/adapter"
	sbCertificate "github.com/sagernet/sing-box/adapter/certificate"
	"github.com/sagernet/sing-box/adapter/endpoint"
	"github.com/sagernet/sing-box/adapter/inbound"
	"github.com/sagernet/sing-box/adapter/outbound"
	"github.com/sagernet/sing-box/adapter/service"
	"github.com/sagernet/sing-box/dns"
	"github.com/sagernet/sing-box/dns/transport"
	"github.com/sagernet/sing-box/dns/transport/fakeip"
	"github.com/sagernet/sing-box/dns/transport/hosts"
	"github.com/sagernet/sing-box/dns/transport/local"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing-box/protocol/block"
	"github.com/sagernet/sing-box/protocol/direct"
	"github.com/sagernet/sing-box/protocol/group"
	"github.com/sagernet/sing-box/protocol/vless"
	E "github.com/sagernet/sing/common/exceptions"
)

func InboundRegistry() *inbound.Registry {
	registry := inbound.NewRegistry()
	inbound.Register[option.VLESSInboundOptions](registry, "vless", func(ctx context.Context, router adapter.Router, logger log.ContextLogger, tag string, options option.VLESSInboundOptions) (adapter.Inbound, error) {
		if options.Transport != nil && options.Transport.Type != "" {
			return nil, E.New("this build supports VLESS over native TCP only; clear transport.type")
		}
		return suiVless.NewInbound(ctx, router, logger, tag, options)
	})
	return registry
}

func OutboundRegistry() *outbound.Registry {
	registry := outbound.NewRegistry()
	// Direct egress and routing groups are needed by the panel's normal routing
	// configuration; they do not open additional proxy protocol listeners.
	direct.RegisterOutbound(registry)
	block.RegisterOutbound(registry)
	group.RegisterSelector(registry)
	group.RegisterURLTest(registry)
	vless.RegisterOutbound(registry)
	return registry
}

func EndpointRegistry() *endpoint.Registry { return endpoint.NewRegistry() }

func DNSTransportRegistry() *dns.TransportRegistry {
	registry := dns.NewTransportRegistry()
	transport.RegisterTCP(registry)
	transport.RegisterUDP(registry)
	transport.RegisterTLS(registry)
	transport.RegisterHTTPS(registry)
	hosts.RegisterTransport(registry)
	local.RegisterTransport(registry)
	fakeip.RegisterTransport(registry)
	return registry
}

func ServiceRegistry() *service.Registry { return service.NewRegistry() }

func CertificateProviderRegistry() *sbCertificate.Registry { return sbCertificate.NewRegistry() }
