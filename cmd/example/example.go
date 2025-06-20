package main

import (
	"fmt"
	"time"

	"github.com/dogecoinfoundation/dogetest/pkg/dogetest"
)

func main() {
	dogeTest, err := dogetest.NewDogeTest(dogetest.DogeTestConfig{
		Host: "localhost",
		Port: 22555,
	})
	if err != nil {
		fmt.Println("Failed to create doge test:", err)
		return
	}

	defer dogeTest.Stop()

	err = dogeTest.Start()
	if err != nil {
		fmt.Println("Failed to start doge test:", err)
		return
	}

	time.Sleep(2 * time.Second)

	addressBook, err := dogeTest.SetupAddresses([]dogetest.AddressSetup{
		{
			Label:          "test1",
			InitialBalance: 100,
		},
		{
			Label:          "test2",
			InitialBalance: 20,
		},
	})
	if err != nil {
		fmt.Println("Failed to setup addresses:", err)
		return
	}

	address, err := addressBook.GetAddress("test1")
	if err != nil {
		fmt.Println("Failed to get address:", err)
		return
	}

	wallet, err := dogeTest.GetWallet(address.Address)
	if err != nil {
		fmt.Println("Failed to get wallet:", err)
		return
	}

	fmt.Println("Balance for address:", address, wallet.GetBalance())

}
