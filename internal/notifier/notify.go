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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

// cached credentials from SSM (loaded once per cold start)
var snowUser string
var snowPass string

func getSSMParameter(svc *ssm.SSM, name string) (string, error) {
	input := &ssm.GetParameterInput{
		Name:           aws.String(name),
		WithDecryption: aws.Bool(true),
	}
	result, err := svc.GetParameter(input)
	if err != nil {
		return "", err
	}
	return *result.Parameter.Value, nil
}

func loadCredentials() error {
	if snowUser != "" && snowPass != "" {
		return nil
	}

	region := os.Getenv("REGION")
	if region == "" {
		region = "eu-west-2"
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return err
	}

	svc := ssm.New(sess)

	userParam := os.Getenv("SSM_SNOW_USERNAME")
	passParam := os.Getenv("SSM_SNOW_PASSWORD")

	if userParam == "" || passParam == "" {
		return errors.New("missing SSM parameter path environment variables")
	}

	snowUser, err = getSSMParameter(svc, userParam)
	if err != nil {
		return err
	}

	snowPass, err = getSSMParameter(svc, passParam)
	if err != nil {
		return err
	}

	log.Println("SNOW credentials loaded from SSM Parameter Store")
	return nil
}

// Notify calls SNOW API and returns internal_identifier to Handler
func (m *Message) Notify() (string, error) {

	mb, err := json.Marshal(m)
	if err != nil {
		return "", err
	}

	log.Printf("the payload that will be sent: %v", string(mb))

	// put payload bytes in reader and construct request
	mbr := bytes.NewReader(mb)

	// load credentials from SSM
	if err := loadCredentials(); err != nil {
		return "", err
	}

	_, ok := os.LookupEnv("SNOW_URL")
	if !ok {
		return "", errors.New("missing environment variable SNOW_URL")
	}

	u, err := url.Parse(os.Getenv("SNOW_URL"))
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", u.String(), mbr)
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(snowUser, snowPass)
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

	// check for error response from SNOW
	if _, hasErr := dat["error"]; hasErr {
		return "", errors.New("SNOW returned error: " + string(body))
	}

	// check for internal_identifier
	rts, ok := dat["result"].(map[string]interface{})
	if !ok {
		return "", errors.New("unexpected SNOW response format: " + string(body))
	}
	ini, ok := rts["internal_identifier"].(string)
	if !ok || ini == "" {
		return "", errors.New("unexpected or missing internal_identifier in SNOW response: " + string(body))
	}

	// check for response type
	rlg, ok := rts["log"].(string)
	if !ok {
		return "", errors.New("could not parse SNOW response log: " + string(body))
	}

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
