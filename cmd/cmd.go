package cmd

import (
	"github.com/spf13/cobra"
)

func Execute() {

	var r = &cobra.Command{
		Use:   "hermes",
		Short: "Hermes is a clipboard service.",
		Long:  "Hermes is a clipboard service.",
	}

	r.AddCommand(
		serverCmd,
		clientCmd,
	)
	r.Execute()
}
