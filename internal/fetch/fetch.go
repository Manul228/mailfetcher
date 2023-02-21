package fetch

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

func Fetch(server, login, password string) {
	log.Println("Connecting to server...")

	// Connect to server
	c, err := client.DialTLS(server, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")

	// Don't forget to logout
	defer c.Logout()

	// Login
	if err := c.Login(login, password); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")

	// List mailboxes
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

	cr0 := imap.NewSearchCriteria()
	cr1 := imap.NewSearchCriteria()
	text := []string{"80135"}
	cr1.Text = text

	since := time.Date(2022, 9, 28, 00, 00, 0, 0, time.UTC)
	// Before date NOT included
	before := time.Date(2022, 9, 29, 23, 59, 0, 0, time.UTC)

	cr0.Since = since
	cr0.Before = before

	seqNums0, err := c.Search(cr0)
	//seqNums1, err := c.Search(cr1)

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
