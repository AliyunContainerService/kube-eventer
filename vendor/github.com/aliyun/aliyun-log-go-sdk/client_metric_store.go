package sls

import "time"

// CreateMetricStore .
func (c *Client) CreateMetricStore(project string, metricStore *LogStore) error {
	metricStore.TelemetryType = "Metrics"
	err := c.CreateLogStoreV2(project, metricStore)
	if err != nil {
		return err
	}
	time.Sleep(time.Second * 3)
	subStore := &SubStore{}
	subStore.Name = "prom"
	subStore.SortedKeyCount = 2
	subStore.TimeIndex = 2
	subStore.TTL = metricStore.TTL
	subStore.Keys = append(subStore.Keys, SubStoreKey{
		Name: "__name__",
		Type: "text",
	}, SubStoreKey{
		Name: "__labels__",
		Type: "text",
	}, SubStoreKey{
		Name: "__time_nano__",
		Type: "long",
	}, SubStoreKey{
		Name: "__value__",
		Type: "double",
	})
	if !subStore.IsValid() {
		panic("metric store invalid")
	}
	return c.CreateSubStore(project, metricStore.Name, subStore)
}

// UpdateMetricStore .
func (c *Client) UpdateMetricStore(project string, metricStore *LogStore) error {
	metricStore.TelemetryType = "Metrics"
	err := c.UpdateLogStoreV2(project, metricStore)
	if err != nil {
		return err
	}
	return c.UpdateSubStoreTTL(project, metricStore.Name, metricStore.TTL)
}

// DeleteMetricStore .
func (c *Client) DeleteMetricStore(project, name string) error {
	return c.DeleteLogStore(project, name)
}

// GetMetricStore .
func (c *Client) GetMetricStore(project, name string) (*LogStore, error) {
	return c.GetLogStore(project, name)
}
