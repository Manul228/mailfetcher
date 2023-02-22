package request

import (
	"archive/zip"
	"fmt"
	"log"
	"os"

	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/nexidian/gocliselect"
)

func check(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func dateEqual(date1, date2 time.Time) bool {
	y1, m1, d1 := date1.Date()
	y2, m2, d2 := date2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

type Request struct {
	Server   string
	Login    string
	Password string
	Text     []string
	Keywords []string
	Since    string
	Before   string
	Output   string
}

func (r Request) Fetch() {
	cr := imap.NewSearchCriteria()
	var err error

	if len(r.Since) == 0 || len(r.Before) == 0 {
		log.Println("The time period is not specified. The search is performed all the time.")
	}

	if len(r.Since) > 0 {
		cr.Since, err = time.Parse("02.01.2006", r.Since)
		log.Println("Since ", cr.Since.String())
		check(err)
	}

	if len(r.Before) > 0 {
		cr.Before, err = time.Parse("02.01.2006", r.Before)
		log.Println("Before ", cr.Before.String())
		check(err)
	}

	if dateEqual(cr.Since, cr.Before) {
		log.Println("Start date must not be equal end date.")
		return
	}

	log.Println("Connecting to server...")

	// Connect to server
	c, err := client.DialTLS(r.Server, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")

	defer c.Logout()

	// Login
	if err := c.Login(r.Login, r.Password); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")

	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", mailboxes)
	}()

	menu := gocliselect.NewMenu("Select mailbox")
	for m := range mailboxes {
		status := imap.NewMailboxStatus(m.Name, []imap.StatusItem{imap.StatusMessages})
		var amount string = ""
		if status.Messages > 0 {
			amount = fmt.Sprint(status.Messages)
		}
		menu.AddItem(m.Name+" "+amount, m.Name)
	}
	chosenMailbox := menu.Display()

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	mbox, err := c.Select(chosenMailbox, false)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Flags for %s %+v:", chosenMailbox, mbox.Flags)

	seqNums, err := c.Search(cr)

	if err != nil {
		log.Fatal(err)
	}

	seqSet := new(imap.SeqSet)
	seqSet.AddNum(seqNums...)

	var section imap.BodySectionName
	items := []imap.FetchItem{section.FetchItem(), imap.FetchEnvelope}

	messages := make(chan *imap.Message, 10)
	go func() {
		if err := c.Fetch(seqSet, items, messages); err != nil {
			log.Fatal("Request failed: \n", err)
		}
	}()

	archive, err := os.Create(r.Output)
	check(err)
	w := zip.NewWriter(archive)

	for msg := range messages {
		f, err := w.Create(msg.Envelope.Subject + ".eml")
		check(err)

		for _, value := range msg.Body {
			len := value.Len()
			buf := make([]byte, len)
			n, err := value.Read(buf)
			if err != nil {
				log.Fatal(err)
			}
			if n != len {
				log.Fatal("Didn't read correct length")
			}

			f.Write(buf)
		}
	}
	w.Close()

	log.Println("Done!")
}
