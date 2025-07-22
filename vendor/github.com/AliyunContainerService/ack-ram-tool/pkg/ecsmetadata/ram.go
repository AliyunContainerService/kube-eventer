package ecsmetadata

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type rawCredentials struct {
	AccessKeyId     string `json:"AccessKeyId"`
	AccessKeySecret string `json:"AccessKeySecret"`
	SecurityToken   string `json:"SecurityToken"`
	Expiration      string `json:"Expiration"`
	LastUpdated     string `json:"LastUpdated"`
	Code            string `json:"Code"`
}

type RoleCredentials struct {
	AccessKeyId     string
	AccessKeySecret string
	SecurityToken   string
	Expiration      time.Time
	LastUpdated     time.Time
	Code            string
}

func (c *Client) GetRoleName(ctx context.Context) (string, error) {
	if c.roleName != "" {
		return c.roleName, nil
	}
	return c.getTidyStringData(ctx, "/latest/meta-data/ram/security-credentials/")
}

func (c *Client) GetRawRoleCredentials(ctx context.Context, roleName string) (string, error) {
	data, err := c.getRawStringData(ctx, "/latest/meta-data/ram/security-credentials/"+roleName)
	return data, err
}

func (c *Client) GetRoleCredentials(ctx context.Context, roleName string) (*RoleCredentials, error) {
	data, err := c.GetRawRoleCredentials(ctx, roleName)
	if err != nil {
		return nil, err
	}

	const format = time.RFC3339
	var raw rawCredentials
	if err := json.Unmarshal([]byte(data), &raw); err != nil {
		return nil, fmt.Errorf("parse credentials failed: %w", err)
	}
	exp, err := time.Parse(format, raw.Expiration)
	if err != nil {
		return nil, fmt.Errorf("parse Expiration (%s) failed: %w", raw.Expiration, err)
	}
	last, _ := time.Parse(format, raw.LastUpdated)

	return &RoleCredentials{
		AccessKeyId:     raw.AccessKeyId,
		AccessKeySecret: raw.AccessKeySecret,
		SecurityToken:   raw.SecurityToken,
		Expiration:      exp,
		LastUpdated:     last,
		Code:            raw.Code,
	}, nil
}
