package main

import (
	"net/http"
	"os"
	"time"
)

func main() {
	mode := os.Getenv("MODE")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if mode == "slow" {
			time.Sleep(3 * time.Second)
		}
		if mode == "partial" && r.URL.Path == "/redfish/v1/Systems/1/Storage" {
			http.Error(w, "storage fixture failure", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"@odata.type":"#ServiceRoot.v1_0_0.ServiceRoot","Status":{"Health":"OK"}}`))
	})
	_ = http.ListenAndServe(":8000", nil)
}
