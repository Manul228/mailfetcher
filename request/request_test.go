package request

import (
	"testing"

	"github.com/emersion/go-imap"
)

var req = &Request{
	Text: []string{"dddd", "aaaa"},
}

func TestBuildSC(t *testing.T) {
	cr := imap.NewSearchCriteria()

	cr.Or = [][2]*imap.SearchCriteria{
		{
			{Text: []string{"sdfsdf"}},
			{Text: []string{"ssss"}},
		},
	}
	t.Log(req)
}
