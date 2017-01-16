package resources

import "net/http"

// API returns the swagger.yml for this service.
func API(api []byte) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/vnd.yaml")
		w.Write(api)
	}
}
