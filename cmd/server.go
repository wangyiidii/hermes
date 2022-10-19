package cmd

import (
	"Hermes/server"
	"github.com/spf13/cobra"
)

var (
	port   int
	wsPort int
)

func init() {
	serverCmd.Flags().IntVarP(&port, "port", "p", 1160, "http server port")
	serverCmd.Flags().IntVarP(&wsPort, "ws-port", "w", 1161, "websocket server port")
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run the Hermes server",
	Long:  `Run the Hermes server. (CN: 启动Hermes服务)`,
	Args:  cobra.ExactArgs(0),
	Run: func(_ *cobra.Command, args []string) {
		server.GlobalServer.Start(wsPort)

	},
}
