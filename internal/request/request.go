package request

import (
	"archive/zip"
	"log"
	"os"

	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

func check(err error) {
	if err != nil {
		log.Panic(err)
	}
}

type Request struct {
	Server   string
	Login    string
	Password string
	Text     []string
	Keywords []string
	Since    string
	Before   string
}

func (r Request) Fetch() {
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

	log.Println("Mailboxes:")
	for m := range mailboxes {
		log.Println("* " + m.Name)
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	mbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Flags for INBOX:", mbox.Flags)

	cr := imap.NewSearchCriteria()

	if len(r.Since) > 0 {
		cr.Since, err = time.Parse("01.02.2006", r.Since)
		check(err)
	}

	if len(r.Before) > 0 {
		cr.Before, err = time.Parse("01.02.2006", r.Before)
		check(err)
	}

	seqNums0, err := c.Search(cr)

	if err != nil {
		log.Fatal(err)
	}

	seqSet := new(imap.SeqSet)
	seqSet.AddNum(seqNums0...)

	var section imap.BodySectionName
	items := []imap.FetchItem{section.FetchItem(), imap.FetchEnvelope}

	messages := make(chan *imap.Message, 10)
	go func() {
		if err := c.Fetch(seqSet, items, messages); err != nil {
			log.Fatal(err)
		}
	}()

	archive, err := os.Create("tmp/archive.zip")
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
