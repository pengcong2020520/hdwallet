package main
import (

	"go_code/hdwallet/cmd"

)



func main() {
	cli := cmd.NewCLI("./data/", "http://localhost:8545", "configs.json")
	cli.Run()
}



