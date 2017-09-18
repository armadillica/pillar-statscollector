package elastic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"
)

// sendJSON sends a JSON document to some URL via HTTP.
// :param tweakrequest: can be used to tweak the request before sending it, for
//    example by adding authentication headers. May be nil.
// :param responsehandler: is called when a non-error response has been read.
//    May be nil.
func sendJSON(logprefix, method string, url *url.URL,
	payload interface{},
	tweakrequest func(req *http.Request),
	responsehandler func(resp *http.Response, body []byte) error,
) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Errorf("%s: Unable to marshal JSON: %s", logprefix, err)
		return err
	}

	// TODO Sybren: enable GZip compression.
	req, err := http.NewRequest("POST", url.String(), bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Errorf("%s: Unable to create request: %s", logprefix, err)
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	if tweakrequest != nil {
		tweakrequest(req)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Warningf("%s: Unable to POST to %s: %s", logprefix, url, err)
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Warningf("%s: Error %d POSTing to %s: %s",
			logprefix, resp.StatusCode, url, err)
		return err
	}

	if resp.StatusCode >= 300 {
		suffix := ""
		if resp.StatusCode != 404 {
			suffix = fmt.Sprintf("\n    body:\n%s", body)
		}
		log.Warningf("%s: Error %d POSTing to %s%s",
			logprefix, resp.StatusCode, url, suffix)
		return fmt.Errorf("%s: Error %d POSTing to %s", logprefix, resp.StatusCode, url)
	}

	if responsehandler != nil {
		return responsehandler(resp, body)
	}

	return nil
}

// Push sends the give stats object to ElasticSearch for storage.
func Push(elasticURL string, stats interface{}) error {
	// Figure out the URL to POST to.
	postURL, err := url.Parse(elasticURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %s", err)
	}

	handleResponse := func(resp *http.Response, body []byte) error {
		var response postResponse
		if decodeErr := json.Unmarshal(body, &response); decodeErr != nil {
			return fmt.Errorf("unable to decode JSON: %s", decodeErr)
		}
		log.Infof("Stats stored with id=%q", response.ID)

		// Parse the Location header from the response.
		location := resp.Header.Get("Location")
		locURL, urlErr := url.Parse(location)
		if urlErr != nil {
			log.Warningf("Unable to determine absolute URL for location %s", location)
		} else {
			absURL := postURL.ResolveReference(locURL)
			log.Infof("Location: %s", absURL.String())
		}
		return nil
	}

	log.Infof("Pushing to ElasticSearch at %s", postURL)
	err = sendJSON("stats push: ", "POST", postURL, stats, nil, handleResponse)
	if err != nil {
		return fmt.Errorf("unable to send JSON: %s", err)
	}

	return nil
}
