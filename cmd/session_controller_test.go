package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/CzarSimon/httputil/client/rpc"
	"github.com/CzarSimon/httputil/id"
	"github.com/CzarSimon/httputil/jwt"
	"github.com/opentracing/opentracing-go"
	"github.com/pion/turn/v2"
	"github.com/rtcheap/dto"
	"github.com/rtcheap/turn-server/internal/repository"
	"github.com/rtcheap/turn-server/internal/service"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestAddNewUserSession(t *testing.T) {
	assert := assert.New(t)
	e, _ := createTestEnv()
	server := newServer(e)

	// Testcase: Happy path - A new session should be created.
	session1 := dto.Session{
		UserID: id.New(),
		Key:    "secret-credential-1",
	}
	key, ok := e.userService.FindKey(session1.UserID, e.cfg.turn.realm, nil)
	assert.False(ok)
	assert.Nil(key)

	req := createTestRequest("/v1/sessions", http.MethodPost, jwt.SystemRole, session1)
	res := performTestRequest(server.Handler, req)
	assert.Equal(http.StatusOK, res.Code)

	expectedKey := turn.GenerateAuthKey(session1.UserID, e.cfg.turn.realm, session1.Key)
	key, ok = e.userService.FindKey(session1.UserID, "", nil)
	assert.True(ok)
	assert.NotNil(key)
	assert.Equal(expectedKey, key)

	// Testcase: Happy path - Another new session should be created.
	session2 := dto.Session{
		UserID: id.New(),
		Key:    "secret-credential-2",
	}
	key, ok = e.userService.FindKey(session2.UserID, e.cfg.turn.realm, nil)
	assert.False(ok)
	assert.Nil(key)

	req = createTestRequest("/v1/sessions", http.MethodPost, jwt.SystemRole, session2)
	res = performTestRequest(server.Handler, req)
	assert.Equal(http.StatusOK, res.Code)

	expectedKey = turn.GenerateAuthKey(session2.UserID, e.cfg.turn.realm, session2.Key)
	key, ok = e.userService.FindKey(session2.UserID, "", nil)
	assert.True(ok)
	assert.NotNil(key)
	assert.Equal(expectedKey, key)

	// Testcase: Sad path a new session should not override an existing one.
	session3 := dto.Session{
		UserID: session1.UserID,
		Key:    "secret-credential-3",
	}

	req = createTestRequest("/v1/sessions", http.MethodPost, jwt.SystemRole, session3)
	res = performTestRequest(server.Handler, req)
	assert.Equal(http.StatusConflict, res.Code)

	expectedKey = turn.GenerateAuthKey(session3.UserID, e.cfg.turn.realm, session1.Key)
	key, ok = e.userService.FindKey(session3.UserID, "", nil)
	assert.True(ok)
	assert.NotNil(key)
	assert.Equal(expectedKey, key)
}

func TestGetSessionStatistics(t *testing.T) {
	assert := assert.New(t)
	e, _ := createTestEnv()
	server := newServer(e)

	// Testcase: Happy path - Initally 0 sessions should have been created.

	req := createTestRequest("/v1/sessions/statistics", http.MethodGet, jwt.SystemRole, nil)
	res := performTestRequest(server.Handler, req)
	assert.Equal(http.StatusOK, res.Code)

	var stats dto.SessionStatistics
	err := rpc.DecodeJSON(res.Result(), &stats)
	assert.NoError(err)
	assert.Equal(uint64(0), stats.Started)
	assert.Equal(uint64(0), stats.InProgress())
	assert.Equal(uint64(0), stats.Ended)

	expectedCount := 10
	for i := 0; i < expectedCount; i++ {
		session := dto.Session{
			UserID: id.New(),
			Key:    fmt.Sprintf("secret-credential-%d", i+1),
		}

		req := createTestRequest("/v1/sessions", http.MethodPost, jwt.SystemRole, session)
		res := performTestRequest(server.Handler, req)
		assert.Equal(http.StatusOK, res.Code)
	}

	req = createTestRequest("/v1/sessions/statistics", http.MethodGet, jwt.SystemRole, nil)
	res = performTestRequest(server.Handler, req)
	assert.Equal(http.StatusOK, res.Code)

	err = rpc.DecodeJSON(res.Result(), &stats)
	assert.NoError(err)
	assert.Equal(uint64(10), stats.Started)
	assert.Equal(uint64(10), stats.InProgress())
	assert.Equal(uint64(0), stats.Ended)
}

func TestHealthCheck(t *testing.T) {
	assert := assert.New(t)
	e, _ := createTestEnv()
	server := newServer(e)

	req := createTestRequest("/health", http.MethodGet, "", nil)
	res := performTestRequest(server.Handler, req)
	assert.Equal(http.StatusOK, res.Code)
}

func TestPermissions(t *testing.T) {
	assert := assert.New(t)
	e, _ := createTestEnv()
	server := newServer(e)

	cases := []struct {
		method string
		route  string
	}{
		{method: http.MethodPost, route: "/v1/sessions"},
		{method: http.MethodGet, route: "/v1/sessions/statistics"},
	}

	badRoles := []string{jwt.AnonymousRole, jwt.AdminRole, ""}

	for _, tc := range cases {
		for _, role := range badRoles {
			req := createTestRequest(tc.route, tc.method, role, nil)
			res := performTestRequest(server.Handler, req)

			expectedStatus := http.StatusForbidden
			if role == "" {
				expectedStatus = http.StatusUnauthorized
			}
			assert.Equal(expectedStatus, res.Code)
		}
	}
}

// ---- Test utils ----

func createTestEnv() (*env, context.Context) {
	cfg := config{
		turn: turnConfig{
			realm: "rtcheap",
		},
		jwtCredentials: getTestJWTCredentials(),
	}
	repo := repository.NewKeyRepository()
	svc := service.UserService{
		Realm: cfg.turn.realm,
		Keys:  repo,
	}

	e := &env{
		cfg:         cfg,
		userService: svc,
	}

	return e, context.Background()
}

func performTestRequest(r http.Handler, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func createTestRequest(route, method, role string, body interface{}) *http.Request {
	client := rpc.NewClient(time.Second)
	req, err := client.CreateRequest(method, route, body)
	if err != nil {
		log.Fatal("Failed to create request", zap.Error(err))
	}

	span := opentracing.StartSpan(fmt.Sprintf("%s.%s", method, route))
	opentracing.GlobalTracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header),
	)

	if role == "" {
		return req
	}

	issuer := jwt.NewIssuer(getTestJWTCredentials())
	token, err := issuer.Issue(jwt.User{
		ID:    "service-registry-user",
		Roles: []string{role},
	}, time.Hour)

	req.Header.Add("Authorization", "Bearer "+token)
	return req
}

func getTestJWTCredentials() jwt.Credentials {
	return jwt.Credentials{
		Issuer: "service-registry-test",
		Secret: "very-secret-secret",
	}
}
