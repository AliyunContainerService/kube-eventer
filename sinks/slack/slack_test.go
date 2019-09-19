package slack

import (
	"fmt"
	"reflect"
	"testing"
)

func TestSlackInit(t *testing.T) {
	s := &Slack{}
	expectedError := fmt.Errorf(slackErrMsg, "Missing slack token or channel")

	var Tests = []struct {
		slack Slack
		err   error
	}{
		{Slack{Token: "foo", Channel: "bar"}, nil},
		{Slack{Token: "foo"}, expectedError},
		{Slack{Channel: "bar"}, expectedError},
		{Slack{}, expectedError},
	}

	for _, tt := range Tests {
		c := &Config{}
		c.Channel = tt.slack.Channel
		c.Token = tt.slack.Token
		if err := s.Init(c); !reflect.DeepEqual(err, tt.err) {
			t.Fatalf("Init(): %v", err)
		}
	}
}