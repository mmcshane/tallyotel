package bridge

import "go.opentelemetry.io/otel/attribute"

// KVsToTags converts a slice of otel key-value pairs to a tally tag set.
func KVsToTags(kvs []attribute.KeyValue) map[string]string {
	tags := make(map[string]string, len(kvs))
	for _, kv := range kvs {
		tags[string(kv.Key)] = kv.Value.Emit()
	}
	return tags
}
