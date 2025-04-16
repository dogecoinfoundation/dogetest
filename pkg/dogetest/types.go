package dogetest

import (
	"fmt"

	"dogecoin.org/dogetest/pkg/rpc"
)

type Address struct {
	Address    string
	PrivateKey string
	Label      string
}

type AddressBook struct {
	Addresses []Address
}

type Wallet struct {
	Address  string
	Unspents []rpc.UTXO
}

func (w *Wallet) GetBalance() float64 {
	balance := 0.0
	for _, unspent := range w.Unspents {
		balance += unspent.Amount
	}

	return balance
}

func (a *AddressBook) AddAddress(address Address) {
	a.Addresses = append(a.Addresses, address)
}

func (a *AddressBook) GetAddress(label string) (Address, error) {
	for _, address := range a.Addresses {
		if address.Label == label {
			return address, nil
		}
	}
	return Address{}, fmt.Errorf("address not found")
}
