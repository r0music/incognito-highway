package monitor

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

func StartMonitorServer(port int, timestep time.Duration, monitors []Monitor) {
	// Start all monitors
	for _, m := range monitors {
		m.Start(timestep)
	}

	// Run http server
	m := &poller{Monitors: monitors, reports: map[string]interface{}{}}
	go m.start(timestep)
	http.Handle("/monitor", m)
	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
			// TODO(@0xbunyip): change to logger and prevent os.Exit()
			log.Fatalf("Error in ListenAndServe: %s", err)
		}
	}()
}

func (p *poller) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	js, err := json.Marshal(p.reports)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (p *poller) start(timestep time.Duration) {
	for ; true; <-time.Tick(timestep) {
		reports := map[string]interface{}{}
		for _, m := range p.Monitors {
			name, value, err := m.ReportJSON()
			if err != nil {
				fmt.Println(errors.WithStack(err))
				continue
			}
			reports[name] = value
		}

		p.reports = reports // No need to lock, only save a reference
	}
}

type poller struct {
	Monitors []Monitor

	reports map[string]interface{}
}
