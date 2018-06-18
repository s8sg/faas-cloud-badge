package function

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
		badge = Image["unknown"]
	}

	// Set Etag for http header

	return badge
}

func init() {
	validateCustomers = os.Getenv("validate_customers")
	if validateCustomers == "" {
		validateCustomers = "false"
	}
	customersURL = os.Getenv("customers_url")
}

func getBadge(query url.Values) (string, error) {
	// https://0341c281.ngrok.io/function/s8sg-of_badge_gen?user=s8sg&repo=regex_go&branch=master

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

	commitStatus, cerr := getCommitStatus(user, repo, branch)
	if cerr != nil {
		return "", fmt.Errorf("failed to get commit status, error: %s", cerr.Error())
	}

	image := Image[commitStatus.State]

	return image, nil
}

// TODO: Once openfaas template supports the static file, it can be read locally
//       To avoid delay we made the image static as Image[name][content]
/*
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
}*/

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
