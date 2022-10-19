package cmd

import (
	"Hermes/client"
	"github.com/spf13/cobra"
)

var (
	serverHost string
	serverPort int
)

func init() {
	clientCmd.Flags().StringVarP(&serverHost, "server-serverHost", "i", "127.0.0.1", "ws serverHost")
	clientCmd.Flags().IntVarP(&serverPort, "server-port", "p", 1161, "ws port")
}

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Run the Hermes client",
	Long:  `Run the Hermes client.`,
	Args:  cobra.ExactArgs(0),
	Run: func(_ *cobra.Command, args []string) {
		client.Start(serverHost, serverPort)
	},
}
