package kibana

import "testing"

func TestKibana(t *testing.T) {
	var esURL = "http://192.168.1.31:9200"
	kibana := &Kibana{KibanaEsURL: esURL,
		Index:            "logstash-*",
		CheckEvery:       "15m",
		SearchField:      "message",
		TimeField:        "@timestamp",
		ShouldPrefixWith: "rem"}
	wrapper, _ := NewWrapper(kibana)
	wrapper.GenerateRules()
}
