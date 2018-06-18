package function

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var (
	validateCustomers = "false"
	customersURL      = ""
	user              = ""
	repo              = ""
	ImageUrls         = map[string]string{
		"failure_BUILD.svg":  "https://img.shields.io/badge/openfaas--cloud-build%20fail-red.svg",
		"failure_DEPLOY.svg": "https://img.shields.io/badge/openfaas--cloud-deploy%20fail-red.svg",
		"pending_BUILD.svg":  "https://img.shields.io/badge/openfaas--cloud-build%20pending-yellow.svg",
		"pending_DEPLOY.svg": "https://img.shields.io/badge/openfaas--cloud-deploy%20pending-yellow.svg",
		"success_DEPLOY.svg": "https://img.shields.io/badge/openfaas--cloud-deployed-green.svg",
		"unknown.svg":        "https://img.shields.io/badge/openfaas--cloud-unknown-lightgrey.svg",
	}
)

// Handle a serverless request
func Handle(req []byte) string {
	m, perr := url.ParseQuery(os.Getenv("Http_Query"))
	if perr != nil {
		log.Fatalf("failed to query, error : %s", perr.Error())
	}
	badge, err := getBadge(m)
	if err != nil {
		fmt.Println(err.Error())
		badge, _ = getImage(ImageUrls["unknown.svg"])
	}
	return string(badge)
}

func init() {
	validateCustomers = os.Getenv("validate_customers")
	if validateCustomers == "" {
		validateCustomers = "false"
	}
	customersURL = os.Getenv("customers_url")
}

func getBadge(query url.Values) ([]byte, error) {
	// https://0341c281.ngrok.io/function/s8sg-of_badge_gen?user=s8sg&repo=regex_go&branch=master

	user := query.Get("user")
	if user == "" {
		return nil, fmt.Errorf("github <user> value cant be empty")
	}
	repo := query.Get("repo")
	if repo == "" {
		return nil, fmt.Errorf("github <repo> value cant be empty")
	}
	branch := query.Get("branch")
	if branch == "" {
		branch = "master"
	}

	if validateCustomers == "true" && validateUser(user) == false {
		return nil, fmt.Errorf("failed to valicate customer: %s", user)
	}

	commitStatus, cerr := getCommitStatus(user, repo, branch)
	if cerr != nil {
		return nil, fmt.Errorf("failed to get commit status, error: %s", cerr.Error())
	}

	imageUrl := ImageUrls[fmt.Sprintf("%s_%s.svg", commitStatus.State, commitStatus.Statuses[0].Context)]

	return getImage(imageUrl)
}

func getImage(imageUrl string) ([]byte, error) {
	c := http.Client{}

	httpReq, _ := http.NewRequest(http.MethodGet, imageUrl, nil)
	res, reqErr := c.Do(httpReq)
	if reqErr != nil {
		return nil, fmt.Errorf("failed to get image file %s, error: %s", imageUrl, reqErr.Error())
	}
	if res.StatusCode > 299 || res.StatusCode < 200 {
		return nil, fmt.Errorf("failed to get image file %s, response code: %d", imageUrl, res.StatusCode)
	}

	var imageContent []byte = nil
	if res.Body != nil {
		defer res.Body.Close()

		imageContent, _ = ioutil.ReadAll(res.Body)
	} else {
		return nil, fmt.Errorf("failed to get image file %s", imageUrl)
	}

	return imageContent, nil
}

func validateUser(user string) bool {
	found := false
	customers, getErr := getCustomers(customersURL)
	if getErr != nil {
		fmt.Println("failed to get customer list, error :", getErr.Error())
		return false
	}

	for _, customer := range customers {
		if customer == user {
			found = true
		}
	}
	return found
}

// getCustomers reads a list of customers separated by new lines
// who are valid users of OpenFaaS cloud
func getCustomers(customerURL string) ([]string, error) {
	customers := []string{}

	c := http.Client{}

	httpReq, _ := http.NewRequest(http.MethodGet, customerURL, nil)
	res, reqErr := c.Do(httpReq)

	if reqErr != nil {
		return customers, reqErr
	}

	if res.Body != nil {
		defer res.Body.Close()

		pageBody, _ := ioutil.ReadAll(res.Body)
		customers = strings.Split(string(pageBody), "\n")
	}

	return customers, nil
}
