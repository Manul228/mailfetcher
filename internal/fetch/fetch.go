package fetch

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"mailfetcher/configs"
	"os"
	"strconv"

	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

func MessageToString(msg *imap.Message) string {
	var buffer bytes.Buffer

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
		n, err = buffer.Write(buf)
		if err != nil {
			log.Fatal(err)
		}
		if n != len {
			log.Fatal("Didn't write correct length")
		}
	}
	return buffer.String()
}

func SaveMessage(message string, fname string) {
	_, err := os.Create(fname)
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.Open(fname)
	if err != nil {
		log.Fatal(err)
	}
	f.WriteString(message)
	f.Close()
}

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
	items := []imap.FetchItem{section.FetchItem()}

	messages := make(chan *imap.Message, 10)
	go func() {
		if err := c.Fetch(seqSet, items, messages); err != nil {
			log.Fatal(err)
		}
	}()

	i := 0
	for msg := range messages {
		f, err := os.Create("tmp/" + strconv.Itoa(i) + ".eml")
		check(err)
		i++

		w := bufio.NewWriter(f)
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

			fmt.Fprintf(w, "%s", buf)
		}
		f.Close()
	}

	log.Println("Done!")
}
