package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
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
	Exchangeratesapi      	string 							`json:"exchangeratesapi"`
	Restcountries       	string  						`json:"restcountries"`
	Version 				string 							`json:"version"`
	Uptime 					string 							`json:"uptime"`
}

// For the response from the restcountries api
type Country []struct {
	Currencies 				[]Currencies 					`json:"currencies"`
	Borders 				[]string 						`json:"borders"`
}

// For the bordering countries and their currency
type BorderCountry struct {
	Name 					string 							`json:"name"`
	Currencies				[]Currencies					`json:"currencies"`
}

// For the currency information of a country, used in the BorderCountry struct
type Currencies struct {
	Code 					string 							`json:"code"`
	Name 					string 							`json:"name"`
	Symbol 					string 							`json:"symbol"`
}

// For the currency data retrieved from exchangeratesapi
type CurrencyData struct {
	Rates 					map[string]float64 				`json:"Rates"`
	Base 					string 							`json:"base"`
	Date 					string 							`json:"date"`
}

// For historical currency data between two dates, used in exchangehistory
type CurrencyRangeRates struct {
	Rates  					map[string]interface {} 		`json:"rates"`
	StartDate 				string							`json:"start_at"`
	BaseCurrency 			string							`json:"base"`
	EndDate 				string							`json:"end_at"`
}

// Returns the uptime of the service based on
// code found at https://stackoverflow.com/questions/37992660/golang-retrieve-application-uptime
func uptime() time.Duration {
	return time.Since(uptimeStart)
}

// Starts the timer for the uptime in exchange/v1/diag/
func init() {
	uptimeStart = time.Now()
}

// Checks the status code
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

