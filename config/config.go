package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	SplitwiseBearerToken string
	UserBSplitwiseID     int64
	UserALunchMoney      LunchMoneyUserConfig
	UserBLunchMoney      LunchMoneyUserConfig
	TestMode             bool
}

type LunchMoneyUserConfig struct {
	BearerToken             string
	SplitwiseAccountAssetID int64
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

	var err error
	cfg.UserBSplitwiseID, err = strconv.ParseInt(os.Getenv("USER_B_SPLITWISE_ID"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid USER_B_SPLITWISE_ID: %w", err)
	}

	userA, err := loadUserConfig("USER_A")
	if err != nil {
		return nil, fmt.Errorf("user A config: %w", err)
	}
	cfg.UserALunchMoney = userA

	userB, err := loadUserConfig("USER_B")
	if err != nil {
		return nil, fmt.Errorf("user B config: %w", err)
	}
	cfg.UserBLunchMoney = userB

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

func loadUserConfig(user string) (LunchMoneyUserConfig, error) {
	var cfg LunchMoneyUserConfig

	cfg.BearerToken = os.Getenv(user + "_LUNCHMONEY_BEARER_TOKEN")
	if cfg.BearerToken == "" {
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
	cfg.SplitwiseAccountAssetID = assetID

	return cfg, nil
}
