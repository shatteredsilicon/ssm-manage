package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/shatteredsilicon/ssm-manage/configurator/config"
	"github.com/shatteredsilicon/ssm-manage/configurator/sshkey"
	"github.com/shatteredsilicon/ssm-manage/configurator/user"
)

var c config.PMMConfig
var SSHKey sshkey.Handler

func main() {
	c = config.ParseConfig()
	user.PMMConfig = c
	SSHKey = sshkey.Init(c)
	SSHKey.RunSSHKeyChecks()

	router := mux.NewRouter().PathPrefix(c.PathPrefix).Subrouter()
	router.HandleFunc("/v1/logs", getLogFileHandler).Methods("GET")

	router.HandleFunc("/v1/sshkey", getSSHKeyHandler).Methods("GET")
	router.HandleFunc("/v1/sshkey", setSSHKeyHandler).Methods("POST")

	router.HandleFunc("/v1/check-instance", checkInstanceHandler).Methods("POST")
	router.HandleFunc("/v1/check-update", runCheckUpdateHandler).Methods("GET")
	router.HandleFunc("/v1/version", getCurrentVersionHandler).Methods("GET")
	router.HandleFunc("/v1/updates", getUpdateListHandler).Methods("GET")
	router.HandleFunc("/v1/updates", runUpdateHandler).Methods("POST")
	router.HandleFunc("/v1/updates/{timestamp}", getUpdateHandler).Methods("GET")
	router.HandleFunc("/v1/updates/{timestamp}", deleteUpdateHandler).Methods("DELETE")

	router.HandleFunc("/v2/version", getCurrentVersionHandlerV2).Methods("GET")
	router.HandleFunc("/v2/check-update", runCheckUpdateHandlerV2).Methods("GET")

	router.HandleFunc("/v1/users", getUserListHandler).Methods("GET")
	router.HandleFunc("/v1/users", createUserHandler).Methods("POST")
	router.HandleFunc("/v1/users/{username}", getUserHandler).Methods("GET")
	router.HandleFunc("/v1/users/{username}", deleteUserHandler).Methods("DELETE")

	// TODO: create separate handler with old password verification
	router.HandleFunc("/v1/users/{username}", createUserHandler).Methods("PATCH")

	log.WithFields(log.Fields{
		"address": c.ListenAddress,
	}).Info("SSM Configurator is started")

	hs := &http.Server{Addr: c.ListenAddress, Handler: router}
	gracefulStop(hs)
	if err := hs.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func gracefulStop(hs *http.Server) {
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGINT)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	go func() {
		<-gracefulStop
		if err := hs.Shutdown(context.Background()); err != nil {
			log.Errorf("%v", err)
		}
		log.Info("Server stopped")
	}()
}

func returnSuccess(w io.Writer) {
	json.NewEncoder(w).Encode(jsonResponce{ // nolint: errcheck
		Code:   http.StatusOK,
		Status: http.StatusText(http.StatusOK),
	})
}

func returnError(w http.ResponseWriter, req *http.Request, httpStatus int, title string, err error) {
	response := jsonResponce{
		Code:   httpStatus,
		Status: http.StatusText(httpStatus),
		Title:  title,
	}
	if err != nil {
		response.Detail = err.Error()
	}

	responseJSON, _ := json.Marshal(response)
	responseJSONQuoted := strings.Trim(strconv.Quote(string(responseJSON)), "\"")
	log.Errorf("%s %s: %s", req.Method, req.URL.String(), responseJSONQuoted)

	http.Error(w, string(responseJSON)+"\n", httpStatus)
}
