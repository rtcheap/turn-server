package main

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/CzarSimon/httputil"
	"github.com/CzarSimon/httputil/client"
	"github.com/CzarSimon/httputil/client/rpc"
	"github.com/CzarSimon/httputil/jwt"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/rtcheap/dto"
	"github.com/rtcheap/service-clients/go/serviceregistry"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"go.uber.org/zap"
)

type env struct {
	cfg             config
	serviceregistry serviceregistry.Client
	traceCloser     io.Closer
}

func (e *env) register() error {
	span, ctx := opentracing.StartSpanFromContext(context.Background(), "register_self")
	defer span.Finish()

	port, err := strconv.Atoi(e.cfg.port)
	if err != nil {
		return fmt.Errorf("failed to parse port. %w", err)
	}

	self := dto.Service{
		Application: "turn-server",
		Location:    e.cfg.serviceName,
		Port:        port,
		Status:      dto.StatusHealty,
	}

	svc, err := e.serviceregistry.Register(ctx, self)
	if err != nil {
		return fmt.Errorf("the service failed to register itself. %w", err)
	}

	e.cfg.serviceID = svc.ID
	return nil
}

func (e *env) checkHealth() error {
	return nil
}

func (e *env) close() {
	err := e.traceCloser.Close()
	if err != nil {
		log.Error("failed to close tracer connection", zap.Error(err))
	}
}

func setupEnv() *env {
	jcfg, err := jaegercfg.FromEnv()
	if err != nil {
		log.Fatal("failed to create jaeger configuration", zap.Error(err))
	}

	tracer, closer, err := jcfg.NewTracer()
	if err != nil {
		log.Fatal("failed to create tracer", zap.Error(err))
	}

	opentracing.SetGlobalTracer(tracer)
	cfg := getConfig()
	registryClient := serviceregistry.NewClient(client.Client{
		Issuer:    jwt.NewIssuer(cfg.jwtCredentials),
		BaseURL:   cfg.registryURL,
		UserAgent: "turn-server",
		Role:      jwt.SystemRole,
		RPCClient: rpc.NewClient(5 * time.Second),
	})

	return &env{
		cfg:             cfg,
		serviceregistry: registryClient,
		traceCloser:     closer,
	}
}

func notImplemented(c *gin.Context) {
	err := httputil.NotImplementedError(nil)
	c.Error(err)
}
