package main

import (
	"GabrielChaves1/caching-proxy/internal/proxy"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var proxyInst *proxy.Proxy

var rootCmd = &cobra.Command{
	Use:   "caching-proxy",
	Short: "A simple caching proxy",
	Run: func(cmd *cobra.Command, args []string) {
		port := viper.GetInt("port")
		origin := viper.GetString("origin")
		clearCache := viper.GetBool("clear-cache")

		if clearCache {
			fmt.Printf("Clearing cache\n")
			proxyInst.ClearCache()
			os.Exit(0)
		}

		if port == 0 {
			fmt.Println("Port is required")
			return
		}

		if origin == "" {
			fmt.Println("Origin is required")
			return
		}

		proxyInst.Origin = origin

		http.Handle("/", proxyInst)
		fmt.Printf("Starting server on :%d\n", port)
		fmt.Printf("Proxying requests to %s\n", origin)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
			fmt.Printf("Error starting server: %s\n", err)
			return
		}
	},
}

func init() {
	rootCmd.Flags().StringP("port", "p", "", "Port to listen on")
	rootCmd.Flags().StringP("origin", "o", "", "Origin server")
	rootCmd.Flags().BoolP("clear-cache", "", false, "Clear the cache")

	viper.BindPFlag("port", rootCmd.Flags().Lookup("port"))
	viper.BindPFlag("origin", rootCmd.Flags().Lookup("origin"))
	viper.BindPFlag("clear-cache", rootCmd.Flags().Lookup("clear-cache"))

	proxyInst = proxy.NewProxy("")
}

func main() {
	rootCmd.Execute()
}
