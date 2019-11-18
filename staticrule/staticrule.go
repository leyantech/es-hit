package staticrule

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"

	elastic "gopkg.in/olivere/elastic.v6"
)

// Rule for es query
type Rule struct {
	Name        string `toml:"name"`
	EsURL       string `toml:"es_url"`
	EsUser      string `toml:"es_user"`
	EsPass      string `toml:"es_pass"`
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

	client, err := elastic.NewClient(elastic.SetURL(rule.EsURL),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false),
		elastic.SetBasicAuth(rule.EsUser, rule.EsPass))

	if err != nil {
		return nil, fmt.Errorf("failed to create ElasticSearch Client for %s, %v", rule.EsURL, err)
	}

	return &Wrapper{Client: client, Rule: rule}, nil
}

// SearchHit tell hit number
func (wrapper *Wrapper) SearchHit(ctx context.Context) (int64, error) {
	from := fmt.Sprintf("now-%s", wrapper.Rule.CheckEvery)
	realQ := elastic.NewBoolQuery()

	queryTime := elastic.NewRangeQuery(wrapper.Rule.TimeField).Gte(from).Lte("now")
	realQ = realQ.Filter(queryTime)

	querySearch := elastic.NewQueryStringQuery(wrapper.Rule.Pattern)
	querySearch = querySearch.DefaultField(wrapper.Rule.SearchField)
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
