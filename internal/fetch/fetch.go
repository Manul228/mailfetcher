package fetch

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"mailfetcher/configs"
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

func Fetch(creds *configs.Credentials) {
	log.Println("Connecting to server...")

	// Connect to server
	c, err := client.DialTLS(configs.Creds.Server, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")

	// Don't forget to logout
	defer c.Logout()

	// Login
	if err := c.Login(configs.Creds.Login, configs.Creds.Password); err != nil {
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

	cr0.Since = time.Date(2022, 9, 28, 00, 00, 0, 0, time.UTC)
	// Before date NOT included
	cr0.Before = time.Date(2022, 9, 29, 23, 59, 0, 0, time.UTC)

	seqNums0, err := c.Search(cr0)
	//seqNums1, err := c.Search(cr1)

	check(err)

	seqSet := new(imap.SeqSet)
	seqSet.AddNum(seqNums0...)

	messages := make(chan *imap.Message, 10)

	items := []imap.FetchItem{"ENVELOPE", "BODY[]"}

	go func() {
		err := c.Fetch(seqSet, items, messages)
		check(err)
	}()

	for msg := range messages {
		filename := "tmp/" + msg.Envelope.Subject + ".eml"
		f, err := os.Create(filename)
		check(err)
		w := bufio.NewWriter(f)
		for _, value := range msg.Body {
			len := value.Len()
			buf := make([]byte, len)
			readed, err := value.Read(buf)
			check(err)
			if readed != len {
				log.Fatal("Didn't read correct length")
			}
			fmt.Fprintf(w, "%s", buf)
		}
		f.Close()
	}
	log.Println("Done!")
}
