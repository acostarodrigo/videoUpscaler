package videoUpscaler

import (
	io "io"
	"net/http"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/janction/videoUpscaler/db"
	"github.com/janction/videoUpscaler/ipfs"
	"github.com/janction/videoUpscaler/videoUpscalerLogger"
)

func (w Worker) RegisterWorker(address string, stake types.Coin, db *db.DB) error {
	time.Sleep(8 * time.Second) // Delay 8 seconds before registering
	db.Addworker(address)
	// Base arguments
	args := []string{
		"tx", "videoUpscaler", "add-worker",
	}
	ip, _ := getPublicIP()
	ipfsId, _ := ipfs.GetIPFSPeerID()
	args = append(args, ip)
	args = append(args, ipfsId)
	args = append(args, stake.String())
	args = append(args, "--from")
	args = append(args, address)
	args = append(args, "--yes")

	err := ExecuteCli(args)
	if err != nil {
		db.DeleteWorker(address)
	}
	return nil
}

// GetPublicIP fetches the public IP of the machine
func getPublicIP() (string, error) {
	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		videoUpscalerLogger.Logger.Error(err.Error())
		return "", err
	}
	defer resp.Body.Close()

	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		videoUpscalerLogger.Logger.Error(err.Error())
		return "", err
	}

	return strings.TrimSpace(string(ip)), nil
}

func (w *Worker) DeclareWinner(payment types.Coin) {
	w.CurrentTaskId = ""
	w.CurrentThreadIndex = 0
	w.Reputation.Points = w.Reputation.Points + 1
	w.Reputation.Solutions = w.Reputation.Solutions + 1
	w.Reputation.Winnings = w.Reputation.Winnings.Add(payment)
}

func (w *Worker) ReleaseValidator() {
	w.CurrentTaskId = ""
	w.CurrentThreadIndex = 0
}
