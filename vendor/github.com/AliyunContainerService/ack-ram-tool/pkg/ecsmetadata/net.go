package ecsmetadata

import (
	"context"
)

func (c *Client) GetVpcId(ctx context.Context) (string, error) {
	return c.getTidyStringData(ctx, "/latest/meta-data/vpc-id")
}

// GetVpcCidrBlockId deprecated, use GetVpcCidrBlock instead
func (c *Client) GetVpcCidrBlockId(ctx context.Context) (string, error) {
	return c.GetVpcCidrBlock(ctx)
}

func (c *Client) GetVpcCidrBlock(ctx context.Context) (string, error) {
	return c.getTidyStringData(ctx, "/latest/meta-data/vpc-cidr-block")
}

func (c *Client) GetVSwitchId(ctx context.Context) (string, error) {
	return c.getTidyStringData(ctx, "/latest/meta-data/vswitch-id")
}

// GetVSwitchCidrBlockId deprecated, use GetVSwitchCidrBlock instead
func (c *Client) GetVSwitchCidrBlockId(ctx context.Context) (string, error) {
	return c.GetVSwitchCidrBlock(ctx)
}

func (c *Client) GetVSwitchCidrBlock(ctx context.Context) (string, error) {
	return c.getTidyStringData(ctx, "/latest/meta-data/vswitch-cidr-block")
}

func (c *Client) GetPrivateIPV4(ctx context.Context) (string, error) {
	return c.getTidyStringData(ctx, "/latest/meta-data/private-ipv4")
}

func (c *Client) GetPublicIPV4(ctx context.Context) (string, error) {
	return c.getTidyStringData(ctx, "/latest/meta-data/public-ipv4")
}

func (c *Client) GetEIPV4(ctx context.Context) (string, error) {
	return c.getTidyStringData(ctx, "/latest/meta-data/eipv4")
}

func (c *Client) GetNetworkType(ctx context.Context) (string, error) {
	return c.getTidyStringData(ctx, "/latest/meta-data/network-type")
}

func (c *Client) GetMac(ctx context.Context) (string, error) {
	return c.getTidyStringData(ctx, "/latest/meta-data/mac")
}

func (c *Client) GetDNSNameServers(ctx context.Context) (string, error) {
	return c.getTidyStringData(ctx, "/latest/meta-data/dns-conf/nameservers")
}

func (c *Client) GetDNSNameServersList(ctx context.Context) ([]string, error) {
	data, err := c.GetDNSNameServers(ctx)
	if err != nil {
		return nil, err
	}
	return parsePathNames(data), nil
}

func (c *Client) GetNTPServers(ctx context.Context) (string, error) {
	return c.getTidyStringData(ctx, "/latest/meta-data/ntp-conf/ntp-servers")
}

func (c *Client) GetNTPServersList(ctx context.Context) ([]string, error) {
	data, err := c.GetNTPServers(ctx)
	if err != nil {
		return nil, err
	}
	return parsePathNames(data), nil
}
