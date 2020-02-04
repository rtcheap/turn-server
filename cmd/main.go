package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/CzarSimon/httputil"
	"github.com/CzarSimon/httputil/jwt"
	"github.com/CzarSimon/httputil/logger"
	"go.uber.org/zap"
)

var log = logger.GetDefaultLogger("turn-server/main")

func main() {
	e := setupEnv()
	defer e.close()
	err := e.register()
	if err != nil {
		log.Fatal("service registration failed", zap.Error(err))
	}

	go startService(e)

	server, err := newTurnServer(e)
	if err != nil {
		log.Fatal("Unexpected error when creating the TURN server", zap.Error(err))
	}

	log.Info(fmt.Sprintf("Started TURN forwarding listening on port: udp/%d", e.cfg.turn.udpPort))
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	err = server.Close()
	if err != nil {
		log.Error("Error encountered when closing the TURN server.", zap.Error(err))
	}
}

func startService(e *env) {
	server := newServer(e)
	log.Info(fmt.Sprintf("Started turn-server listening on port: tcp/%d", e.cfg.service.port))

	err := server.ListenAndServe()
	if err != nil {
		log.Error("Unexpected error stopped server.", zap.Error(err))
	}
}

func newServer(e *env) *http.Server {
	r := httputil.NewRouter("turn-server", e.checkHealth)

	rbac := httputil.RBAC{
		Verifier: jwt.NewVerifier(e.cfg.jwtCredentials, time.Minute),
	}
	v1 := r.Group("/v1", rbac.Secure(jwt.SystemRole))

	v1.POST("/sessions", e.addSession)
	v1.GET("/sessions/statistics", e.getSessionStatistics)

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", e.cfg.service.port),
		Handler: r,
	}
}
