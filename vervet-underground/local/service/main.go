package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

var versions = []string{"2021-06-01"}

const v20210601 = `{
"openapi": "3.0.0",
"info": {
  "title": "Test Service",
  "version": "0.0.0"
},
"paths": {
  "/test": {
    "get": {
      "operation": "getTest",
      "summary": "Test endpoint",
      "responses": {
        "204": {
          "description": "An empty response"
        }
      }
    }
  }
}}`

func openapiHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	output, err := json.Marshal(versions)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}

	if _, err := w.Write(output); err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}
}

func main() {
	http.HandleFunc("/openapi", openapiHandler)
	http.HandleFunc("/openapi/2021-06-01", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := fmt.Fprintf(w, v20210601); err != nil {
			fmt.Println(err)
			w.WriteHeader(500)
			return
		}
	})
	http.ListenAndServe(":8080", nil)
}
