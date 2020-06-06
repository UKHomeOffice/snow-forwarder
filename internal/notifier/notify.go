package notifier

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// Notify calls SNOW API and returns internal_identifier to Handler
func (m *Message) Notify() (string, error) {

	mb, err := json.Marshal(m)
	if err != nil {
		return "", err
	}

	log.Printf("the payload that will be sent: %v", string(mb))

	// put payload bytes in reader and construct request
	mbr := bytes.NewReader(mb)

	vars := []string{"SNOW_USERNAME", "SNOW_PASSWORD", "SNOW_URL"}
	for _, v := range vars {
		_, ok := os.LookupEnv(v)
		if !ok {
			return "", errors.New("missing environment variable")
		}
	}

	user := os.Getenv("SNOW_USERNAME")
	pass := os.Getenv("SNOW_PASSWORD")

	u, err := url.Parse(os.Getenv("SNOW_URL"))
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", u.String(), mbr)
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(user, pass)
	req.Header.Set("Content-Type", "application/json")

	// call SNOW and log full response
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)

	log.Printf("sent request, SNOW replied with: %v", string(body))

	// dynamically decode SNOW response
	var dat map[string]interface{}
	err = json.Unmarshal(body, &dat)
	if err != nil {
		return "", err
	}

	// check for internal_identifier
	rts := dat["result"].(map[string]interface{})
	ini := rts["internal_identifier"].(string)
	if ini == "" {
		return ini, errors.New("request failed, SNOW did not return a Change ID")
	}

	// check for response type
	rlg := rts["log"].(string)

	if strings.Contains(rlg, "Inserting") {
		log.Printf("SNOW replied with new Change ID: %v", ini)
		return ini, nil
	} else if strings.Contains(rlg, "Updating") {
		log.Printf("SNOW updated Change ID: %v", ini)
		return "", nil
	} else {
		return "", errors.New("could not understand SNOW response")
	}
}
