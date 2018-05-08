package es

import (
	"context"
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"

	elastic "gopkg.in/olivere/elastic.v5"
)

// Rule for es query rule
type Rule struct {
	Name        string `toml:"name"`
	EsURL       string `toml:"es_url"`
	Index       string `toml:"index"`
	CheckEvery  string `toml:"check_every"`
	SearchField string `toml:"search_field"`
	TimeField   string `toml:"time_field"`
	Pattern     string `toml:"pattern"`
	Enabled     bool   `toml:"enabled"`
}

// Wrapper is main structure for query target elasticsearch
type Wrapper struct {
	Client *elastic.Client
	Rule   *Rule
}

// Log is the match hit entry
type Log struct {
	Message string `json:"message"`
}

// NewWrapper create new ES instance
func NewWrapper(rule *Rule) (*Wrapper, error) {
	client, err := elastic.NewClient(elastic.SetURL(rule.EsURL))

	if err != nil {
		return nil, fmt.Errorf("failed to create ElasticSearch Client for %s, %v", rule.EsURL, err)
	}

	return &Wrapper{Client: client, Rule: rule}, nil
}

// SearchHit tell whether have hit or not
func (wrapper *Wrapper) SearchHit(ctx context.Context) (int64, error) {
	from := fmt.Sprintf("now-%s", wrapper.Rule.CheckEvery)
	realQ := elastic.NewBoolQuery()

	queryTime := elastic.NewRangeQuery(wrapper.Rule.TimeField).Gte(from).Lte("now")
	log.Debugln(jsonOfRangeQuery(queryTime))
	realQ = realQ.Filter(queryTime)

	querySearch := elastic.NewQueryStringQuery(wrapper.Rule.Pattern)
	querySearch = querySearch.DefaultField(wrapper.Rule.SearchField)
	log.Debugln(jsonOfStringQuery(querySearch))
	realQ = realQ.Must(querySearch)

	searchResult, err := wrapper.Client.Search().
		Index(wrapper.Rule.Index).
		Query(realQ).
		Do(ctx)
	if err != nil {
		log.Errorf("Failed to search %v", err)
		return -1, err
	}

	hitsCount := searchResult.TotalHits()
	log.Debugf("Query took %d milliseconds found %d log match hit", searchResult.TookInMillis, hitsCount)

	return hitsCount, nil
}

func jsonOfStringQuery(q *elastic.QueryStringQuery) string {
	src, err := q.Source()
	if err != nil {
		log.Error(err)
		return ""
	}
	data, err := json.MarshalIndent(src, "", " ")
	if err != nil {
		log.Errorf("marshaling to JSON failed: %v", err)
		return ""
	}
	return string(data)
}

func jsonOfRangeQuery(q *elastic.RangeQuery) string {
	src, err := q.Source()
	if err != nil {
		log.Error(err)
		return ""
	}
	data, err := json.MarshalIndent(src, "", " ")
	if err != nil {
		log.Errorf("marshaling to JSON failed: %v", err)
		return ""
	}
	return string(data)
}

func printSearchHits(hits *elastic.SearchHits, outputLimit int64) {
	total := len(hits.Hits)
	// Reverse order print, latest on top
	for i := 1; int64(i) <= outputLimit && i <= total; i++ {
		var logEntry Log
		jsonResult, _ := hits.Hits[total-i].Source.MarshalJSON()
		json.Unmarshal(jsonResult, &logEntry)
		fmt.Println(logEntry.Message)
	}
}
