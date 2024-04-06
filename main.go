package relayer

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use: "relayer",
	}

	var ss58_address string
	var amount string

	var delegateCmd = &cobra.Command{
		Use: "delegate",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(RunPython3Command([]string{
				"btcli/delegate.py", "delegate",
				"--ss58-address", ss58_address,
				"--amount", string(amount),
			}))
		},
	}
	delegateCmd.Flags().StringVarP(&ss58_address, "ss58-address", "", "", "Destination address")
	delegateCmd.MarkFlagRequired("ss58-address")
	delegateCmd.Flags().StringVarP(&amount, "amount", "a", "", "TAO amount")
	delegateCmd.MarkFlagRequired("amount")
	rootCmd.AddCommand(delegateCmd)

	var undelegateCmd = &cobra.Command{
		Use: "undelegate",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(RunPython3Command([]string{
				"btcli/delegate.py", "undelegate",
				"--ss58-address", ss58_address,
				"--amount", string(amount),
			}))
		},
	}
	undelegateCmd.Flags().StringVarP(&ss58_address, "ss58-address", "", "", "Source address")
	undelegateCmd.MarkFlagRequired("ss58-address")
	undelegateCmd.Flags().StringVarP(&amount, "amount", "a", "", "TAO amount")
	undelegateCmd.MarkFlagRequired("amount")
	rootCmd.AddCommand(undelegateCmd)

	var serveCmd = &cobra.Command{
		Use: "serve",
		Run: func(cmd *cobra.Command, args []string) {
			Serve()
		},
	}
	rootCmd.AddCommand(serveCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
