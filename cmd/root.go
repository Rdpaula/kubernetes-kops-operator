package cmd

import (
	"github.com/spf13/cobra"
)

var (
	clusterName string
	namespace   string
)

var rootCmd = &cobra.Command{
	Use:   "operator-cli",
	Short: "operator-cli",
	Long:  `operator-cli to run some operations`,
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.PersistentFlags().StringVar(&clusterName, "name", "", "name of the CR of the cluster to be planned")
	rootCmd.PersistentFlags().StringVar(&namespace, "namespace", "", "namespace of the cluster to be planned")
	rootCmd.MarkFlagsRequiredTogether("name", "namespace")
}
