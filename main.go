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
	Exchangeratesapi      	string 	`json:"exchangeratesapi"`
	Restcountries       	string  `json:"restcountries"`
	Version 				string 	`json:"version"`
	Uptime 					string 	`json:"uptime"`
}

/*
For the response from the restcountries api
*/
type Country []struct {
	Currencies 				[]Currencies 	`json:"currencies"`
	Borders 				[]string 		`json:"borders"`
}

/* For the bordering countries and their currency */
type BorderCountry struct {
	Name 					string 			`json:"name"`
	Currencies				[]Currencies	`json:"currencies"`
}

/* For the currency information of a country */
type Currencies struct {
	Code 					string 			`json:"code"`
	Name 					string 			`json:"name"`
	Symbol 					string 			`json:"symbol"`
}

/* For the output currency data of each bordering country */
type OutCountry struct {
	CountryName 				  	string
	CountryCurrency					string
	CountryRate						float64
}

// For the currency data retrieved from exchangeratesapi
// Contains all possible currencies
type CurrencyData struct {
	Rates struct {
		CAD float64 `json:"CAD"`
		HKD float64 `json:"HKD"`
		ISK float64 `json:"ISK"`
		PHP float64 `json:"PHP"`
		DKK float64 `json:"DKK"`
		HUF float64 `json:"HUF"`
		CZK float64 `json:"CZK"`
		GBP float64 `json:"GBP"`
		RON float64 `json:"RON"`
		SEK float64 `json:"SEK"`
		IDR float64 `json:"IDR"`
		INR float64 `json:"INR"`
		BRL float64 `json:"BRL"`
		RUB float64 `json:"RUB"`
		HRK float64 `json:"HRK"`
		JPY float64 `json:"JPY"`
		THB float64 `json:"THB"`
		CHF float64 `json:"CHF"`
		EUR float64 `json:"EUR"`
		MYR float64 `json:"MYR"`
		BGN float64 `json:"BGN"`
		TRY float64 `json:"TRY"`
		CNY float64 `json:"CNY"`
		NOK float64 `json:"NOK"`
		NZD float64 `json:"NZD"`
		ZAR float64 `json:"ZAR"`
		USD float64 `json:"USD"`
		MXN float64 `json:"MXN"`
		SGD float64 `json:"SGD"`
		AUD float64 `json:"AUD"`
		ILS float64 `json:"ILS"`
		KRW float64 `json:"KRW"`
		PLN float64 `json:"PLN"`
	} `json:"rates"`
	Base string `json:"base"`
	Date string `json:"date"`
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

/* Checks the status code */
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

// Todo Add error handling for status code other than 200 on request to external apis
// Handles the exchange/v1/exchangeborder/ request
func exchangeborder(w http.ResponseWriter, r *http.Request){

	// Modified code based on code retrieved from the "RESTstudent" example at
	//"https://git.gvk.idi.ntnu.no/course/prog2005/prog2005-2021/-/blob/master/RESTstudent/cmd/students_server.go"
	// Retrieves the country name from the url after the trailing slash
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

	// Url request code based on RESTclient found at
	//"https://git.gvk.idi.ntnu.no/course/prog2005/prog2005-2021/-/blob/master/RESTclient/cmd/main.go"
	// URL to invoke
	url := fmt.Sprintf("https://restcountries.eu/rest/v2/name/%s?fields=borders;currencies", name)

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
	// HTTP Header content
	fmt.Println("Status:", res.Status)
	fmt.Println("Status code:", res.StatusCode)

	fmt.Println("Content type:", res.Header.Get("content-type"))
	fmt.Println("Protocol:", res.Proto)
	*/

	// If the http statuscode retrieved from restcountries is not 200 / "OK"
	if res.StatusCode != 200 {
		status := http.StatusNotFound
		http.Error(w, "Error in request to the restcountries api", status)
		return
	}

	// Print output
	output, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Errorf("Error when reading response: ", err.Error())
	}


	//var borderingCountries[5] string

	/* JSON into struct */

	var data Country

	// The currency code of the base currency, used for conversion and the final json output
	//var baseCurrency = data[0].Currencies[0].Name

	err = json.Unmarshal([]byte(string(output)), &data)

	if err != nil {
		// TODO proper error handling
		fmt.Printf("\n ERROR IN UNMARSHAL cancelling")
		return
	}

	var bordercountryCount int

	// Modified code retrieved from "https://stackoverflow.com/questions/22593259/check-if-string-is-int"
	if i, err := strconv.Atoi(limit); err != nil {
		// If it is not an integer
		//fmt.Println(name)
		bordercountryCount = len(data[0].Borders)
	} else {
		// If it is an integer and the integer is less than or equal to the size of the input data
		// then the limit variable is used, else it just uses the full list
		if i <= len(data[0].Borders) {
			bordercountryCount = i
		} else {
			bordercountryCount = len(data[0].Borders)
		}

		//fmt.Println(limit)
		//fmt.Println(name)
	}

	// Array of the country data used for final output, uses the OutCountry struct
	// Size of 16 as that is the maximum amount of bordering countries a country in the world has.
	var outCountries [16]OutCountry
	// For the data from the bordering countries
	var inputData BorderCountry
	// Returns all the currency rates based on the base currency's value
	var currencyData = returnCurrencyVal(data[0].Currencies[0].Name)

	/*
		For loop that goes through up to the first <bordercountryCount> bordering countries
		and sends restcountries requests for their currency information
	*/
	for i := 0; i < bordercountryCount; i++ {
		fmt.Println(i)
		// Get country information from the restcountries api
		// SAMPLE REQUEST FOR TESTING https://restcountries.eu/rest/v2/alpha/FIN?fields=name;currencies
		url := fmt.Sprintf("https://restcountries.eu/rest/v2/alpha/%s?fields=name;currencies", data[0].Borders[i])

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

		if res.StatusCode != 200 {
			status := http.StatusNotFound
			http.Error(w, "Error in request to the restcountries api", status)
			return
		}

		// Print output
		output, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Errorf("Error when reading response: ", err.Error())
		}

		// The currency code of the base currency, used for conversion and the final json output
		//var baseCurrency = data[0].Currencies[0].Name

		err = json.Unmarshal([]byte(string(output)), &inputData)

		if err != nil {
			// TODO proper error handling
			fmt.Printf("\n ERROR IN UNMARSHAL cancelling")
			return
		}

		// Put it into a map used later for currency info as well
		outCountries[i].CountryName = inputData.Name
		outCountries[i].CountryCurrency = inputData.Currencies[0].Code
		outCountries[i].CountryRate = 1.1
	}

	fmt.Println(currencyData)
	fmt.Println(outCountries[1])
	fmt.Println(bordercountryCount)
	fmt.Println(data[0].Borders[0])
	fmt.Printf("\n")
	fmt.Println(data[0].Borders[1])
	fmt.Printf("\n")
	fmt.Println(data[0].Currencies[0])
	fmt.Printf("\n")
	//fmt.Println(data.currencies[0])
	fmt.Printf("\nArray size %i", len(data[0].Borders))

	fmt.Printf("\n Now from the other one \n")

	fmt.Println(string(output))
}

