# Doge Test
This package allows for easy integration into the Doge Regtest network.
DogeTest starts Dogecoin daemon in a docker container with temp storage, so each run is a clean run.

# Features
- Starting/Stopping
- Setting up Addresses with Initial Balance
- Getting Address by 'Label'
- Getting Wallet by address (balance etc.)
- Function to generate a confirmed block

# Features to come
- Functions to query the Doge system (i.e. transfers made, inspection of any signatures/scripts)
- Functions to query wallet balances and addresses

# Example Usage
```
dogeTest, err := dogetest.NewDogeTest(dogetest.DogeTestConfig{
    Host:             "localhost",
    Port: 22555,
})
defer dogeTest.Stop()

err = dogeTest.Start()

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

address, err := addressBook.GetAddress("test1")
wallet, err := dogeTest.GetWallet(address.Address)

fmt.Println("Balance for address:", address, wallet.GetBalance())
```

# Example App

`go run cmd/example` 