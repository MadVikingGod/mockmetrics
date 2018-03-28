package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

)

var (
	addr = flag.String("listen-address", ":8080", "The address for the http server to listen on")

	currentFlag float64 = 0

	FlagMetric = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "mock_metrics",
			Name:      "user_flag",
			Help:      "User controlled flag.  Settable at /flag",
		},
	)
	SimpleCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "mock_metrics",
			Name:      "simple_counter",
			Help:      "A count that runs about 1 per second",
		},
	)
	SimpleGague = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "mock_metrics",
			Name:      "simple_gague",
			Help:      "A random gague that changes about 1 per second",
		},
	)
	SimpleSummary = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace: "mock_metrics",
			Name:      "simple_summary",
			Help:      "A random summary that updates about 10 per second",
		},
	)
)

func init() {
	prometheus.MustRegister(FlagMetric)
	prometheus.MustRegister(SimpleCounter)
	prometheus.MustRegister(SimpleGague)
	prometheus.MustRegister(SimpleSummary)
}

func updateCounter() {
	for {
		SimpleCounter.Inc()
		time.Sleep(1 * time.Second)
	}
}

func updateGague() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for {

		SimpleGague.Set(r.Float64() * 100)
		time.Sleep(1 * time.Second)
	}

}

func updateSummary() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for {
		for i := 0; i < 10; i++ {
			SimpleSummary.Observe(r.Float64() * 100)
		}
		time.Sleep(1 * time.Second)
	}
}


func updateFlag(w http.ResponseWriter, r *http.Request) {
	var n struct {
		Number float64 `json:"number"`
	}

	if r.Method == "GET" {
		fmt.Fprintf(w, "{\"number\": %f}", currentFlag)
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
		return
	}
	if r.Body == nil {
		http.Error(w, "Please include a number", http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("internal error", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(body, &n)
	if err != nil {
		log.Error("incorrect json", "error", err)
		http.Error(w, "incorrect json", http.StatusBadRequest)
		return
	}

	currentFlag = n.Number
	FlagMetric.Set(currentFlag)

	fmt.Fprintf(w, "{\"success\": ture,\"number\": %f}", currentFlag)
}

var log *zap.SugaredLogger

func main() {
	flag.Parse()
	go updateCounter()
	go updateGague()
	go updateSummary()

	logger := zap.NewExample()
	log = logger.Sugar()
	defer log.Sync()

	FlagMetric.Set(currentFlag)

	http.HandleFunc("/flag", updateFlag)
	http.Handle("/metrics", promhttp.Handler())
	log.Infof("Beging http server", "listen-address", *addr)
	log.Fatalf("http serv error", "error", http.ListenAndServe(*addr, nil))
}
