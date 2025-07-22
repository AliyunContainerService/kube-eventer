package ecsmetadata

import (
	"context"
	"fmt"
)

func (c *Client) GetMacs(ctx context.Context) ([]string, error) {
	data, err := c.getTidyStringData(ctx, "/latest/meta-data/network/interfaces/macs/")
	if err != nil {
		return nil, err
	}
	return parsePathNames(data), nil
}

func (c *Client) GetInterfaceIdByMac(ctx context.Context, mac string) (string, error) {
	return c.getTidyStringData(ctx,
		fmt.Sprintf(
			"/latest/meta-data/network/interfaces/macs/%s/network-interface-id",
			mac))
}

func (c *Client) GetNetMaskByMac(ctx context.Context, mac string) (string, error) {
	return c.getTidyStringData(ctx,
		fmt.Sprintf(
			"/latest/meta-data/network/interfaces/macs/%s/netmask",
			mac))
}

func (c *Client) GetVSwitchCidrBlockIdByMac(ctx context.Context, mac string) (string, error) {
	return c.getTidyStringData(ctx,
		fmt.Sprintf(
			"/latest/meta-data/network/interfaces/macs/%s/vswitch-cidr-block",
			mac))
}

func (c *Client) GetPrivateIPV4sByMac(ctx context.Context, mac string) ([]string, error) {
	data, err := c.getRawData(ctx,
		fmt.Sprintf(
			"/latest/meta-data/network/interfaces/macs/%s/private-ipv4s",
			mac))
	if err != nil {
		return nil, err
	}
	return parseJSONStringArray(data)
}

func (c *Client) GetVpcIPV6CidrBlocksByMac(ctx context.Context, mac string) ([]string, error) {
	data, err := c.getRawData(ctx,
		fmt.Sprintf(
			"/latest/meta-data/network/interfaces/macs/%s/vpc-ipv6-cidr-blocks",
			mac))
	if err != nil {
		return nil, err
	}
	return parseJSONStringArray(data)
}

func (c *Client) GetVSwitchIdByMac(ctx context.Context, mac string) (string, error) {
	return c.getTidyStringData(ctx,
		fmt.Sprintf(
			"/latest/meta-data/network/interfaces/macs/%s/vswitch-id",
			mac))
}

func (c *Client) GetVpcIdByMac(ctx context.Context, mac string) (string, error) {
	return c.getTidyStringData(ctx,
		fmt.Sprintf(
			"/latest/meta-data/network/interfaces/macs/%s/vpc-id",
			mac))
}

func (c *Client) GetPrimaryIPAddressByMac(ctx context.Context, mac string) (string, error) {
	return c.getTidyStringData(ctx,
		fmt.Sprintf(
			"/latest/meta-data/network/interfaces/macs/%s/primary-ip-address",
			mac))
}

func (c *Client) GetGatewayByMac(ctx context.Context, mac string) (string, error) {
	return c.getTidyStringData(ctx,
		fmt.Sprintf(
			"/latest/meta-data/network/interfaces/macs/%s/gateway",
			mac))
}

func (c *Client) GetIPV6sByMac(ctx context.Context, mac string) ([]string, error) {
	data, err := c.getRawData(ctx,
		fmt.Sprintf(
			"/latest/meta-data/network/interfaces/macs/%s/ipv6s",
			mac))
	if err != nil {
		return nil, err
	}
	return parseJSONStringArray(data)
}

func (c *Client) GetIPV6GatewayByMac(ctx context.Context, mac string) (string, error) {
	return c.getTidyStringData(ctx,
		fmt.Sprintf(
			"/latest/meta-data/network/interfaces/macs/%s/ipv6-gateway",
			mac))
}

func (c *Client) GetVSwitchIPV6CidrBlockByMac(ctx context.Context, mac string) (string, error) {
	return c.getTidyStringData(ctx,
		fmt.Sprintf(
			"/latest/meta-data/network/interfaces/macs/%s/vswitch-ipv6-cidr-block",
			mac))
}

func (c *Client) GetIPV4PrefixesByMac(ctx context.Context, mac string) (string, error) {
	return c.getTidyStringData(ctx,
		fmt.Sprintf(
			"/latest/meta-data/network/interfaces/macs/%s/ipv4-prefixes",
			mac))
}
