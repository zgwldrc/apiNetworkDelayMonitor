package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type GaugeFuncEntry struct {
	Name    string
	GaugeFunc func() float64
}

var GaugeFuncEntrys []GaugeFuncEntry

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
	GaugeFuncEntrys = []GaugeFuncEntry{
		{
			"network_delay_huobi",
			Frame(viper.GetString("apis.huobi")),
		},
		{
			"network_delay_bian",
			Frame(viper.GetString("apis.bian")),
		},
		{
			"network_delay_okex",
			Frame(viper.GetString("apis.okex")),
		},
	}
}

func main() {

	for _, e := range GaugeFuncEntrys {
		if err := prometheus.Register(prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name:      e.Name,
				Help:      "",
			},
			e.GaugeFunc,
		)); err == nil {
			fmt.Printf("GaugeFunc '%s' registered.\n", e.Name)
		}
	}
	http.Handle("/metrics", promhttp.Handler())

	log.Fatal(http.ListenAndServe(":8080", nil))

}

func Frame(url string) func() float64 {
	return func() float64 {
		start := time.Now()
		res, err := http.Get(url)
		if err != nil {
			log.Errorf("get %s error: %s", url, err)
			return -1
		}
		end := time.Now()
		duration := end.Sub(start)
		if _, err := io.Copy(ioutil.Discard, res.Body); err != nil {
			log.Errorf("get %s error: %s", url, err)
			return -1
		}
		res.Body.Close()
		return float64(duration.Milliseconds())
	}
}
