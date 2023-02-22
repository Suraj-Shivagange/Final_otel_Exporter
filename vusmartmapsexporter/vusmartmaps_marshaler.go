package kafkaexporter

import (
	"encoding/json"
	"time"

	"github.com/Shopify/sarama"
	"github.com/gogo/protobuf/proto"
	"github.com/nqd/flat"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

type vusmartmapsLogsMarshaler struct {
	marshaler plog.Marshaler
	encoding  string
}

func (p vusmartmapsLogsMarshaler) Marshal(ld plog.Logs, topic string) ([]*sarama.ProducerMessage, error) {
	bts, err := p.marshaler.MarshalLogs(ld)
	if err != nil {
		return nil, err
	}
	return []*sarama.ProducerMessage{
		{
			Topic: topic,
			Value: sarama.ByteEncoder(bts),
		},
	}, nil
}

func (p vusmartmapsLogsMarshaler) Encoding() string {
	return p.encoding
}

func newvusmartmapsLogsMarshaler(marshaler plog.Marshaler, encoding string) LogsMarshaler {
	return vusmartmapsLogsMarshaler{
		marshaler: marshaler,
		encoding:  encoding,
	}
}

type vusmartmapsMetricsMarshaler struct {
	marshaler pmetric.Marshaler
	encoding  string
}

func (p vusmartmapsMetricsMarshaler) Marshal(ld pmetric.Metrics, topic string) ([]*sarama.ProducerMessage, error) {
	bts, err := p.marshaler.MarshalMetrics(ld)
	if err != nil {
		return nil, err
	}
	return []*sarama.ProducerMessage{
		{
			Topic: topic,
			Value: sarama.ByteEncoder(bts),
		},
	}, nil
}

func (p vusmartmapsMetricsMarshaler) Encoding() string {
	return p.encoding
}

func newvusmartmapsMetricsMarshaler(marshaler pmetric.Marshaler, encoding string) MetricsMarshaler {
	return vusmartmapsMetricsMarshaler{
		marshaler: marshaler,
		encoding:  encoding,
	}
}

type vusmartmapsTracesMarshaler struct {
	marshaler ptrace.Marshaler
	encoding  string
}

func (p vusmartmapsTracesMarshaler) Marshal(td ptrace.Traces, topic string) ([]*sarama.ProducerMessage, error) {
	var messages []*sarama.ProducerMessage
	// Final list of all the spans
	allSpans := []map[string]interface{}{}

	// Name for specific Kind value in span.Kind
	var Span_SpanKind_name = map[int32]string{
		0: "SPAN_KIND_UNSPECIFIED",
		1: "SPAN_KIND_INTERNAL",
		2: "SPAN_KIND_SERVER",
		3: "SPAN_KIND_CLIENT",
		4: "SPAN_KIND_PRODUCER",
		5: "SPAN_KIND_CONSUMER",
	}

	// Iterate through the ResourceSpans[]
	for i := 0; i < td.ResourceSpans().Len(); i++ {
		// Current ResourceSpan
		resourceSpan := td.ResourceSpans().At(i)
		// Current scopeSpans[]
		scopeSpans := resourceSpan.ScopeSpans()
		// Iterate through the scopeSpans[]
		for j := 0; j < scopeSpans.Len(); j++ {

			// Current spans[]
			spans := scopeSpans.At(j).Spans()
			//Current scope
			scope := scopeSpans.At(j).Scope()
			// Iterate through the spans[]
			for k := 0; k < spans.Len(); k++ {

				// Current span
				span := spans.At(k)

				//traceid,  spanID,  Name,  kind,  parentSpanId as feilds
				traceId := span.TraceID().HexString()
				spanId := span.SpanID().HexString()
				parentSpanId := span.ParentSpanID().HexString()
				span_name := span.Name()
				kind := span.Kind()

				// from struct of Span_SpanKind_name gives the requied name we want from the span.kind int value
				kind_name := proto.EnumName(Span_SpanKind_name, int32(kind))

				//evs := span.Events()

				//span end time and start time informat of data and time // check time.format for other avaliable formats
				span_end_time := span.EndTimestamp().AsTime().Format(time.RFC3339Nano)
				span_start_time := span.StartTimestamp().AsTime().Format(time.RFC3339Nano)

				//start and end time in UnixNano
				endTimeUnixNano := span.EndTimestamp().AsTime().UnixNano()
				startTimeUnixNano := span.StartTimestamp().AsTime().UnixNano()
				durationNano := span.EndTimestamp().AsTime().Nanosecond()

				//Application Name
				applicationname, _ := resourceSpan.Resource().Attributes().Get("service.name")
				application_name := applicationname.AsRaw()

				//  metricset
				metricset := map[string]interface{}{
					"type": "span",
				}

				// resource attributes --> unflattening the keys

				resource := resourceSpan.Resource().Attributes().AsRaw()
				out, _ := flat.Unflatten(resource, &flat.Options{
					Delimiter: ".",
				})

				// resource attributes for the current span

				resourceAttibutes := map[string]interface{}{
					"attributes": out,
				}

				//resourceAttibutes := resourceSpan.Resource().Attributes().AsRaw()

				//Span Details
				spanDetails := map[string]interface{}{
					"parentSpanId":      parentSpanId,
					"name":              span_name,
					"kind":              kind_name,
					"startTimeUnixNano": startTimeUnixNano,
					"endTimeUnixNano":   endTimeUnixNano,
					"durationNano":      durationNano,
					"attributes":        span.Attributes().AsRaw(),
					//"event":             event,
					"status": span.Status(),
				}

				//span Details along with resource attributes and scope
				spanWithAllDetails := map[string]interface{}{
					"application_name": application_name,
					"traceID":          traceId,
					"span_end_time":    span_end_time,
					"spanId":           spanId,
					"span_start_time":  span_start_time,
					"@timestamp":       span_end_time,
					"metricset":        metricset,
					"resource":         resourceAttibutes,
					"scope": map[string]interface{}{
						"name":    scope.Name(),
						"version": scope.Version(),
					},
					"span": spanDetails,
				}
				allSpans = append(allSpans, spanWithAllDetails)
			}
		}
	}

	// Convert to JSON
	outputjson, err := json.Marshal(allSpans)
	if err != nil {
		return nil, err
	}
	// Produce the JSON to Kafka
	messages = append(messages, &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(outputjson),
	})

	return messages, nil

}

func (p vusmartmapsTracesMarshaler) Encoding() string {
	return p.encoding
}

func newvusmartmapsTracesMarshaler(marshaler ptrace.Marshaler, encoding string) TracesMarshaler {
	return vusmartmapsTracesMarshaler{
		marshaler: marshaler,
		encoding:  encoding,
	}
}
