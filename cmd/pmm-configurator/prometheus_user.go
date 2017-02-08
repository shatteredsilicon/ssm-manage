package main

import (
	"io/ioutil"
	"net/http"
	"strings"
)

func replacePrometheusUser(newUser htuser) error {
	input, err := ioutil.ReadFile(prometheusConfPath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(input), "\n")
	for i, line := range lines {
		if strings.Contains(line, "      username:") {
			lines[i] = "      username: " + newUser.Username
		}
		if strings.Contains(line, "      password:") {
			lines[i] = "      password: " + newUser.Password
		}
	}
	output := strings.Join(lines, "\n")

	if err := ioutil.WriteFile(prometheusConfPath, []byte(output), 0644); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "http://127.0.0.1:9090/prometheus/-/reload", nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	if _, err := client.Do(req); err != nil {
		return err
	}

	return nil
}