// Gets an input currency code and a base currency and outputs the rate
func returnCurrencyVal(currencyName string) CurrencyData{
	var currencyInfo CurrencyData


	return currencyInfo
}

// Handles the exchange/v1/exchangehistory/ request
func exchangehistory(w http.ResponseWriter, r *http.Request){

	// Modified code based on code retrieved from the "RESTstudent" example at
	//"https://git.gvk.idi.ntnu.no/course/prog2005/prog2005-2021/-/blob/master/RESTstudent/cmd/students_server.go"
	// Retrieves the country name from the url after the trailing slash
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 5 {
		status := http.StatusBadRequest
		http.Error(w, "Expecting format .../exchange/v1/exchangehistory/countryname", status)
		return 	}
	name := parts[4]

	// Checks if the provided country name is empty
	if name == "" {
		status := http.StatusBadRequest
		http.Error(w, "Missing country name. Expecting format .../exchange/v1/exchangehistory/countryname", status)
		return
	}

	var limit = r.FormValue("limit")

	// Modified code retrieved from "https://stackoverflow.com/questions/22593259/check-if-string-is-int"
	if _, err := strconv.Atoi(limit); err != nil {
		// If it is not an integer
		fmt.Println(name)
	} else {
		// If it is an integer
		fmt.Println(limit)
		fmt.Println(name)
	}

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
		http.Error(w, "Expecting format .../exchange/<versionnumber (example: v1)>/diag/", status)
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