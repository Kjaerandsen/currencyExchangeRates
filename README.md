# Prog2005 Assignment1

Assignment 1 of the Prog2005 course. Showing currency exchange rates of different countries. Written in go.

Contains the following three resource root paths:
/exchange/v1/exchangehistory/
/exchange/v1/exchangeborder/
/exchange/v1/diag/

## Exchange Rate History for Given Currency

The initial endpoint focuses on return the history of exchange rates (against EUR as a fixed base currency) for the currency of a given country, where start and end date are provided. The currency is to be determined based on the country name, and where a country has multiple currencies, only the first one is considered.

### Request

```
Method: GET
Path: exchangehistory/{:country_name}/{:begin_date-end_date}
```

`{:country_name}` refers to the English name for the country as supported by https://restcountries.eu/.

`{:begin_date-end_date}` indicates the begin date (i.e., the earliest date to be reported) of the exchange rate and the end date (i.e., the latest date of the range) of the period over which exchange rates are reported.

Example request: `exchangehistory/norway/2020-12-01-2021-01-31`

### Response

* Content type: `application/json`
* Status code: 200 if everything is OK, appropriate error code otherwise.

Body (Example):
```
{
    "rates": {
        "2020-12-01": {
        "NOK": 1.1969
        },
        "2020-12-02": {
            "NOK": 1.1633
        },
        ...
        "2021-01-31": {
            "NOK": 1.1754
        }
    },
    "start_at": "2020-12-01",
    "base": "EUR",
    "end_at": "2021-01-31"
}
```

## Current Exchange Rate Bordering Countries

The second endpoint provides an overview of the *current exchange rates* of a given country (which is then the base currency) with all bordering countries.

### Request

```
Method: GET
Path: exchangeborder/{:country_name}{?limit={:number}}
```


`{:country_name}` refers to the English name for the country as supported by https://restcountries.eu/.

`{?limit={:number}}` is an optional parameter that limits the number of currencies (`number`) of surrounding countries to be reported.
The limit parameter will be ignored if it does not contain an integer or if it is empty. 

Where countries have multiple currencies, only the first one provided is reported. Where no currency is reported, the country is ignored.

Example request: 
`exchangeborder/norway?limit=5`

### Response

* Content type: `application/json`
* Status code: 200 if everything is OK, appropriate error code otherwise.

Body (Example):
```
{
    "rates": {
        "Sweden": {
            "currency": "SEK", 
            "rate": 1.1703
        },
        "Russia": {
            "currency": "RUB",
            "rate": 72.05
        }, 
    ...
    }
    "base": "NOK"
}
```

## Diagnostics interface

The diagnostics interface indicates the availability of individual services this service depends on. The reporting occurs based on status codes returned by the dependent services, and it further provides information about the uptime of the service.

### Request

```
Method: GET
Path: diag/
```

### Response

* Content type: `application/json`
* Status code: 200 if everything is OK, appropriate error code otherwise. 

Body:
```
{
    "exchangeratesapi": "<http status code for exchangeratesapi API>",
    "restcountries": "<http status code for restcountries API>",
    "version": "v1",
    "uptime": <time in seconds from the last service restart>
}
```

Note: `<some value>` indicates placeholders for values to be populated by the service.

Credits:
The description of the project contains a modified version the assignment instructions to describe the project.
The url request is based on code from RESTclient found at
"https://git.gvk.idi.ntnu.no/course/prog2005/prog2005-2021/-/blob/master/RESTclient/cmd/main.go"
The url parsing is based on a modified version of code in the "RESTstudent" example at 
"https://git.gvk.idi.ntnu.no/course/prog2005/prog2005-2021/-/blob/master/RESTstudent/cmd/students_server.go"
The code to check if a string is an integer was retrieved from 
"https://stackoverflow.com/questions/22593259/check-if-string-is-int"
