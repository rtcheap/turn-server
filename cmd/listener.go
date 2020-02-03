package main

import (
	"fmt"
	"net"

	"github.com/pion/turn/v2"
)

func newTurnServer(e *env) (*turn.Server, error) {
	address := fmt.Sprintf("0.0.0.0:%d", e.cfg.turn.udpPort)
	udpListener, err := net.ListenPacket("udp4", address)
	if err != nil {
		return nil, fmt.Errorf("failed to create TURN server listener: %w", err)
	}

	serverCfg := turn.ServerConfig{
		Realm:       e.cfg.turn.realm,
		AuthHandler: e.userService.FindKey,
		PacketConnConfigs: []turn.PacketConnConfig{
			{
				PacketConn: udpListener,
				RelayAddressGenerator: &turn.RelayAddressGeneratorStatic{
					RelayAddress: net.ParseIP(e.cfg.turn.ip),
					Address:      "0.0.0.0",
				},
			},
		},
	}

	s, err := turn.NewServer(serverCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create TURN server: %w", err)
	}

	return s, nil
}
