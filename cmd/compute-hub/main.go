package main

import (
	provider "Compute-Hub/Compute-Hub/internal/Network/Provider"
	user "Compute-Hub/Compute-Hub/internal/Network/User"
	"fmt"
)

func main() {
	fmt.Println("Enter P for Provider and U for User:")
	var input string
	fmt.Scanln(&input)

	if input == "P" {
		provider.SetupProvider()
	}
	if input == "U" {
		user.SetupReceiver()
	}
}
