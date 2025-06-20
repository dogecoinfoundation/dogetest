package dogetest

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/dogecoinfoundation/dogetest/pkg/rpc"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

//go:embed Dockerfile.dogecoin
var dockerfileData []byte

type DogeTest struct {
	Host      string
	Rpc       *rpc.RpcTransport
	config    DogeTestConfig
	Container testcontainers.Container
}

type DogeTestConfig struct {
	Host          string
	NetworkName   string
	LogContainers bool
	Port          int
}

type AddressSetup struct {
	Label          string
	InitialBalance int
}

func (d *DogeTest) SetupAddresses(addressSetups []AddressSetup) (*AddressBook, error) {
	addresses := make([]Address, len(addressSetups))

	_, err := d.Rpc.Generate(100)
	if err != nil {
		return nil, err
	}

	for i, addressSetup := range addressSetups {
		address, err := d.Rpc.GetNewAddress()
		if err != nil {
			return nil, err
		}

		privKey, err := d.Rpc.DumpPrivKey(address)
		if err != nil {
			return nil, err
		}

		err = d.Rpc.SendToAddress(address, float64(addressSetup.InitialBalance))
		if err != nil {
			return nil, err
		}

		addresses[i] = Address{
			Label:      addressSetup.Label,
			Address:    address,
			PrivateKey: privKey,
		}
	}

	blocks, err := d.ConfirmBlocks()
	if err != nil {
		return nil, err
	}

	return &AddressBook{
		Addresses: addresses,
		Blocks:    blocks,
	}, nil
}

func NewDogeTest(config DogeTestConfig) (*DogeTest, error) {
	return &DogeTest{
		config: config,
	}, nil
}

func (d *DogeTest) GetWallet(address string) (*Wallet, error) {
	unspents, err := d.Rpc.ListUnspent(address)
	if err != nil {
		return nil, err
	}

	return &Wallet{
		Address:  address,
		Unspents: unspents,
	}, nil
}

func (d *DogeTest) ConfirmBlocks() ([]string, error) {
	blocks, err := d.Rpc.Generate(1)
	if err != nil {
		return nil, err
	}

	return blocks, nil
}

func (d *DogeTest) Start() error {
	portVal := strconv.Itoa(d.config.Port)

	logConsumers := []testcontainers.LogConsumer{}
	if d.config.LogContainers {
		logConsumer := &StdoutLogConsumer{
			Name: "dogecoin",
		}

		logConsumers = append(logConsumers, logConsumer)
	}

	networks := []string{}
	if d.config.NetworkName != "" {
		networks = append(networks, d.config.NetworkName)
	} else {
		ctx := context.Background()

		net, err := network.New(ctx, network.WithDriver("bridge"))
		if err != nil {
			return err
		}
		networks = append(networks, net.Name)
	}

	dockerfilePath, err := WriteDockerfileToDisk()
	if err != nil {
		return err
	}

	dockerFolderPath := filepath.Dir(dockerfilePath)
	dockerFilename := filepath.Base(dockerfilePath)

	log.Println("Dockerfile path:", dockerfilePath)
	log.Println("Dockerfile folder path:", dockerFolderPath)
	log.Println("Dockerfile filename:", dockerFilename)

	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    dockerFolderPath,
			Dockerfile: dockerFilename,
			KeepImage:  false,
			BuildArgs: map[string]*string{
				"PORT": &portVal,
			},
		},
		Networks:     networks,
		Name:         "dogecoin-" + portVal,
		ExposedPorts: []string{portVal + "/tcp"},
		Env: map[string]string{
			"PORT": portVal,
		},
		WaitingFor: wait.ForLog("init message: Done loading").WithStartupTimeout(10 * time.Second),
		LogConsumerCfg: &testcontainers.LogConsumerConfig{
			Opts:      []testcontainers.LogProductionOption{testcontainers.WithLogProductionTimeout(10 * time.Second)},
			Consumers: logConsumers,
		},
	}

	ctx := context.Background()

	dogecoinContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return err
	}

	for {
		if dogecoinContainer.IsRunning() {
			break
		}

		time.Sleep(1 * time.Second)
	}

	for {
		handlerPort, err := dogecoinContainer.MappedPort(ctx, nat.Port(portVal+"/tcp"))
		if err == nil {
			fmt.Printf("Handler port mapped to %s\n", handlerPort.Port())
			break
		}

		fmt.Printf("Waiting for handler port to be mapped... %s\n", err)

		time.Sleep(1 * time.Second)
	}

	ip, _ := dogecoinContainer.Host(ctx)
	mappedPort, _ := dogecoinContainer.MappedPort(ctx, nat.Port(portVal+"/tcp"))

	fmt.Printf("Dogecoin is running at %s:%s\n", ip, mappedPort.Port())

	d.Container = dogecoinContainer

	d.Rpc = rpc.NewRpcTransport(&rpc.Config{
		RpcUrl:  "http://" + d.config.Host + ":" + mappedPort.Port(),
		RpcUser: "test",
		RpcPass: "test",
	})

	return nil
}

func (d *DogeTest) Stop() error {
	if d.Container != nil {
		d.Container.Terminate(context.Background())
	}

	return nil
}

func WriteDockerfileToDisk() (string, error) {
	tempDir := path.Join(os.TempDir(), "dogetest") + strconv.Itoa(rand.Intn(1000000))
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		return "", err
	}

	tmpfile, err := os.CreateTemp(tempDir, "Dockerfile-*")
	if err != nil {
		return "", err
	}
	if _, err := tmpfile.Write(dockerfileData); err != nil {
		tmpfile.Close()
		return "", err
	}
	tmpfile.Close()
	return tmpfile.Name(), nil
}
