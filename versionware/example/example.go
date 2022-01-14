package example

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// PrintResp prints the response and error from an http client request.  This
// is used in example tests.
func PrintResp(resp *http.Response, err error) {
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(resp.StatusCode, strings.TrimSpace(string(contents)))
}
