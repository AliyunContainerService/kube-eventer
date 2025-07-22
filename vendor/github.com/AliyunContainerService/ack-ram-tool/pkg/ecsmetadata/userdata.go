package ecsmetadata

import (
	"context"
)

// https://help.aliyun.com/zh/ecs/user-guide/customize-the-initialization-configuration-for-an-instance

func (c *Client) GetUserData(ctx context.Context) (string, error) {
	return c.getRawStringData(ctx, "/latest/user-data")
}
