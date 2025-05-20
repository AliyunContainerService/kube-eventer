package sls

const (
	EventStoreTelemetryType = "Event"
	EventStoreIndex         = `{
		"max_text_len": 16384,
		"ttl": 7,
		"log_reduce": false,
		"line": {
			"caseSensitive": false,
			"chn": true,
			"token": [",", " ", "'", "\"", ";", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"]
		},
		"keys": {
			"specversion": {
				"type": "text",
				"doc_value": true,
				"alias": "",
				"caseSensitive": false,
				"chn": false,
				"token": [",", " ", "'", "\"", ";", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"]
			},
			"id": {
				"type": "text",
				"doc_value": true,
				"alias": "",
				"caseSensitive": false,
				"chn": false,
				"token": [",", " ", "'", "\"", ";", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"]
			},
			"source": {
				"type": "text",
				"doc_value": true,
				"alias": "",
				"caseSensitive": false,
				"chn": false,
				"token": [",", " ", "'", "\"", ";", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", "\n", "\t", "\r"]
			},
			"type": {
				"type": "text",
				"doc_value": true,
				"alias": "",
				"caseSensitive": false,
				"chn": false,
				"token": [",", " ", "'", "\"", ";", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"]
			},
			"subject": {
				"type": "text",
				"doc_value": true,
				"alias": "",
				"caseSensitive": false,
				"chn": false,
				"token": [",", " ", "'", "\"", ";", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"]
			},
			"datacontenttype": {
				"type": "text",
				"doc_value": true,
				"alias": "",
				"caseSensitive": false,
				"chn": false,
				"token": [",", " ", "'", "\"", ";", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"]
			},
			"dataschema": {
				"type": "text",
				"doc_value": true,
				"alias": "",
				"caseSensitive": false,
				"chn": false,
				"token": [",", " ", "'", "\"", ";", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"]
			},
			"data": {
				"type": "json",
				"doc_value": true,
				"alias": "",
				"caseSensitive": false,
				"chn": false,
				"token": [",", " ", "'", "\"", ";", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"],
				"index_all": true,
				"max_depth": -1,
				"json_keys": {}
			},
			"time": {
				"type": "text",
				"doc_value": true,
				"alias": "",
				"caseSensitive": false,
				"chn": false,
				"token": [",", " ", "'", "\"", ";", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"]
			},
			"title": {
				"type": "text",
				"doc_value": true,
				"alias": "",
				"caseSensitive": false,
				"chn": false,
				"token": [",", " ", "'", "\"", ";", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"]
			},
			"message": {
				"type": "text",
				"doc_value": true,
				"alias": "",
				"caseSensitive": false,
				"chn": false,
				"token": [",", " ", "'", "\"", ";", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"]
			},
			"status": {
				"type": "text",
				"doc_value": true,
				"alias": "",
				"caseSensitive": false,
				"chn": false,
				"token": [",", " ", "'", "\"", ";", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"]
			}
		}
	}`
)

func (c *Client) CreateEventStore(project string, eventStore *LogStore) error {
	eventStore.TelemetryType = EventStoreTelemetryType
	err := c.CreateLogStoreV2(project, eventStore)
	if err != nil {
		return err
	}
	return c.CreateIndexString(project, eventStore.Name, EventStoreIndex)
}

func (c *Client) UpdateEventStore(project string, eventStore *LogStore) error {
	eventStore.TelemetryType = EventStoreTelemetryType
	return c.UpdateLogStoreV2(project, eventStore)
}

func (c *Client) DeleteEventStore(project, name string) error {
	return c.DeleteLogStore(project, name)
}

func (c *Client) GetEventStore(project, name string) (*LogStore, error) {
	return c.GetLogStore(project, name)
}

func (c *Client) ListEventStore(project string, offset, size int) ([]string, error) {
	return c.ListLogStoreV2(project, offset, size, EventStoreTelemetryType)
}
