//go:build m1ss_minimal

package core

import (
	suiVless "github.com/alireza0/s-ui/core/protocol/vless"
	"github.com/alireza0/s-ui/core/usersession"
	"github.com/sagernet/sing-box/option"
)

func (c *Core) UpdateInboundUsers(config []byte) (bool, error) {
	box, err := c.running()
	if err != nil {
		return false, err
	}
	var inboundConfig option.Inbound
	if err = inboundConfig.UnmarshalJSONContext(box.scope.Context(), config); err != nil {
		return false, err
	}
	inb, found := box.inbound.Get(inboundConfig.Tag)
	if !found {
		return false, nil
	}
	if options, ok := inboundConfig.Options.(*option.VLESSInboundOptions); ok {
		if in, ok := inb.(*suiVless.Inbound); ok {
			return true, in.UpdateUsers(options.Users)
		}
	}
	return false, nil
}

func (c *Core) CloseInboundUserSessions(tag string, keep map[string]struct{}) int {
	box, err := c.running()
	if err != nil {
		return 0
	}
	inb, found := box.inbound.Get(tag)
	if !found {
		return 0
	}
	if closer, ok := inb.(usersession.Closer); ok {
		return closer.CloseUserSessions(keep)
	}
	return 0
}

func (c *Core) KickUserSessions(user string) int {
	box, err := c.running()
	if err != nil {
		return 0
	}
	kicked := 0
	for _, inb := range box.inbound.Inbounds() {
		if closer, ok := inb.(usersession.Closer); ok {
			kicked += closer.KickUserSessions(user)
		}
	}
	return kicked
}
