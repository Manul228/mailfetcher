package zipmail

import (
	"testing"
)

func TestZip(m *testing.T) {

	items := [2]Item{}
	items[0].Name = "name0"
	items[0].Body.WriteString("body0")
	items[1].Name = "name1"
	items[1].Body.WriteString("body1")

	itemsChan := make(chan Item, 2)

	go func() {
		for _, item := range items {
			itemsChan <- item
		}
		close(itemsChan)
	}()

	addToZipArchive("../../tmp/test.zip", itemsChan)
}
