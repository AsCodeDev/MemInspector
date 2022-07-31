/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// listenCmd represents the listen command
var listenCmd = &cobra.Command{
	Use:   "listen",
	Short: "start rpc daemon",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("listen called")
	},
}

func init() {
	rootCmd.AddCommand(listenCmd)
}
