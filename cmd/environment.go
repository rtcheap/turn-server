package main

import (
    "github.com/CzarSimon/httputil"
    "github.com/gin-gonic/gin"
)

type env struct {
    cfg config
}

func notImplemented(c *gin.Context) {
    err := httputil.NotImplementedError(nil)
    c.Error(err)
}
