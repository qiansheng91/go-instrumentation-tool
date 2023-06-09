package config

import (
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
)

func TestParser(t *testing.T) {
	path, _ := filepath.Abs("./configuration.yaml")
	instrumentationInfo, _ := Parser(path)

	assert.Equal(t, 1, len(instrumentationInfo.Plugins()))
	injectPlugin := instrumentationInfo.Plugins()[0]

	assert.Equal(t, "otel-gin-plugin", injectPlugin.PluginName())
	assert.Truef(t, injectPlugin.CheckIfPluginIsEnabled(), "plugin %s should be enabled",
		injectPlugin.PluginName())
	targetP := injectPlugin.CheckIfMatchInjectPlugin("github.com/gin-gonic/gin")
	assert.NotNil(t, targetP)

	assert.Equal(t, 1, len(targetP.PointCuts()))
	assert.NotNil(t, targetP.PointCuts()["New"])

	pointCut := targetP.PointCuts()["New"]
	assert.Equal(t, "New", pointCut.TargetSignature())
	assert.Equal(t, "(*Engine).ServeHTTP", pointCut.BeforeAdvice())
	assert.Equal(t, "func(*gin.Context)", pointCut.AfterAdvice())
}
