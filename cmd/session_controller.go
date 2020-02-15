package main

import (
	"fmt"
	"net/http"

	"github.com/CzarSimon/httputil"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	tracelog "github.com/opentracing/opentracing-go/log"
	"github.com/rtcheap/dto"
)

func (e *env) addSession(c *gin.Context) {
	span, _ := opentracing.StartSpanFromContext(c.Request.Context(), "controller_add_session")
	defer span.Finish()

	var session dto.Session
	err := c.BindJSON(&session)
	if err != nil {
		err = httputil.BadRequestError(fmt.Errorf("failed to parse request body. %w", err))
		span.LogFields(tracelog.Error(err))
		c.Error(err)
		return
	}

	err = e.userService.CreateKey(session)
	if err != nil {
		span.LogFields(tracelog.Error(err))
		c.Error(err)
		return
	}

	httputil.SendOK(c)
}

func (e *env) removeSession(c *gin.Context) {
	span, _ := opentracing.StartSpanFromContext(c.Request.Context(), "controller_add_session")
	defer span.Finish()

	err := e.userService.Remove(c.Param("userID"))
	if err != nil {
		span.LogFields(tracelog.Error(err))
		c.Error(err)
		return
	}

	httputil.SendOK(c)
}

func (e *env) getSessionStatistics(c *gin.Context) {
	span, _ := opentracing.StartSpanFromContext(c.Request.Context(), "controller_get_session_statistics")
	defer span.Finish()

	stats := e.userService.GetStatistics()
	c.JSON(http.StatusOK, stats)
}
