package main

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"io"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type Route struct {
	Path    string
	Handler http.HandlerFunc
}
type Routes []Route

var MyRoutes Routes

func init() {
	viper.SetConfigName("config")                        // name of config file (without extension)
	viper.SetConfigType("yaml")                          // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("/etc/apiNetworkDelayMonitor/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/.apiNetworkDelayMonitor") // call multiple times to add many search paths
	viper.AddConfigPath(".")                             // optionally look for config in the working directory
	err := viper.ReadInConfig()                          // Find and read the config file
	if err != nil {                                      // Handle errors reading the config file
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			panic(fmt.Errorf("Fatal error config file: %s \n", err))
		} else {
			// Config file was found but another error was produced
			panic(fmt.Errorf("Fatal error config file: %s \n", err))
		}

	}
	MyRoutes = Routes{
		Route{
			"/huobi",
			Frame(viper.GetString("apis.huobi")),
		},
		Route{
			"/bian",
			Frame(viper.GetString("apis.bian")),
		},
		Route{
			"/okex",
			Frame(viper.GetString("apis.okex")),
		},
	}
}

func main() {

	for _, r := range MyRoutes {
		http.HandleFunc(r.Path, r.Handler)
	}
	log.Fatal(http.ListenAndServe(":8080", nil))

}

func Frame(url string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatal(err)
		}

		client := http.DefaultClient
		start := time.Now()
		res, err := client.Do(req)
		if err != nil {
			statsFailedResponse := StatsFailedResponse{
				Target: url,
				ErrMsg: err.Error(),
			}
			respBody, _ := json.MarshalIndent(statsFailedResponse, "", "  ")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(respBody)
			log.Error(err)
			return
		}
		end := time.Now()
		duration := end.Sub(start)
		if _, err := io.Copy(ioutil.Discard, res.Body); err != nil {
			log.Fatal(err)
		}
		res.Body.Close()
		statsResponse := StatsResponse{
			Target:       url,
			Milliseconds: duration.Milliseconds(),
		}

		respBody, _ := json.MarshalIndent(statsResponse, "", "  ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(respBody)
	}

}

type StatsResponse struct {
	Target       string `json:"Target"`
	Milliseconds int64  `json:"Milliseconds"`
}
type StatsFailedResponse struct {
	Target string `json:"Target"`
	ErrMsg string `json:"ErrMsg"`
}
