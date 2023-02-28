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
	if date1.IsZero() || date2.IsZero() {
		return false
	}
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

func (r Request) search(seq *[]uint32, c *client.Client) error {
	var err error

	if len(r.Text) > 0 && len(r.Keywords) > 0 {
		return fmt.Errorf("the specified arguments --text and --keywords cannot be present at the same time")
	}

	if len(r.Since) == 0 && len(r.Before) == 0 {
		log.Println("The time period is not specified. The search is performed all the time.")
	}

	var since, before time.Time

	if len(r.Since) > 0 {
		since, err = time.Parse("02.01.2006", r.Since)
		if err != nil {
			return fmt.Errorf("since parameter is incorrect")
		}
	}

	if len(r.Before) > 0 {
		before, err = time.Parse("02.01.2006", r.Before)
		if err != nil {
			return fmt.Errorf("before parameter is incorrect")
		}
	}

	if dateEqual(since, before) {
		log.Println("Start date must not be equal end date.")
	}

	seqNumsSet := make(map[uint32]struct{})

	if len(r.Keywords) > 0 {
		keywords := make(chan string, 10)

		go func() {
			for _, kw := range r.Keywords {
				keywords <- kw
			}
			close(keywords)
		}()

		for kw := range keywords {
			sc := imap.NewSearchCriteria()
			sc.Since = since
			sc.Before = before
			sc.Text = append(sc.Text, kw)
			log.Println("Searching for --keyword", kw)
			seqNums, err := c.Search(sc)
			if err != nil {
				return fmt.Errorf("keyword search failed")
			}

			for _, n := range seqNums {
				seqNumsSet[n] = struct{}{}
			}
		}
	}

	if len(r.Text) > 0 {
		sc := imap.NewSearchCriteria()
		sc.Since = since
		sc.Before = before
		sc.Text = append(sc.Text, r.Text...)
		log.Println("Searching for --text", r.Text)
		seqNums, err := c.Search(sc)
		if err != nil {
			return fmt.Errorf("text search failed")
		}

		for _, n := range seqNums {
			seqNumsSet[n] = struct{}{}
		}
	}

	for num := range seqNumsSet {
		*seq = append(*seq, num)
	}

	return nil
}

func (r Request) Fetch() {
	var err error

	log.Println("Connecting to server...")
	c, err := client.DialTLS(r.Server, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")

	defer c.Logout()

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

	var seqNums []uint32
	err = r.search(&seqNums, c)
	if len(seqNums) == 0 {
		log.Fatalln("No messages found!")
	}
	log.Printf("Found %d items!", len(seqNums))

	if err != nil {
		log.Fatal(err)
	}

	seqSet := new(imap.SeqSet)
	seqSet.AddNum(seqNums...)

	var section imap.BodySectionName
	items := []imap.FetchItem{section.FetchItem(), imap.FetchEnvelope}

	start := time.Now()
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
		prefix := msg.Envelope.From[0].Address() + " " + msg.Envelope.MessageId + " "
		f, err := w.Create(prefix + msg.Envelope.Subject + ".eml")
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
	finish := time.Since(start)
	log.Println("Done in ", finish)

}
