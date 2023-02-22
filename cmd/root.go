package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"

	"mailfetcher/internal/request"
)

var req request.Request

var rootCmd = &cobra.Command{
	Use:   "mailfetcher",
	Short: "Utility for receiving mail messages via IMAP protocol according to specified criteria.",
	Run: func(cmd *cobra.Command, args []string) {
		req.Fetch()
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&req.Server, "server", "a", "", "Server like imap.yandex.ru:993")
	rootCmd.PersistentFlags().StringVarP(&req.Login, "login", "l", "", "Login.")
	rootCmd.PersistentFlags().StringVarP(&req.Password, "password", "p", "", "Password.")

	rootCmd.PersistentFlags().StringVarP(&req.Since, "since", "s", "", "Search start date.")
	rootCmd.PersistentFlags().StringVarP(&req.Before, "before", "b", "", "Search end date (not included in the range).")
	rootCmd.PersistentFlags().StringVarP(&req.Output, "output", "o", "backup.zip", "Output location.")

	rootCmd.PersistentFlags().StringSliceVarP(&req.Text, "text", "t", []string{},
		`Words separated by commas like --text=\"abra,kadabra\", which must be present at the same time.`)
	rootCmd.PersistentFlags().StringSliceVarP(&req.Keywords, "keywords", "k", []string{},
		`Keywords separated by commas like --keywords=\"abra,kadabra\", which may not necessarily be present at the same time.`)

	required := []string{"server", "login", "password"}

	for _, r := range required {
		if err := rootCmd.MarkPersistentFlagRequired(r); err != nil {
			log.Fatalln(err)
		}
	}
}
