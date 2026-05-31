package controller

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-gonic/gin"
)

type copilotDeviceCompleteRequest struct {
	DeviceCode string `json:"device_code"`
}

func StartCopilotDeviceAuth(c *gin.Context) {
	startCopilotDeviceAuthWithChannelID(c, 0)
}

func StartCopilotDeviceAuthForChannel(c *gin.Context) {
	channelID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, fmt.Errorf("invalid channel id: %w", err))
		return
	}
	startCopilotDeviceAuthWithChannelID(c, channelID)
}

func startCopilotDeviceAuthWithChannelID(c *gin.Context, channelID int) {
	channelProxy := ""
	if channelID > 0 {
		ch, ok := validateCopilotChannel(c, channelID)
		if !ok {
			return
		}
		channelProxy = ch.GetSetting().Proxy
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
	defer cancel()

	flow, err := service.StartCopilotDeviceFlow(ctx, channelProxy)
	if err != nil {
		common.SysError("failed to start copilot device flow: " + err.Error())
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "GitHub device authorization start failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"device_code":               flow.DeviceCode,
			"user_code":                 flow.UserCode,
			"verification_uri":          flow.VerificationURI,
			"verification_uri_complete": flow.VerificationURIComplete,
			"expires_in":                flow.ExpiresIn,
			"interval":                  flow.Interval,
		},
	})
}

func CompleteCopilotDeviceAuth(c *gin.Context) {
	completeCopilotDeviceAuthWithChannelID(c, 0)
}

func CompleteCopilotDeviceAuthForChannel(c *gin.Context) {
	channelID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, fmt.Errorf("invalid channel id: %w", err))
		return
	}
	completeCopilotDeviceAuthWithChannelID(c, channelID)
}

func completeCopilotDeviceAuthWithChannelID(c *gin.Context, channelID int) {
	req := copilotDeviceCompleteRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}

	channelProxy := ""
	if channelID > 0 {
		ch, ok := validateCopilotChannel(c, channelID)
		if !ok {
			return
		}
		channelProxy = ch.GetSetting().Proxy
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
	defer cancel()

	tokenRes, err := service.CompleteCopilotDeviceFlow(ctx, req.DeviceCode, channelProxy)
	if err != nil {
		common.SysError("failed to complete copilot device flow: " + err.Error())
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}

	if channelID > 0 {
		if err := model.DB.Model(&model.Channel{}).Where("id = ?", channelID).Update("key", tokenRes.AccessToken).Error; err != nil {
			common.ApiError(c, err)
			return
		}
		model.InitChannelCache()
		service.ResetProxyClientCache()
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "saved",
			"data": gin.H{
				"channel_id": channelID,
				"token_type": tokenRes.TokenType,
				"scope":      tokenRes.Scope,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "generated",
		"data": gin.H{
			"key":        tokenRes.AccessToken,
			"token_type": tokenRes.TokenType,
			"scope":      tokenRes.Scope,
		},
	})
}

func validateCopilotChannel(c *gin.Context, channelID int) (*model.Channel, bool) {
	ch, err := model.GetChannelById(channelID, false)
	if err != nil {
		common.ApiError(c, err)
		return nil, false
	}
	if ch == nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "channel not found"})
		return nil, false
	}
	if ch.Type != constant.ChannelTypeCopilot {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "channel type is not Copilot"})
		return nil, false
	}
	return ch, true
}
