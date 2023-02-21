package cmd

import (
	"log"
	"mailfetcher/internal/fetch"
	"os"

	"github.com/spf13/cobra"
)

var Server string
var Login string
var Password string

var rootCmd = &cobra.Command{
	Use:   "mailfetcher",
	Short: "Utility for receiving mail messages via IMAP protocol according to specified criteria.",
	Run: func(cmd *cobra.Command, args []string) {
		fetch.Fetch(Server, Login, Password)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&Server, "server", "s", "", "Server like imap.yandex.ru:993")
	rootCmd.PersistentFlags().StringVarP(&Login, "login", "l", "", "Login")
	rootCmd.PersistentFlags().StringVarP(&Password, "password", "p", "", "Password")

	err := rootCmd.MarkPersistentFlagRequired("server")
	if err != nil {
		log.Fatalln(err)
	}

	err = rootCmd.MarkPersistentFlagRequired("login")
	if err != nil {
		log.Fatalln(err)
	}

	err = rootCmd.MarkPersistentFlagRequired("password")
	if err != nil {
		log.Fatalln(err)
	}

}
