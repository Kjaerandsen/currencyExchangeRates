package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var uptimeStart time.Time // Uptime calculation start value

/*
diag represents the diagnostic feedback structure.
It is of the form:
{
	"exchangeratesapi": 	<http status code for exchangeratesapi>, 	e.g. 200
	"restcountries": 		<http status code for restcountries>		e.g. 200
	"version":				<versionnumber>								e.g. "v1"
    "uptime": 				<uptime in seconds>							e.g. 1200
}
*/
type Diagnostic struct {
	Exchangeratesapi      	string 	`json:"exchangeratesapi"`
	Restcountries       	string  `json:"restcountries"`
	Version 				string 	`json:"version"`
	Uptime 					string 	`json:"uptime"`
}


type Currencies struct {
	Code string 	`json:"currency"`
	Name string 	`json:"name"`
	Symbol string 	`json:"symbol"`
}

/*
For the response from the restcountries api
*/
type Country struct {
	currencies []Currencies
	borders []string `json:"border"`
}

// Returns the uptime of the service
func uptime() time.Duration {
	return time.Since(uptimeStart)
}

// Starts the timer for the uptime in exchange/v1/diag/
func init() {
	uptimeStart = time.Now()
}

func checkStatus(webUrl string) int {
	// Get http status codes
	resp, err := http.Head(webUrl)
	if err != nil {
		log.Fatal(err)
	}
	return resp.StatusCode
}

// Redirects urls missing the trailing slash
func redirect(w http.ResponseWriter, r *http.Request){
	http.Redirect(w, r, r.URL.Path + "/", http.StatusSeeOther)
}

// Handles the diag exchange/v1/diag/ request
func diag(w http.ResponseWriter, r *http.Request) {

	// Checks if the method is get, if not sends an error
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, err := w.Write([]byte("405 Method not allowed, please use GET."))
		if err != nil {
			// TODO add error message here
		}
		return
	}

	// Checks if the url is correct
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 5 {
		status := http.StatusBadRequest
		http.Error(w, "Expecting format .../exchange/v<versionnumber>/diag/", status)
		return
	}


	// Prints the diagnostics

	//var Timeupdated string
	//Timeupdated = strconv.FormatFloat(uptime().Seconds(), 'f', 6, 64)
	//Timeupdated = fmt.Sprintf("%ss", Timeupdated)


	// Creates the diagnostic information
	w.Header().Set("Content-Type", "application/json")
	diagnosticData := &Diagnostic{
		Exchangeratesapi: fmt.Sprintf("%d", checkStatus("https://api.exchangeratesapi.io/latest")),
		Restcountries: fmt.Sprintf("%d", checkStatus("https://restcountries.eu/rest/v2/")),
		Version: "v1",
		Uptime: fmt.Sprintf("%ds", int(uptime().Seconds())),
	}

	// Converts the diagnosticData into json
	data, _ := json.Marshal(diagnosticData)
	// Writes the json
	_, err := w.Write(data)
	//_, err := w.Write([]byte(string(data)))
	// Error handling with code response
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("500 Internal Server Error"))
		if err != nil {
			// TODO add error message here
		}
	}
}


// Todo Add error handling for status code other than 200 on request to external apis
// Handles the exchange/v1/exchangeborder/ request
func exchangeborder(w http.ResponseWriter, r *http.Request){

	// URL to invoke
	url := "https://restcountries.eu/rest/v2/name/norway"

	r, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		fmt.Errorf("Error in creating request:", err.Error())
	}

	// Setting content type -> effect depends on the service provider
	r.Header.Add("content-type", "application/json")
	// Instantiate the client
	client := &http.Client{}

	// Issue request
	res, err := client.Do(r)
	//res, err := client.Get(url) // Alternative: Direct issuing of requests, but fewer configuration options
	if err != nil {
		fmt.Errorf("Error in response:", err.Error())
	}

	/*
	w.WriteHeader(http.StatusNotImplemented)
	_, err := w.Write([]byte("501 Not Implemented"))
	if err != nil {
		// TODO add error message here
	}
	 */
}

// Handles the exchange/v1/exchangehistory/ request
func exchangehistory(w http.ResponseWriter, r *http.Request){
	w.WriteHeader(http.StatusNotImplemented)
	_, err := w.Write([]byte("501 Not Implemented"))
	if err != nil {
		// TODO add error message here
	}
}

// Main function, opens the port and sends the requests on
func main() {
	// Sets up the port of the application to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// http request handlers
	http.HandleFunc("/exchange/v1/diag/", diag)
	http.HandleFunc("/exchange/v1/exchangehistory/", exchangehistory)
	http.HandleFunc("/exchange/v1/exchangeborder/", exchangeborder)

	// redirect if missing the trailing slash
	http.HandleFunc("/exchange/v1/diag", redirect)
	http.HandleFunc("/exchange/v1/exchangehistory", redirect)
	http.HandleFunc("/exchange/v1/exchangeborder", redirect)

	fmt.Println("Listening on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}