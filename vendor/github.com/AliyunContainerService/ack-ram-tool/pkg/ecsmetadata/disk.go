package ecsmetadata

import (
	"context"
	"fmt"
)

func (c *Client) GetDisks(ctx context.Context) ([]string, error) {
	data, err := c.getTidyStringData(ctx, "/latest/meta-data/disks/")
	if err != nil {
		return nil, err
	}
	return parsePathNames(data), nil
}

func (c *Client) GetDiskId(ctx context.Context, diskSerial string) (string, error) {
	return c.getTidyStringData(ctx,
		fmt.Sprintf("/latest/meta-data/disks/%s/id", diskSerial))
}

func (c *Client) GetDiskName(ctx context.Context, diskSerial string) (string, error) {
	return c.getTidyStringData(ctx,
		fmt.Sprintf("/latest/meta-data/disks/%s/name", diskSerial))
}
