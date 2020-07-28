package listener

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/tidwall/gjson"
)

// setEnv sets some test envars
func setEnv() {

	os.Setenv("TABLE_NAME", "foo")
	os.Setenv("ISSUE_ID_FIELD", "issue.key")
	os.Setenv("SUMMARY_FIELD", "issue.fields.summary")
	os.Setenv("STATUS_FIELD", "issue.fields.status.name")
	os.Setenv("DESCRIPTION_FIELD", "issue.fields.description")
	os.Setenv("START_TIME_FIELD", "issue.fields.customfield_10109")
	os.Setenv("FINISH_TIME_FIELD", "issue.fields.customfield_10110")
}

//  getMsg gets some test input
func getMsg(p int) (string, error) {

	body, err := ioutil.ReadFile("payloads.json")
	if err != nil {
		return "", err
	}

	path := fmt.Sprintf("cases.%v", p)
	res := gjson.GetManyBytes(body, path)

	return res[0].Raw, nil
}

func TestParseHandler(t *testing.T) {

	tt := []struct {
		name        string
		input       int
		supplierRef string
		status      string
		title       string
		description string
		starts      string
		ends        string
		err         string
	}{
		{name: "good", input: 0, supplierRef: "abc-1", status: "scheduled", title: "foo change",
			description: "\nFor the most up-to-date info, visit /abc-1\nlorem impsum", starts: "2020-09-01 18:30:00", ends: "2020-09-01 19:30:00"},
		{name: "missing", input: 1, err: "missing value in payload"},
		{name: "time", input: 2, err: "cannot parse"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			// init env & object
			setEnv()
			var rec Record

			// create inbound payload
			m, err := getMsg(tc.input)
			if err != nil {
				t.Fatalf("could not get message: %v", err)
			}

			rawM := json.RawMessage(m)
			p, err := json.Marshal(rawM)
			if err != nil {
				t.Fatalf("could not make incoming payload: %v", err)
			}
			pld := bytes.NewReader(p)

			// create inbound request
			r, err := http.NewRequest("POST", "/", pld)
			if err != nil {
				t.Fatalf("could not make incoming request: %v", err)
			}

			// create response recorder
			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(rec.ParseHandler)
			handler.ServeHTTP(rr, r)

			res := rr.Result()
			defer res.Body.Close()

			b, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("could not read response: %v", err)
			}

			if tc.err == "" {
				if res.StatusCode != http.StatusOK {
					t.Errorf("expected status OK, got %v", res.Status)
				}
				if rec.SupplierRef != tc.supplierRef {
					t.Errorf("expected %v, got %v", tc.supplierRef, rec.SupplierRef)
				}
				if rec.Status != tc.status {
					t.Errorf("expected %v, got %v", tc.status, rec.Status)
				}
				if rec.Title != tc.title {
					t.Errorf("expected %v, got %v", tc.title, rec.Title)
				}
				if rec.Description != tc.description {
					t.Errorf("expected %v, got %v", tc.description, rec.Description)
				}
				if rec.Starts != tc.starts {
					t.Errorf("expected %v, got %v", tc.starts, rec.Starts)
				}
				if rec.Ends != tc.ends {
					t.Errorf("expected %v, got %v", tc.ends, rec.Ends)
				}
			}

			if msg := string(bytes.TrimSpace(b)); !strings.Contains(msg, tc.err) {
				t.Errorf("expected error %q, got: %q", tc.err, msg)
			}

		})
	}
}
