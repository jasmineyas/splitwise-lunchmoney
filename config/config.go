package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	SplitwiseBearerToken string
	UserA                UserConfig
	UserB                UserConfig
	TestMode             bool
}

type UserConfig struct {
	LunchMoneyBearerToken             string
	LunchMoneySplitwiseAccountAssetID int64
}

func Load() (*Config, error) {

	// shared state, load once, pass around

	_ = godotenv.Load()

	cfg := &Config{}

	cfg.SplitwiseBearerToken = os.Getenv("USER_A_SPLITWISE_BEARER_TOKEN")
	if cfg.SplitwiseBearerToken == "" {
		cfg.SplitwiseBearerToken = os.Getenv("USER_B_SPLITWISE_BEARER_TOKEN")
	}
	if cfg.SplitwiseBearerToken == "" {
		return nil, fmt.Errorf("missing SPLITWISE_BEARER_TOKEN in environment")
	}

	userA, err := loadUserConfig("USER_A")
	if err != nil {
		return nil, fmt.Errorf("user A config: %w", err)
	}
	cfg.UserA = userA

	userB, err := loadUserConfig("USER_B")
	if err != nil {
		return nil, fmt.Errorf("user B config: %w", err)
	}
	cfg.UserB = userB

	testModeStr := os.Getenv("TEST")
	if testModeStr != "" {
		testMode, err := strconv.ParseBool(testModeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid TEST value: %w", err)
		}
		cfg.TestMode = testMode
	} else {
		cfg.TestMode = false
	}

	return cfg, nil
}

func loadUserConfig(user string) (UserConfig, error) {
	var cfg UserConfig

	cfg.LunchMoneyBearerToken = os.Getenv(user + "_LUNCHMONEY_BEARER_TOKEN")
	if cfg.LunchMoneyBearerToken == "" {
		return cfg, fmt.Errorf("missing %s_LUNCHMONEY_BEARER_TOKEN in environment", user)
	}

	assetIDStr := os.Getenv(user + "_LUNCHMONEY_SPLITWISE_ACCOUNT_ASSET_ID")
	if assetIDStr == "" {
		return cfg, fmt.Errorf("missing %s_LUNCHMONEY_SPLITWISE_ACCOUNT_ASSET_ID in environment", user)
	}

	assetID, err := strconv.ParseInt(assetIDStr, 10, 64)
	if err != nil {
		return cfg, fmt.Errorf("invalid %s_LUNCHMONEY_SPLITWISE_ACCOUNT_ASSET_ID: %w", user, err)
	}
	cfg.LunchMoneySplitwiseAccountAssetID = assetID

	return cfg, nil
}
