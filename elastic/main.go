package elastic

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// Push sends the give stats object to ElasticSearch for storage.
func Push(stats Stats) error {

	asJSON, err := json.MarshalIndent(&stats, "", "    ")
	if err != nil {
		return fmt.Errorf("unable to marshal to JSON: %s", err)
	}
	os.Stdout.Write(asJSON)
	os.Stdout.Write([]byte("\n"))

	return errors.New("Not implemented yet")
}
