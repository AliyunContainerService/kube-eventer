package sls

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func (c *Client) CreateStoreView(project string, storeView *StoreView) error {
	body, err := json.Marshal(storeView)
	if err != nil {
		return NewClientError(err)
	}
	h := map[string]string{
		"Content-Type":      "application/json",
		"x-log-bodyrawsize": fmt.Sprintf("%v", len(body)),
	}
	uri := "/storeviews"
	r, err := c.request(project, "POST", uri, h, body)
	if err != nil {
		return err
	}
	r.Body.Close()
	return nil
}

func (c *Client) UpdateStoreView(project string, storeView *StoreView) error {
	body, err := json.Marshal(storeView)
	if err != nil {
		return NewClientError(err)
	}
	h := map[string]string{
		"Content-Type":      "application/json",
		"x-log-bodyrawsize": fmt.Sprintf("%v", len(body)),
	}
	uri := "/storeviews/" + storeView.Name
	r, err := c.request(project, "PUT", uri, h, body)
	if err != nil {
		return err
	}
	r.Body.Close()
	return nil
}

func (c *Client) DeleteStoreView(project string, storeViewName string) error {
	h := map[string]string{
		"Content-Type":      "application/json",
		"x-log-bodyrawsize": "0",
	}
	uri := "/storeviews/" + storeViewName
	r, err := c.request(project, "DELETE", uri, h, nil)
	if err != nil {
		return err
	}
	r.Body.Close()
	return nil
}

func (c *Client) GetStoreView(project string, storeViewName string) (*StoreView, error) {
	h := map[string]string{
		"Content-Type":      "application/json",
		"x-log-bodyrawsize": "0",
	}
	uri := "/storeviews/" + storeViewName
	r, err := c.request(project, "GET", uri, h, nil)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	res := &StoreView{}
	if err = json.Unmarshal(buf, res); err != nil {
		return nil, NewClientError(err)
	}
	res.Name = storeViewName
	return res, nil
}

func (c *Client) ListStoreViews(project string, req *ListStoreViewsRequest) (*ListStoreViewsResponse, error) {
	h := map[string]string{
		"Content-Type":      "application/json",
		"x-log-bodyrawsize": "0",
	}
	uri := fmt.Sprintf("/storeviews?offset=%d&line=%d", req.Offset, req.Size)
	r, err := c.request(project, "GET", uri, h, nil)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	res := &ListStoreViewsResponse{}
	if err = json.Unmarshal(buf, res); err != nil {
		return nil, NewClientError(err)
	}
	return res, nil
}

func (c *Client) GetStoreViewIndex(project string, storeViewName string) (*GetStoreViewIndexResponse, error) {
	h := map[string]string{
		"Content-Type":      "application/json",
		"x-log-bodyrawsize": "0",
	}
	uri := fmt.Sprintf("/storeviews/%s/index", storeViewName)
	r, err := c.request(project, "GET", uri, h, nil)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	res := &GetStoreViewIndexResponse{}
	if err = json.Unmarshal(buf, res); err != nil {
		return nil, NewClientError(err)
	}
	return res, nil
}
