package api

import (
	"github.com/alireza0/s-ui/updater"
	"github.com/gin-gonic/gin"
)

func (a *ApiService) CheckUpdate(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	release, err := updater.Default().Check(c.Request.Context())
	jsonObj(c, release, err)
}

func (a *ApiService) UpdateStatus(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	jsonObj(c, updater.Default().Status(), nil)
}

func (a *ApiService) StartUpdate(c *gin.Context) {
	state, err := updater.Default().Start(c.Request.Context(), c.PostForm("version"))
	jsonObj(c, state, err)
}
