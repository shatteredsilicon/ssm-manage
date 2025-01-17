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

	"github.com/shatteredsilicon/ssm-manage/configurator/config"
	"golang.org/x/mod/semver"
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

var versionRegexp = regexp.MustCompile(`^v?\d+\.\d+\.\d+(-\d+)?$`)

func runCheckUpdateHandler(w http.ResponseWriter, req *http.Request) {
	var version, oriVersion, releaseDate string

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

LOOP:
	for list.Next != nil && *list.Next != "" {
		select {
		case <-timer.C:
			break LOOP
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

			oriVersionStr := string(versionRegexp.Find([]byte(result.Name)))
			if len(oriVersionStr) == 0 {
				continue
			}
			versionStr := oriVersionStr
			if versionStr[0] != 'v' {
				versionStr = "v" + versionStr
			}

			for _, image := range result.Images {
				if image.Architecture != runtime.GOARCH ||
					image.OS != runtime.GOOS ||
					image.Status != statusDockerHubAPIActive {
					continue
				}

				if version == "" || semver.Compare(versionStr, version) > 0 {
					version = versionStr
					oriVersion = oriVersionStr
					releaseDate = image.LastPushed
				}

				break
			}
		}
	}

	ssmVersion := os.Getenv("SSM_VERSION")
	if ssmVersion[0] != 'v' {
		ssmVersion = "v" + ssmVersion
	}

	// compare vMajor.Minor.Patch only
	ssmVersion = ssmVersion[0 : len(ssmVersion)-len(semver.Build(ssmVersion))]
	ssmVersion = ssmVersion[0 : len(ssmVersion)-len(semver.Prerelease(ssmVersion))]
	version = version[0 : len(version)-len(semver.Build(version))]
	version = version[0 : len(version)-len(semver.Prerelease(version))]

	json.NewEncoder(w).Encode(versionResponce{
		Version:      oriVersion,
		ReleaseDate:  releaseDate,
		UpdateNeeded: semver.Compare(version, ssmVersion) > 0,
	})
}

func getCurrentVersionHandler(w http.ResponseWriter, req *http.Request) {
	versionResp := versionResponce{
		Version: os.Getenv("SSM_VERSION"),
	}

	// variable number not exists, use package-level version number
	if strings.TrimSpace(versionResp.Version) == "" {
		versionResp.Version = config.Version
	}

	json.NewEncoder(w).Encode(versionResp)
}
