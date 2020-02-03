package main

import (
	"context"
	"io"
	"time"

	"github.com/CzarSimon/httputil"
	"github.com/CzarSimon/httputil/client"
	"github.com/CzarSimon/httputil/client/rpc"
	"github.com/CzarSimon/httputil/jwt"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/rtcheap/dto"
	"github.com/rtcheap/service-clients/go/serviceregistry"
	"github.com/rtcheap/turn-server/internal/repository"
	"github.com/rtcheap/turn-server/internal/service"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"go.uber.org/zap"
)

type env struct {
	cfg             config
	userService     service.UserService
	serviceregistry serviceregistry.Client
	traceCloser     io.Closer
}

func (e *env) register() error {
	span, ctx := opentracing.StartSpanFromContext(context.Background(), "register_self")
	defer span.Finish()

	self := dto.Service{
		Application: "turn-server",
		Location:    e.cfg.turn.ip,
		Port:        e.cfg.service.port,
		Status:      dto.StatusHealty,
	}

	svc, err := e.serviceregistry.Register(ctx, self)
	if err != nil {
		return err
	}

	e.cfg.service.id = svc.ID
	return nil
}

func (e *env) unregister() error {
	span, ctx := opentracing.StartSpanFromContext(context.Background(), "unregister_self")
	defer span.Finish()

	err := e.serviceregistry.SetStatus(ctx, e.cfg.service.id, dto.StatusUnhealthy)
	if err != nil {
		return err
	}

	return nil
}

func (e *env) checkHealth() error {
	return nil
}

func (e *env) close() {
	err := e.unregister()
	if err != nil {
		log.Error("the service failed to unregister itself", zap.Error(err))
	}

	err = e.traceCloser.Close()
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

	userService := service.UserService{
		Realm: cfg.turn.realm,
		Keys:  repository.NewKeyRepository(),
	}

	return &env{
		cfg:             cfg,
		userService:     userService,
		serviceregistry: registryClient,
		traceCloser:     closer,
	}
}

func notImplemented(c *gin.Context) {
	err := httputil.NotImplementedError(nil)
	c.Error(err)
}
