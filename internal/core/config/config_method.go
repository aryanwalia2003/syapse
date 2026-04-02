package config

import (
	"fmt"

	"github.com/aryanwalia/synapse/internal/core/errors"
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

func Load() (*SynapseConfig, error) {
	_ = godotenv.Load()

	cfg := &SynapseConfig{}
	if err := env.Parse(cfg); err != nil {
		return nil, errors.Wrap(err, errors.CodeInternal, fmt.Sprintf("failed to parse configuration: %v", err))
	}

	return cfg, nil
}
