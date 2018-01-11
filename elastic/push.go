package elastic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"
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
	logger := log.WithFields(log.Fields{
		"url":    url,
		"method": method,
	})
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logger.WithError(err).Errorf("%s: Unable to marshal JSON", logprefix)
		return err
	}

	// TODO Sybren: enable GZip compression.
	req, err := http.NewRequest(method, url.String(), bytes.NewBuffer(payloadBytes))
	if err != nil {
		logger.WithError(err).Errorf("%s: Unable to create request", logprefix)
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	if tweakrequest != nil {
		tweakrequest(req)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.WithError(err).Errorf("%s: error performing HTTP request", logprefix)
		return err
	}

	logger = logger.WithField("code", resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		logger.WithError(err).Errorf("%s: error reading HTTP response body", logprefix)
		return err
	}
	if resp.StatusCode >= 300 {
		suffix := ""
		if resp.StatusCode != 404 {
			suffix = fmt.Sprintf("\n    body:\n%s", body)
		}
		logger.Warningf("%s: error response from Elastic%s", logprefix, suffix)
		return fmt.Errorf("%s: Error %d POSTing to %s", logprefix, resp.StatusCode, url)
	}

	logger.Debug("HTTP request to Elastic OK")
	if responsehandler != nil {
		return responsehandler(resp, body)
	}

	return nil
}

// Push sends the give stats object to ElasticSearch for storage, and returns the document ID.
func Push(elasticURL string, stats interface{}) (string, error) {
	// Figure out the URL to POST to.
	postURL, err := url.Parse(elasticURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %s", err)
	}

	var ID string
	handleResponse := func(resp *http.Response, body []byte) error {
		var response postResponse
		if decodeErr := json.Unmarshal(body, &response); decodeErr != nil {
			return fmt.Errorf("unable to decode JSON: %s", decodeErr)
		}
		log.WithField("ID", response.ID).Info("stored stats in ElasticSearch")
		ID = response.ID

		// Parse the Location header from the response.
		location := resp.Header.Get("Location")
		locURL, urlErr := url.Parse(location)
		if urlErr != nil {
			log.Warningf("Unable to determine absolute URL for location %s", location)
		} else {
			absURL := postURL.ResolveReference(locURL)
			log.WithField("url", absURL.String()).Info("location on Elastic")
		}
		return nil
	}

	// Determine some details about the HTTP request.
	method := "POST"
	url := postURL

	useID := func(strID string) {
		if strID == "" {
			log.Debug("found a pre-existing ID field, but it's empty")
			return
		}
		log.WithField("ID", strID).Debug("found a pre-existing ID field, going to use that")

		method = "PUT"
		url, err = url.Parse(strID)
		if err != nil {
			log.WithFields(log.Fields{
				"_id":      strID,
				"base_url": postURL.String(),
			}).Fatal("unable to construct URL for this _id")
		}
	}

	switch typed := stats.(type) {
	case bson.M:
		// This document comes from MongoDB, and has an ID that we should maintain.
		objectID := typed["_id"]
		switch strID := objectID.(type) {
		case string:
			useID(strID)
		}
		delete(typed, "_id")
		// asJSON, _ := json.MarshalIndent(typed, "", "    ")
		// log.WithFields(log.Fields{
		// 	"method": method,
		// 	"url":    url,
		// }).Debug("writing bson.M object: " + string(asJSON))
	case Stats:
		useID(typed.ID)
		typed.ID = ""
	default:
		log.WithField("payload", stats).Panic("unknown payload type")
	}

	log.WithField("url", url).Debug("Pushing to ElasticSearch")
	err = sendJSON("stats push: ", method, url, stats, nil, handleResponse)
	if err != nil {
		return "", fmt.Errorf("unable to send JSON: %s", err)
	}

	return ID, nil
}
