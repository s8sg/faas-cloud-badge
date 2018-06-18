package function

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type CommitStatus struct {
	State    string `json:"state"`
	Statuses []struct {
		Context string `json:"context"`
	}
}

func getCommitStatus(user string, repo string, branch string) (CommitStatus, error) {
	c := http.Client{}
	commitStatus := CommitStatus{}
	url := "https://api.github.com/repos/" + user + "/" + repo + "/commits/" + branch + "/status"
	httpreq, _ := http.NewRequest(http.MethodGet, url, nil)
	res, reqErr := c.Do(httpreq)
	if reqErr != nil {
		return commitStatus, reqErr
	}
	if res.Body == nil {
		return commitStatus, fmt.Errorf("response is blank")
	}
	defer res.Body.Close()
	resBody, _ := ioutil.ReadAll(res.Body)
	err := json.Unmarshal(resBody, &commitStatus)
	if err != nil {
		return commitStatus, err
	}
	return commitStatus, nil
}