// Handles the exchange/v1/exchangeborder/ request
func exchangeborder(w http.ResponseWriter, r *http.Request){

	// For the data from the primary country
	var data Country
	// For the amount of bordering countries to handle
	var bordercountryCount int
	// For the data from the bordering countries
	var inputData BorderCountry
	// Returns all the currency rates based on the base currency's value
	var currencyData CurrencyData
	// Rate of the currency based on the base currency
	var rate float64
	// The text that is output to the user at the end
	var outputText string

	// Start of the output text for the json response
	outputText = `{"rates":{`

	/*
		Modified code based on code retrieved from the "RESTstudent" example at
		"https://git.gvk.idi.ntnu.no/course/prog2005/prog2005-2021/-/blob/master/RESTstudent/cmd/students_server.go"
		Retrieves the country name from the url after the trailing slash
	*/
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 5 {
		status := http.StatusBadRequest
		http.Error(w, "Expecting format .../exchange/v1/exchangeborder/countryname", status)
		return 	}
	name := parts[4]

	// Checks if the provided country name is empty
	if name == "" {
		status := http.StatusBadRequest
		http.Error(w, "Missing country name. Expecting format .../exchange/v1/exchangeborder/countryname", status)
		return
	}

	// Limit from the url of the request "?limit=x"
	var limit = r.FormValue("limit")

	/*
		Url request code based on RESTclient found at
		"https://git.gvk.idi.ntnu.no/course/prog2005/prog2005-2021/-/blob/master/RESTclient/cmd/main.go"
		URL to invoke
	 */
	url := fmt.Sprintf("https://restcountries.eu/rest/v2/name/%s?fields=borders;currencies", name)

	r, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		status := http.StatusInternalServerError
		http.Error(w, "Error in creating request", status)
		return
	}

	// Setting content type -> effect depends on the service provider
	r.Header.Add("content-type", "application/json")
	// Instantiate the client
	client := &http.Client{}

	// Issue request
	res, err := client.Do(r)
	if err != nil {
		status := http.StatusInternalServerError
		http.Error(w, "Error in parsing data", status)
		return
	}

	// If the http statuscode retrieved from restcountries is not 200 / "OK"
	if res.StatusCode != 200 {
		status := http.StatusNotFound
		http.Error(w, "Error in request to the restcountries api, country name provided is probably wrong", status)
		return
	}

	// Reading the data
	output, err := ioutil.ReadAll(res.Body)
	if err != nil {
		status := http.StatusInternalServerError
		http.Error(w, "Error in parsing data", status)
		return
	}

	// JSON into struct
	err = json.Unmarshal(output, &data)

	if err != nil {
		status := http.StatusInternalServerError
		http.Error(w, "Error in parsing data", status)
		return
	}

	// Modified code retrieved from "https://stackoverflow.com/questions/22593259/check-if-string-is-int"
	if i, err := strconv.Atoi(limit); err != nil {
		// If it is not an integer
		bordercountryCount = len(data[0].Borders)
	} else {
		// If it is an integer and the integer is less than or equal to the size of the input data
		// then the limit variable is used, else it just uses the full list
		if i <= len(data[0].Borders) {
			bordercountryCount = i
		} else {
			bordercountryCount = len(data[0].Borders)
		}
	}

	// Get currency information from exchangerates api
	url = fmt.Sprintf("https://api.exchangeratesapi.io/latest?base=%s", data[0].Currencies[0].Code)

	r, err = http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		status := http.StatusInternalServerError
		http.Error(w, "Error in sending request", status)
		return
	}

	// Setting content type -> effect depends on the service provider
	r.Header.Add("content-type", "application/json")
	// Instantiate the client
	client = &http.Client{}

	// Issue request
	res, err = client.Do(r)
	//res, err := client.Get(url) // Alternative: Direct issuing of requests, but fewer configuration options
	if err != nil {
		status := http.StatusInternalServerError
		http.Error(w, "Error in parsing data", status)
		return
	}

	if res.StatusCode != 200 {
		status := http.StatusNotFound
		http.Error(w, "Error in request to the exchangerates api", status)
		return
	}

	// Reading the data
	output, err = ioutil.ReadAll(res.Body)
	if err != nil {
		status := http.StatusInternalServerError
		http.Error(w, "Error in parsing data", status)
		return
	}

	// The currency code of the base currency, used for conversion and the final json output
	err = json.Unmarshal(output, &currencyData)

	if err != nil {
		status := http.StatusInternalServerError
		http.Error(w, "Error in parsing data", status)
		return
	}

	/*
		For loop that goes through up to the first <bordercountryCount> bordering countries
		and sends restcountries requests for their currency information
	*/
	for i := 0; i < bordercountryCount; i++ {
		// Get country information from the restcountries api
		// SAMPLE REQUEST FOR TESTING https://restcountries.eu/rest/v2/alpha/FIN?fields=name;currencies
		url := fmt.Sprintf("https://restcountries.eu/rest/v2/alpha/%s?fields=name;currencies", data[0].Borders[i])

		r, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			status := http.StatusInternalServerError
			http.Error(w, "Error in sending request", status)
			return
		}

		// Setting content type -> effect depends on the service provider
		r.Header.Add("content-type", "application/json")
		// Instantiate the client
		client := &http.Client{}

		// Issue request
		res, err := client.Do(r)
		if err != nil {
			status := http.StatusInternalServerError
			http.Error(w, "Error in parsing data", status)
			return
		}

		if res.StatusCode != 200 {
			status := http.StatusNotFound
			http.Error(w, "Error in request to the restcountries api, please check your country spelling", status)
			return
		}

		// Reading the data
		output, err := ioutil.ReadAll(res.Body)
		if err != nil {
			status := http.StatusInternalServerError
			http.Error(w, "Error in parsing data", status)
			return
		}

		err = json.Unmarshal(output, &inputData)

		if err != nil {
			status := http.StatusInternalServerError
			http.Error(w, "Error in parsing data", status)
			return
		}

		// Gets the rate
		rate = currencyData.Rates[fmt.Sprintf("%s",inputData.Currencies[0].Code)]

		// Check if the country has a currency specified in restcountries, ignore if not
		if inputData.Currencies[0].Code != "" {
		// Format the data for the json output
		outputText = fmt.Sprintf(`%s"%s":{"currency":"%s","rate":"%f"}`,
			outputText,
			inputData.Name,
			inputData.Currencies[0].Code,
			rate)
		}

		// Adds trailing comma if it is not the last country
		if i != bordercountryCount-1 {
			outputText = fmt.Sprintf("%s,", outputText)
		}
	}
	outputText = fmt.Sprintf(`%s},"base":"%s"}`, outputText, data[0].Currencies[0].Code)

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write([]byte(outputText))
	// Error handling with code response
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("500 Internal Server Error"))
		if err != nil {
			status := http.StatusInternalServerError
			http.Error(w, "Error", status)
			return
		}
	}
}

