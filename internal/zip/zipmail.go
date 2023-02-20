package zipmail

import (
	"archive/zip"
	"bytes"
	"fmt"
	"log"
	"os"
)

type Item struct {
	Name string
	Body bytes.Buffer
}

func addToZipArchive(archiveName string, items chan Item) {
	archive, err := os.Create(archiveName)
	if err != nil {
		log.Fatal(err)
	}
	defer archive.Close()

	zipWriter := zip.NewWriter(archive)

	for item := range items {
		w, err := zipWriter.Create(item.Name)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(w, "%s", &item.Body)
	}
}
