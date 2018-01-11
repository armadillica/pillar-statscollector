package elastic

import (
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"
)

// ResetIndex deletes all Cloud stats from ElasticSearch.
func ResetIndex(elasticURL string) {
	logger := log.WithField("url", elasticURL)

	url, err := url.Parse(elasticURL)
	if err != nil {
		logger.WithError(err).Fatal("unable to parse Elastic URL")
		return
	}

	url, err = url.Parse("..")
	if err != nil {
		logger.WithError(err).Fatal("unable to construct URL to index")
		return
	}
	logger = log.WithField("url", url.String())

	client := &http.Client{}

	req, err := http.NewRequest("DELETE", url.String(), nil)
	if err != nil {
		logger.WithError(err).Fatal("unable to create DELETE request")
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		logger.WithError(err).Fatal("unable to perform DELETE request")
		return
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		logger.WithFields(log.Fields{
			"code": resp.StatusCode,
			"text": resp.Status,
		}).Fatal("error from DELETE request")
		return
	}
	logger.WithField("code", resp.StatusCode).Info("ElasticSearch index deleted OK")
}
