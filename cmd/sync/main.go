package main 

import (
	"fmt"
	"github.com/jasmineyas/splitwise-lunchmoney/config"
)

func main(){
	fmt.Println("Hello, Splitwise-LunchMoney Sync!")
	
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	fmt.Printf("Config loaded successfully: %+v\n", cfg)

}