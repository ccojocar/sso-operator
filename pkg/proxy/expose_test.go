package proxy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExposeConfig(t *testing.T) {
	config := &ExposeConfig{
		Domain:   "test",
		Exposer:  "Ingress",
		PathMode: "",
		HTTP:     true,
		TLSAcme:  false,
		Services: []string{"test"},
	}

	strConfig, err := renderExposeConfig(config)

	assert.NoError(t, err, "should render expose config to YAML without error")
	assert.NotEmpty(t, strConfig, "expose config should not be empty")
}
