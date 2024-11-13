package main

import (
	"fmt"
	"net/http"
	"skillfact/finalproject/svccomments/api"
	"skillfact/finalproject/svccomments/database/inmemory"
)

func main() {
	api := api.New(inmemory.NewDB(""))

	db := api.GetDB()
	db.AddTestingComments(5)

	api.RegistryAPI()
	mux := api.Mux()
	fmt.Println(1)
	fmt.Println(http.ListenAndServe("127.0.0.1:8087", mux))
	fmt.Println(2)
}
