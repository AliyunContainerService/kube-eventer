package ecsmetadata

import (
	"context"
	"encoding/json"
	"net/url"
)

// https://help.aliyun.com/zh/ecs/user-guide/use-instance-identities

type Document struct {
	AccountId      string `json:"account-id"` // always is empty?
	OwnerAccountId string `json:"owner-account-id"`
	InstanceId     string `json:"instance-id"`
	Mac            string `json:"mac"`
	RegionId       string `json:"region-id"`
	SerialNumber   string `json:"serial-number"`
	ZoneId         string `json:"zone-id"`
	InstanceType   string `json:"instance-type"`
	ImageId        string `json:"image-id"`
	PrivateIp      string `json:"private-ip"`
}

func (c *Client) GetDocument(ctx context.Context) (*Document, error) {
	data, err := c.GetRawDocument(ctx)
	if err != nil {
		return nil, err
	}

	var doc Document
	if err := json.Unmarshal([]byte(data), &doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

func (c *Client) GetRawDocument(ctx context.Context) (string, error) {
	return c.getRawStringData(ctx, "/latest/dynamic/instance-identity/document")
}

func (c *Client) NewDocumentPKCS7Signature(ctx context.Context, audience string) (string, error) {
	path := "/latest/dynamic/instance-identity/pkcs7"
	if audience != "" {
		val := url.Values{}
		val.Set("audience", audience)
		path = path + "?" + val.Encode()
	}

	return c.getRawStringData(ctx, path)
}
