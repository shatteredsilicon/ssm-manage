package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	// maxDockerHubAPIPageSize maximum page size of docker hub API
	maxDockerHubAPIPageSize = 100

	// statusDockerHubAPIActive active status value of docker hub API
	statusDockerHubAPIActive = "active"
)

type dockerHubTag struct {
	ID     uint64 `json:"id"`
	Name   string `json:"name"`
	Images []struct {
		Architecture string `json:"architecture"`
		OS           string `json:"os"`
		Status       string `json:"status"`
		LastPushed   string `json:"last_pushed"`
	} `json:"images"`
	TagStatus     string `json:"tag_status"`
	TagLastPushed string `json:"tag_last_pushed"`
}

type dockerHubTags struct {
	Next     *string        `json:"next"`
	Previous *string        `json:"previous"`
	Results  []dockerHubTag `json:"results"`
}

var versionRegexp = regexp.MustCompile(`^v?\d+\.\d+\.\d+\.\d+\.\d+\.\d+$`)

func runCheckUpdateHandler(w http.ResponseWriter, req *http.Request) {
	var version, releaseDate string

	client := http.Client{}
	timer := time.NewTimer(5 * time.Minute)

	u, err := url.Parse(c.DockerHubRepoAPIPrefix)
	if err != nil {
		returnError(w, req, http.StatusInternalServerError, "Fetching ssm-server version from docker hub failed", err)
		return
	}
	u.Path = path.Join(u.Path, "tags")

	apiURL := u.String()
	list := dockerHubTags{
		Next: &apiURL,
	}
	for list.Next != nil && *list.Next != "" {
		select {
		case <-timer.C:
			break
		default:
		}

		u, err := url.Parse(*list.Next)
		if err != nil {
			returnError(w, req, http.StatusInternalServerError, "Fetching ssm-server version from docker hub failed", err)
			return
		}
		q := u.Query()
		q.Set("page_size", strconv.Itoa(maxDockerHubAPIPageSize))
		u.RawQuery = q.Encode()

		resp, err := client.Get(u.String())
		if err != nil {
			returnError(w, req, http.StatusInternalServerError, "Fetching ssm-server version from docker hub failed", err)
			return
		}

		err = json.NewDecoder(resp.Body).Decode(&list)
		resp.Body.Close()
		if err != nil {
			returnError(w, req, http.StatusInternalServerError, "Fetching ssm-server version from docker hub failed", err)
			return
		}

		for _, result := range list.Results {
			if result.TagStatus != statusDockerHubAPIActive {
				continue
			}

			versionBytes := versionRegexp.Find([]byte(result.Name))
			if versionBytes == nil || len(versionBytes) == 0 {
				continue
			}

			for _, image := range result.Images {
				if image.Architecture != runtime.GOARCH ||
					image.OS != runtime.GOOS ||
					image.Status != statusDockerHubAPIActive {
					continue
				}

				if strings.TrimLeft(string(versionBytes), "v") > strings.TrimLeft(version, "v") &&
					strings.TrimLeft(string(versionBytes), "v") > strings.TrimLeft(os.Getenv("SSM_VERSION"), "v") {
					version = string(versionBytes)
					releaseDate = image.LastPushed
				}

				break
			}
		}
	}

	json.NewEncoder(w).Encode(versionResponce{
		Version:     version,
		ReleaseDate: releaseDate,
	})
}

func getCurrentVersionHandler(w http.ResponseWriter, req *http.Request) {
	versionResp := versionResponce{
		Version: os.Getenv("SSM_VERSION"),
	}

	if versionResp.Version == "" {
		json.NewEncoder(w).Encode(versionResp)
		return
	}

	u, err := url.Parse(c.DockerHubRepoAPIPrefix)
	if err != nil {
		returnError(w, req, http.StatusInternalServerError, "Fetching ssm-server version from docker hub failed", err)
		return
	}
	u.Path = path.Join(u.Path, "tags", os.Getenv("SSM_VERSION"))

	client := http.Client{}
	resp, err := client.Get(u.String())
	if err != nil {
		returnError(w, req, http.StatusInternalServerError, "Fetching ssm-server version from docker hub failed", err)
		return
	}

	var tag dockerHubTag
	err = json.NewDecoder(resp.Body).Decode(&tag)
	defer resp.Body.Close()
	if err != nil {
		returnError(w, req, http.StatusInternalServerError, "Fetching ssm-server version from docker hub failed", err)
		return
	}

	if tag.TagStatus == statusDockerHubAPIActive {
		versionResp.ReleaseDate = tag.TagLastPushed
	}

	json.NewEncoder(w).Encode(versionResp)
}
