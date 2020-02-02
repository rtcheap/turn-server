package main

import (
	"strconv"

	"github.com/CzarSimon/httputil/environ"
	"github.com/CzarSimon/httputil/jwt"
	"go.uber.org/zap"
)

type config struct {
	service        serviceConfig
	turn           turnConfig
	registryURL    string
	jwtCredentials jwt.Credentials
}

func getConfig() config {
	return config{
		service:        getServiceConfig(),
		turn:           getTurnConfig(),
		registryURL:    environ.Get("SERVICE_REGISTRY_URL", "http://service-registry:8080"),
		jwtCredentials: getJwtCredentials(),
	}
}

type turnConfig struct {
	udpPort int
	tcpPort int
	realm   string
	ip      string
}

func getTurnConfig() turnConfig {
	return turnConfig{
		udpPort: getIntFromEnvironment("TURN_UDP_PORT", 3478),
		realm:   environ.Get("TURN_REALM", "rtcheap"),
		ip:      environ.MustGet("TURN_PUBLIC_IP"),
	}
}

type serviceConfig struct {
	id   string
	port int
}

func getServiceConfig() serviceConfig {
	return serviceConfig{
		port: getIntFromEnvironment("SERVICE_PORT", 8080),
	}
}

func getJwtCredentials() jwt.Credentials {
	return jwt.Credentials{
		Issuer: environ.MustGet("JWT_ISSUER"),
		Secret: environ.MustGet("JWT_SECRET"),
	}
}

func getIntFromEnvironment(name string, defaultValue int) int {
	val := environ.Get(name, strconv.Itoa(defaultValue))
	intVal, err := strconv.Atoi(val)
	if err != nil {
		log.Fatal("failed to parse integer environment variable "+name, zap.String("value", val))
	}

	return intVal
}
