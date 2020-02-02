package main

import (
	"github.com/CzarSimon/httputil/environ"
	"github.com/CzarSimon/httputil/jwt"
)

type config struct {
	port           string
	serviceID      string
	serviceName    string
	registryURL    string
	jwtCredentials jwt.Credentials
}

func getConfig() config {
	return config{
		port:           environ.Get("SERVICE_PORT", "8080"),
		serviceName:    environ.MustGet("SERVICE_NAME"),
		registryURL:    environ.Get("SERVICE_REGISTRY_URL", "http://service-registry:8080"),
		jwtCredentials: getJwtCredentials(),
	}
}

func getJwtCredentials() jwt.Credentials {
	return jwt.Credentials{
		Issuer: environ.MustGet("JWT_ISSUER"),
		Secret: environ.MustGet("JWT_SECRET"),
	}
}
