package riemanngo

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"time"

	pb "github.com/golang/protobuf/proto"
	"github.com/riemann/riemann-go-client/proto"
)

// convert an event to a protobuf Event
func EventToProtocolBuffer(event *Event) (*proto.Event, error) {
	if event.Host == "" {
		event.Host, _ = os.Hostname()
	}
	if event.Time.IsZero() {
		event.Time = time.Now()
	}

	var e proto.Event
	e.Host = pb.String(event.Host)
	e.Time = pb.Int64(event.Time.Unix())
	e.TimeMicros = pb.Int64(event.Time.UnixNano() / int64(time.Microsecond))
	if event.Service != "" {
		e.Service = pb.String(event.Service)
	}

	if event.State != "" {
		e.State = pb.String(event.State)
	}
	if event.Description != "" {
		e.Description = pb.String(event.Description)
	}
	e.Tags = event.Tags
	var attrs []*proto.Attribute

	// sort keys
	var keys []string
	for key := range event.Attributes {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		attrs = append(attrs, &proto.Attribute{
			Key:   pb.String(key),
			Value: pb.String(event.Attributes[key]),
		})
	}
	e.Attributes = attrs
	if event.Ttl != 0 {
		e.Ttl = pb.Float32(event.Ttl)
	}

	if event.Metric != nil {
		switch reflect.TypeOf(event.Metric).Kind() {
		case reflect.Int, reflect.Int32, reflect.Int64:
			e.MetricSint64 = pb.Int64(reflect.ValueOf(event.Metric).Int())
		case reflect.Float32:
			e.MetricD = pb.Float64(reflect.ValueOf(event.Metric).Float())
		case reflect.Float64:
			e.MetricD = pb.Float64(reflect.ValueOf(event.Metric).Float())
		case reflect.Uint, reflect.Uint32, reflect.Uint64:
			e.MetricSint64 = pb.Int64(int64(reflect.ValueOf(event.Metric).Uint()))
		default:
			return nil, fmt.Errorf("Metric of invalid type (type %v)",
				reflect.TypeOf(event.Metric).Kind())
		}
	}
	return &e, nil
}

// converts an array of proto.Event to an array of Event
func ProtocolBuffersToEvents(pbEvents []*proto.Event) []Event {
	var events []Event
	for _, event := range pbEvents {
		e := Event{
			State:       event.GetState(),
			Service:     event.GetService(),
			Host:        event.GetHost(),
			Description: event.GetDescription(),
			Ttl:         event.GetTtl(),
			Tags:        event.GetTags(),
		}
		if event.TimeMicros != nil {
			e.Time = time.Unix(0, event.GetTimeMicros()*int64(time.Microsecond))
		} else if event.Time != nil {
			e.Time = time.Unix(event.GetTime(), 0)
		}
		if event.MetricF != nil {
			e.Metric = event.GetMetricF()
		} else if event.MetricD != nil {
			e.Metric = event.GetMetricD()
		} else {
			e.Metric = event.GetMetricSint64()
		}
		if event.Attributes != nil {
			e.Attributes = make(map[string]string, len(event.GetAttributes()))
			for _, attr := range event.GetAttributes() {
				e.Attributes[attr.GetKey()] = attr.GetValue()
			}
		}
		events = append(events, e)
	}
	return events
}
