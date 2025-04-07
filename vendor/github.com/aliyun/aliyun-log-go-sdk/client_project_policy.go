package sls

import (
	"fmt"
	"io/ioutil"
)

func (c *Client) GetProjectPolicy(project string) (policy string, err error) {
	h := map[string]string{
		"x-log-bodyrawsize": "0",
		"Content-Type":      "application/json",
	}
	uri := "/policy"
	r, err := c.request(project, "GET", uri, h, nil)
	if err != nil {
		return "", err
	}
	defer r.Body.Close()
	buf, _ := ioutil.ReadAll(r.Body)
	policy = string(buf)
	return policy, nil
}

func (c *Client) UpdateProjectPolicy(project string, policy string) error {
	body := []byte(policy)
	h := map[string]string{
		"x-log-bodyrawsize": fmt.Sprintf("%v", len(body)),
		"Content-Type":      "application/json",
	}
	uri := "/policy"
	r, err := c.request(project, "POST", uri, h, body)
	if err != nil {
		return err
	}
	r.Body.Close()
	return nil
}

func (c *Client) DeleteProjectPolicy(project string) error {
	h := map[string]string{
		"x-log-bodyrawsize": "0",
		"Content-Type":      "application/json",
	}
	uri := "/policy"
	r, err := c.request(project, "DELETE", uri, h, nil)
	if err != nil {
		return err
	}
	r.Body.Close()
	return nil
}
