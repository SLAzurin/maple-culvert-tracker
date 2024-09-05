package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/slazurin/maple-culvert-tracker/internal/apiredis"
)

type SettingsController struct{}

type EditableSetting struct {
	HumanReadableDescription apiredis.HumanReadableDescriptions
	Value                    string `json:"value"`
	Key                      string `json:"key"`
}

func (s SettingsController) GETEditable(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"erm": "yea under construction",
	})
}
