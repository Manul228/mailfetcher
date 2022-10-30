package fetch

import (
	"log"
	"mailfetcher/configs"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

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
	
	err := <-done; err != nil {
		log.Fatal(err)
	}

	 
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Flags for INBOX:", mbox.Flags)

	// Get the last 4 messages
	from := uint32(1)
	to := mbox.Messages
	if mbox.Messages > 3 {
		// We're using unsigned integers here, only subtract if the result is > 0
		from = mbox.Messages - 3
	}
	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)

	messages := make(chan *imap.Message, 10)
	done = make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, messages)
	}()

	// https://stackoverflow.com/questions/55203878/how-to-find-attachments-and-download-them-with-mxk-go-imap
	// https://godocs.io/github.com/emersion/go-message#example-Read
	log.Println("Last 4 messages:")
	for msg := range messages {
		log.Println("* " + msg.Envelope.Subject)
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	log.Println("Done!")
}
