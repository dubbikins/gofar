package main

import (
	"fmt"
	"net/http"
	"os/exec"

	"github.com/dubbikins/glam/util"
	"github.com/dubbikins/gofar/routers"
	"github.com/shopspring/decimal"
)

// main function
func main() {
	exec.Command("cd ui").Run()
	exec.Command("npm run build")
	exec.Command("cd ..").Run()
	decimal.MarshalJSONWithoutQuotes = true
	router := routers.AppRouter()

	//router.Use(middleware.WithHeader("Access-Control-Allow-Origin", "*"))
	fmt.Println(util.Tree(router, &util.TreeConfig{
		WithColor: true,
	}).String())

	//...setup router
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		fmt.Println(err.Error())
	}

}