// Handles the exchange/v1/exchangehistory/ request
func exchangehistory(w http.ResponseWriter, r *http.Request){

	// For storing the data retrieved from the restcountries api
	var data Country
	var exchangeHistory CurrencyRangeRates

	// Modified code based on code retrieved from the "RESTstudent" example at
	//"https://git.gvk.idi.ntnu.no/course/prog2005/prog2005-2021/-/blob/master/RESTstudent/cmd/students_server.go"
	// Retrieves the country name from the url after the trailing slash
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 6 {
		status := http.StatusBadRequest
		http.Error(w, "Expecting format .../exchange/v1/{:country_name}/{:begin_date-end_date}", status)
		return 	}
	name := parts[4]

	// Checks if the provided country name is empty
	if name == "" {
		status := http.StatusBadRequest
		http.Error(w, "Missing country name. Expecting format .../{:country_name}/{:begin_date-end_date}", status)
		return
	}

	// Check the start and end dates
	dateparts := strings.Split(parts[5], "-")
	if len(dateparts) != 6 {
		status := http.StatusBadRequest
		http.Error(w, "Date format wrong. Expecting format .../{:country_name}/{:begin_date-end_date}", status)
		return
	}

	// Check if the start date is before or equal to the end date
	if dateparts[0] > dateparts[3] {
		status := http.StatusBadRequest
		http.Error(w, "Date format wrong. Expecting format .../{:country_name}/{:begin_date-end_date}", status)
		return
	} else if dateparts[0] == dateparts[3] {
		if dateparts[1] > dateparts[4] {
			status := http.StatusBadRequest
			http.Error(w, "Date format wrong. Expecting format .../{:country_name}/{:begin_date-end_date}", status)
			return
		} else if dateparts[1] == dateparts[4] && dateparts[2] >= dateparts[5] {
			status := http.StatusBadRequest
			http.Error(w, "Date format wrong. Expecting format .../{:country_name}/{:begin_date-end_date}", status)
			return
		}
	}

	// Puts the start date into a variable
	var startdate = fmt.Sprintf("%s-%s-%s", dateparts[0], dateparts[1], dateparts[2])
	// Puts the end date into a variable
	var enddate = fmt.Sprintf("%s-%s-%s", dateparts[3], dateparts[4], dateparts[5])

	/*
		Request for the country name to restcountries api
		Url request code based on RESTclient found at
		"https://git.gvk.idi.ntnu.no/course/prog2005/prog2005-2021/-/blob/master/RESTclient/cmd/main.go"
		URL to invoke
	 */
	url := fmt.Sprintf("https://restcountries.eu/rest/v2/name/%s?fields=borders;currencies", name)

	r, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		status := http.StatusInternalServerError
		http.Error(w, "Error in sending request", status)
		return
	}

	// Setting content type -> effect depends on the service provider
	r.Header.Add("content-type", "application/json")
	// Instantiate the client
	client := &http.Client{}

	// Issue request
	res, err := client.Do(r)
	if err != nil {
		status := http.StatusInternalServerError
		http.Error(w, "Error in parsing data", status)
		return
	}

	// If the http statuscode retrieved from restcountries is not 200 / "OK"
	if res.StatusCode != 200 {
		status := http.StatusNotFound
		http.Error(w, "Error in request to the restcountries api", status)
		return
	}

	// Print output
	output, err := ioutil.ReadAll(res.Body)
	if err != nil {
		status := http.StatusInternalServerError
		http.Error(w, "Error in parsing data", status)
		return
	}

	// JSON into struct
	err = json.Unmarshal(output, &data)

	if err != nil {
		status := http.StatusNotFound
		http.Error(w, "Error in parsing data from the restcountries api", status)
		return
	}

	// Request for the currency name to exchangerates api
	// Get currency information from exchangerates api
	url = fmt.Sprintf("https://api.exchangeratesapi.io/history?start_at=%s&end_at=%s&symbols=%s",
		startdate ,enddate ,data[0].Currencies[0].Code)

	r, err = http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		status := http.StatusInternalServerError
		http.Error(w, "Error in sending request", status)
		return
	}

	// Setting content type -> effect depends on the service provider
	r.Header.Add("content-type", "application/json")
	// Instantiate the client
	client = &http.Client{}

	// Issue request
	res, err = client.Do(r)
	if err != nil {
		status := http.StatusInternalServerError
		http.Error(w, "Error in parsing data", status)
		return
	}

	if res.StatusCode != 200 {
		status := http.StatusNotFound
		http.Error(w, "Error in request to the exchangerates api", status)
		return
	}

	// Print output
	output, err = ioutil.ReadAll(res.Body)
	if err != nil {
		status := http.StatusInternalServerError
		http.Error(w, "Error in parsing data", status)
		return
	}

	err = json.Unmarshal(output, &exchangeHistory)

	if err != nil {
		status := http.StatusInternalServerError
		http.Error(w, "Error in parsing data", status)
		return
	}

	// Exporting the data
	w.Header().Set("Content-Type", "application/json")
	// Converts the diagnosticData into json
	outData, _ := json.Marshal(exchangeHistory)
	// Writes the json
	_, err = w.Write(outData)
	// Error handling with code response
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("500 Internal Server Error"))
		if err != nil {
			status := http.StatusInternalServerError
			http.Error(w, "500 Internal Server Error", status)
			return
		}
	}
}

// Handles the diag exchange/v1/diag/ request
func diag(w http.ResponseWriter, r *http.Request) {

	// Checks if the method is get, if not sends an error
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, err := w.Write([]byte("405 Method not allowed, please use GET."))
		if err != nil {
			status := http.StatusInternalServerError
			http.Error(w, "500 Internal Server Error", status)
			return
		}
		return
	}

	// Checks if the url is correct
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 5 {
		status := http.StatusBadRequest
		http.Error(w, "Expecting format .../exchange/{:version_number}/diag , example: /exchange/v1/diag", status)
		return
	}

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
	// Error handling with code response
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("500 Internal Server Error"))
		if err != nil {
			status := http.StatusInternalServerError
			http.Error(w, "500 Internal Server Error", status)
			return
		}
	}
}

/*
	Main function, opens the port and sends the requests on to functions that handle them
	// Based on code found at
	https://git.gvk.idi.ntnu.no/course/prog2005/prog2005-2021/-/blob/master/RESTstudent/cmd/students_server.go
*/
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