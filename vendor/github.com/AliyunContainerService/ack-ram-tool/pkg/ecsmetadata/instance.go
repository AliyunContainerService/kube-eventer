package ecsmetadata

import (
	"context"
	"time"
)

func (c *Client) GetInstanceType(ctx context.Context) (string, error) {
	return c.getTidyStringData(ctx, "/latest/meta-data/instance/instance-type")
}

func (c *Client) GetInstanceName(ctx context.Context) (string, error) {
	return c.getTidyStringData(ctx, "/latest/meta-data/instance/instance-name")
}

func (c *Client) GetInstanceId(ctx context.Context) (string, error) {
	return c.getTidyStringData(ctx, "/latest/meta-data/instance-id")
}

func (c *Client) GetImageId(ctx context.Context) (string, error) {
	return c.getTidyStringData(ctx, "/latest/meta-data/image-id")
}

func (c *Client) GetSerialNumber(ctx context.Context) (string, error) {
	return c.getTidyStringData(ctx, "/latest/meta-data/serial-number")
}

func (c *Client) GetSpotTerminationTime(ctx context.Context) (time.Time, error) {
	data, err := c.getTidyStringData(ctx,
		"/latest/meta-data/instance/spot/termination-time")
	if err != nil {
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339, data)
}
