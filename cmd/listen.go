/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"MemInspector/rpc/server"
	"fmt"

	"github.com/spf13/cobra"
)

// listenCmd represents the listen command
var listenCmd = &cobra.Command{
	Use:   "listen",
	Short: "start rpc listening",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("listening...")
		server.StartListening()
	},
}

func init() {
	rootCmd.AddCommand(listenCmd)
}
