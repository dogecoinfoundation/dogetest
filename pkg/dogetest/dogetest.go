package dogetest

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dogecoinfoundation/dogetest/pkg/rpc"
	"github.com/shirou/gopsutil/process"
)

type DogeTest struct {
	Host         string
	Port         int
	installation string
	cmd          *exec.Cmd
	Rpc          *rpc.RpcTransport
	config       DogeTestConfig
}

type DogeTestConfig struct {
	Host             string
	InstallationPath string
	ConfigPath       string
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
	os.RemoveAll(config.ConfigPath)

	port, err := findAvailablePort(config.Host)
	if err != nil {
		return nil, err
	}

	rpcClient := rpc.NewRpcTransport(&rpc.Config{
		RpcUrl:  "http://localhost:" + strconv.Itoa(port),
		RpcUser: "test",
		RpcPass: "test",
	})

	return &DogeTest{
		Host:         config.Host,
		Port:         port,
		installation: config.InstallationPath,
		Rpc:          rpcClient,
		config:       config,
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
	configPath, err := d.writeTempConfig(d.Port)
	if err != nil {
		return err
	}

	cmd := exec.Command(d.installation, "-regtest", "-reindex-chainstate", "-min", "-splash=0", "-conf="+configPath)
	d.cmd = cmd

	err = cmd.Start()
	if err != nil {
		return err
	}

	err = d.waitForPort("localhost:"+strconv.Itoa(d.Port), 10*time.Second)
	if err != nil {
		return err
	}

	for {
		_, err := d.Rpc.GetInfo()
		if err == nil {
			break // success
		}

		time.Sleep(1 * time.Second) // or however long you want between tries
	}

	fmt.Println("DogeTest started")

	return nil
}

func (d *DogeTest) Stop() error {
	d.cmd.Process.Kill()

	err := os.RemoveAll(d.config.ConfigPath)
	if err != nil {
		return err
	}

	return nil
}

func (d *DogeTest) ClearProcess() error {
	processes, err := process.Processes()
	if err != nil {
		fmt.Println("Error fetching processes:", err)
		return err
	}

	targetName := "dogecoind.exe"

	for _, p := range processes {
		name, err := p.Name()

		if err == nil && strings.Contains(strings.ToLower(name), strings.ToLower(targetName)) {
			pid := p.Pid
			fmt.Printf("Found process: %s (PID: %d)\n", name, pid)
			err = p.Kill()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (d *DogeTest) waitForPort(address string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for {
		conn, err := net.DialTimeout("tcp", address, 1*time.Second)
		if err == nil {
			conn.Close()
			return nil // Port is open!
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for %s", address)
		}

		time.Sleep(200 * time.Millisecond) // small delay before retry
	}
}

func (d *DogeTest) writeTempConfig(port int) (string, error) {
	tempDir, err := os.MkdirTemp("", "doge-test")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %v", err)
	}

	content := fmt.Sprintf(`regtest=1
rpcuser=test
rpcpassword=test
rpcport=%d
server=1
txindex=1  # Optional: Enables transaction index for address tracking
rpcbind=0.0.0.0
rpcallowip=0.0.0.0/0  # Temporarily allow all IPs for testing`, port)

	filePath := filepath.Join(tempDir, "dogecoin.conf")
	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write config file: %v", err)
	}

	return filePath, nil
}

func findAvailablePort(host string) (int, error) {
	for port := 18000; port < 19000; port++ {
		l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
		if err == nil {
			l.Close()
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available port found")
}

func findDogeInstallation() (string, error) {
	possiblePaths := []string{
		"C:\\Program Files\\Dogecoin\\daemon\\dogecoind.exe",
		"/usr/local/bin/dogecoind",
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	exeName := "dogecoind"
	path, err := exec.LookPath(exeName)
	if err != nil {
		return "", fmt.Errorf("doge installation not found")
	}

	return path, nil
}
