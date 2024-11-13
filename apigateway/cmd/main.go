package main

import (
	"fmt"
	"net/http"
	"skillfact/finalproject/apigateway/api"
)

func main() {
	l := map[string]string{
		"apigateway":  "http://localhost:8085",
		"svcnews":     "http://localhost:8086",
		"svccomments": "http://localhost:8087",
	}
	api := api.New(l, nil)

	api.RegistryAPI()
	mux := api.Mux()
	fmt.Println(1)
	fmt.Println(http.ListenAndServe("localhost:8085", mux))
	fmt.Println(2)
}
