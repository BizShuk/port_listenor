package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/bizshuk/port_listenor/svc"
)

var (
	checkPorts   string
	checkTimeout string
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Run a one-time port status check",
	Long:  `Check defined ports once and print the results immediately to the console.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var portsToCheck []svc.PortEntry

		if checkPorts != "" {
			parts := strings.Split(checkPorts, ",")
			for _, pStr := range parts {
				pStr = strings.TrimSpace(pStr)
				portNum, err := strconv.Atoi(pStr)
				if err != nil {
					return fmt.Errorf("invalid port number: %s", pStr)
				}
				portsToCheck = append(portsToCheck, svc.PortEntry{
					Port: portNum,
					Name: fmt.Sprintf("port-%d", portNum),
				})
			}
		}

		cfg := &svc.CheckConfig{
			PortsToCheck: portsToCheck,
			TimeoutVal:   checkTimeout,
			Writer:       os.Stdout,
		}

		return svc.RunOneTimeCheck(cfg, GetGlobalConfig())
	},
}

func init() {
	checkCmd.Flags().StringVarP(&checkPorts, "ports", "p", "", "comma-separated list of ports to check (e.g. 80,443,3000)")
	checkCmd.Flags().StringVar(&checkTimeout, "timeout", "", "connection timeout (e.g. 2s, 5s)")
	portCmd.AddCommand(checkCmd)
}
