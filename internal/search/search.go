package search

import (
	"github.com/emersion/go-imap"
)

func GetSCByText(text []string) *imap.SearchCriteria {
	sc := imap.NewSearchCriteria()
	sc.Text = text
	return sc
}
