package elastic

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"
)

type fetchAllQuery struct {
	Size     int      `json:"size,omitempty"`
	Sort     []string `json:"sort,omitempty"`
	Scroll   string   `json:"scroll,omitempty"`
	ScrollID string   `json:"scroll_id,omitempty"`
}

type scrollResponse struct {
	ScrollID string `json:"_scroll_id"`
	Hits     struct {
		Total int   `json:"total"`
		Hits  []Hit `json:"hits"`
	} `json:"hits"`
}

// Hit represents a single search result from Elastic.
type Hit struct {
	Index  string `json:"_index"`
	Type   string `json:"_type"`
	ID     string `json:"_id"`
	Score  string `json:"_score"`
	Source bson.M `json:"_source"` // the document itself
}

var errHTTPError = errors.New("HTTP error communicating with Elastic")

func fetch(client *http.Client, searchURL string, size int, lastScrollID *string) (*scrollResponse, error) {
	logger := log.WithField("url", searchURL)
	payload := fetchAllQuery{}
	if *lastScrollID == "" {
		logger.Info("Performing first request to Elastic")
		payload.Size = size
		payload.Sort = []string{"_doc"} // Scroll requests have optimizations that make them faster when the sort order is _doc.
	} else {
		logger.Info("Performing scroll request to Elastic")
		payload.Scroll = "1m"
		payload.ScrollID = *lastScrollID
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.WithError(err).WithField("payload", payload).Error("Unable to marshal JSON")
		return nil, errHTTPError
	}

	logger.WithField("payload", string(payloadBytes)).Debug("Elastic request payload")
	req, err := http.NewRequest("GET", searchURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		logger.WithError(err).Error("unable to create request")
		return nil, errHTTPError
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.WithError(err).Error("unable to GET from Elastic")
		return nil, errHTTPError
	}
	logger = logger.WithField("code", resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		logger.WithError(err).Error("unable to read response from Elastic")
		return nil, errHTTPError
	}

	if resp.StatusCode >= 300 {
		if resp.StatusCode != 404 {
			logger = logger.WithField("body", string(body))
		}
		logger.WithError(err).Error("received error code from Elastic")
		return nil, errHTTPError
	}
	// logger.Debug("Elastic response: " + string(body))

	var response scrollResponse
	if decodeErr := json.Unmarshal(body, &response); decodeErr != nil {
		logger.WithError(err).WithField("payload", string(body)).Error("unable to decode body")
		return nil, errHTTPError
	}

	*lastScrollID = response.ScrollID

	return &response, nil
}

func deleteScroll(client *http.Client, elasticURL *url.URL, scrollID string) {
	logger := log.WithField("scroll_id", scrollID)
	url, err := elasticURL.Parse("/_search/scroll/" + scrollID)
	if err != nil {
		logger.WithError(err).Warning("unable to construct scroll deletion URL")
		return
	}

	logger = log.WithField("url", url.String())
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		logger.WithError(err).Warning("unable to create request for scroll deletion")
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		logger.WithError(err).Warning("unable to perform request for scroll deletion")
		return
	}
	if resp.StatusCode != http.StatusOK {
		logger.WithField("status_code", resp.StatusCode).Warning("error on request for scroll deletion")
		return
	}
	logger.Debug("ElasticSearch scroll deleted OK")
}

// ReverseImport pulls data from ElasticSearch and sends it to the returned channel.
func ReverseImport(elasticURL string) chan Hit {
	url, err := url.Parse(elasticURL)
	if err != nil {
		log.WithError(err).WithField("url", elasticURL).Fatal("error parsing URL")
	}
	log.WithField("elastic", elasticURL).Warning("reverse-importing from ElasticSearch")

	fetchURL, err := url.Parse("_search?scroll=1m")
	if err != nil {
		log.WithError(err).WithField("url", elasticURL).Fatal("unable to construct fetch URL")
	}
	scrollURL, err := url.Parse("/_search/scroll")
	if err != nil {
		log.WithError(err).WithField("url", elasticURL).Fatal("unable to construct scroll URL")
	}
	requestURL := fetchURL.String()

	client := http.Client{}

	ch := make(chan Hit)
	go func() {
		defer close(ch)

		resultsPerPage := 500
		seenResults := 0
		lastScrollID := ""

		defer func() {
			// Delete the scroll with a DELETE HTTP request.
			if lastScrollID != "" {
				deleteScroll(&client, url, lastScrollID)
			}
		}()

		for {
			resp, err := fetch(&client, requestURL, resultsPerPage, &lastScrollID)
			if err != nil {
				log.Fatal("aborting due to communication error with Elastic")
			}

			for idx := range resp.Hits.Hits {
				ch <- resp.Hits.Hits[idx]
			}

			seenResults += len(resp.Hits.Hits)
			if seenResults >= resp.Hits.Total {
				log.WithFields(log.Fields{
					"seen":  seenResults,
					"total": resp.Hits.Total,
				}).Info("seen all results, stopping import")
				return
			}

			// Reset some things for the 2nd and subsequent requests.
			requestURL = scrollURL.String()
			resultsPerPage = 0
		}
	}()

	return ch
}
