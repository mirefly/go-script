package gitsyncer

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestReadToml(t *testing.T) {
	c := viper.New()
	c.SetConfigName(".gitsyncer")
	c.AddConfigPath("./testdata")

	_, err := loadConfig(c)
	assert.NoError(t, err)
}
