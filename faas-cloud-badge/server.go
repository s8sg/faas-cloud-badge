package main

import (
	"crypto/sha256"
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
	state             = ""
	Images            = map[string]string{
		"failure": "failure.svg",
		"pending": "pending.svg",
		"success": "success.svg",
		"unknown": "unknown.svg",
	}
	content_type = "image/svg+xml"
	cache_ctl    = "no-cache"
)

func requestHandler(w http.ResponseWriter, r *http.Request) {

	m := r.URL.Query()
	badgeFile, err := getBadge(m)
	if err != nil {
		fmt.Println(err.Error())
		badgeFile = Images["unknown"]
	}

	log.Printf("Serving file: %s", badgeFile)

	// Get the file
	sendFile(w, r, badgeFile)

	// Set Etag for http header
	etag := getEtag(state)

	// Set headers
	w.Header().Set("Content-Type", content_type)
	w.Header().Set("Cache-Control", cache_ctl)
	w.Header().Set("ETag", etag)

	return
}

func initialize() {
	validateCustomers = os.Getenv("validate_customers")
	if validateCustomers == "" {
		validateCustomers = "false"
	}
	customersURL = os.Getenv("customers_url")
}

func getBadge(query url.Values) (string, error) {
	// https://<openfaas>/function/s8sg-of_badge_gen?user=s8sg&repo=regex_go&branch=master

	user := query.Get("user")
	if user == "" {
		return "", fmt.Errorf("github <user> value cant be empty")
	}
	repo := query.Get("repo")
	if repo == "" {
		return "", fmt.Errorf("github <repo> value cant be empty")
	}
	branch := query.Get("branch")
	if branch == "" {
		branch = "master"
	}

	if validateCustomers == "true" && validateUser(user) == false {
		return "", fmt.Errorf("failed to valicate customer: %s", user)
	}

	log.Printf("Getting badge for user '%s', repo '%s' and branch '%s'", user, repo, branch)

	commitStatus, cerr := getCommitStatus(user, repo, branch)
	if cerr != nil {
		return "", fmt.Errorf("failed to get commit status, error: %s", cerr.Error())
	}

	state = commitStatus.State
	log.Printf("State of 'github.com/%s/%s:%s' is '%s'", user, repo, branch, state)
	file, ok := Images[state]
	if !ok {
		return "", fmt.Errorf("invalid state %s", state)
	}

	return file, nil
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

// generate a unique ETag for make cache reloading effective
func getEtag(state string) string {
	h := sha256.New()
	h.Write([]byte(state))
	return fmt.Sprintf("%x", h.Sum(nil))[:31]
}

// Static file request handler
func sendFile(w http.ResponseWriter, r *http.Request, file string) {
	filepath := "/home/app/assets/image/" + file
	http.ServeFile(w, r, filepath)
}

func main() {

	initialize()
	log.Printf("successfully initialized")

	http.HandleFunc("/", requestHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
