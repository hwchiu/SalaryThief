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
		if mode == "inventory-partial" && r.URL.Path == "/redfish/v1/Systems/1/Inventory" {
			http.Error(w, "inventory fixture failure", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/redfish/v1/Systems/1/Inventory" {
			if mode == "inventory-replaced" {
				_, _ = w.Write([]byte(`{"components":[{"type":"drive","component_id":"drive-03","location":"Bay 03","serial":"BBB","firmware":"2.0","health":1},{"type":"dimm","component_id":"dimm-a1","location":"DIMM A1","serial":"DIMM001","health":1}]}`))
				return
			}
			_, _ = w.Write([]byte(`{"components":[{"type":"storage_controller","component_id":"ctrl-0","location":"Controller 0","health":1},{"type":"enclosure","component_id":"enc-1","location":"Enclosure 1","health":1},{"type":"drive","component_id":"drive-03","location":"Bay 03","serial":"AAA","firmware":"1.0","health":1},{"type":"dimm","component_id":"dimm-a1","location":"DIMM A1","serial":"DIMM001","health":1},{"type":"fan","component_id":"fan-1","location":"Fan 1","health":1}]}`))
			return
		}
		_, _ = w.Write([]byte(`{"@odata.type":"#ServiceRoot.v1_0_0.ServiceRoot","Status":{"Health":"OK"}}`))
	})
	_ = http.ListenAndServe(":8000", nil)
}
