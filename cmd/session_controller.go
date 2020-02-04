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
	span, _ := opentracing.StartSpanFromContext(c.Request.Context(), "controller.addSession")
	defer span.Finish()

	var session dto.Session
	err := c.BindJSON(&session)
	if err != nil {
		err = httputil.BadRequestError(fmt.Errorf("failed to parse request body. %w", err))
		span.LogFields(tracelog.Bool("success", false), tracelog.Error(err))
		c.Error(err)
		return
	}

	err = e.userService.CreateKey(session)
	if err != nil {
		span.LogFields(tracelog.Bool("success", false), tracelog.Error(err))
		c.Error(err)
		return
	}

	span.LogFields(tracelog.Bool("success", true))
	httputil.SendOK(c)
}

func (e *env) getSessionStatistics(c *gin.Context) {
	span, _ := opentracing.StartSpanFromContext(c.Request.Context(), "controller.getSessionStatistics")
	defer span.Finish()

	stats := e.userService.GetStatistics()

	span.LogFields(tracelog.Bool("success", true))
	c.JSON(http.StatusOK, stats)
}
