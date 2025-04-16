# Doge Test
This package allows for easy integration into the Doge Regtest network.
Each time the DogeTest starts it cleans up the data from the regtest folder and launches the doge daemon in the background with a randomly generated port.
When the DogeTest library is stoppped it shuts down the Doge process and cleans up the data folder.

**WARNING** This library deletes the `ConfigPath` folder on startup/shutdown, so ensure it is configured to a correct path that you are happy with it deleting.

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
    InstallationPath: "C:\\Program Files\\Dogecoin\\daemon\\dogecoind.exe",
    ConfigPath:       "C:\\Users\\danielw\\AppData\\Roaming\\Dogecoin\\regtest",
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
You may need to configure it to point to the correct path of the config + installation and what host you want to listen on.

`go run cmd/example` 