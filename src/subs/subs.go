package subs

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
)

type TimedText struct {
	XMLName xml.Name `xml:"timedtext"`
	Body    Body     `xml:"body"`
}

type Body struct {
	Paragraphs []Paragraph `xml:"p"`
}

type Paragraph struct {
	Text   string `xml:",chardata"`
	Start  int    `xml:"t,attr"` // Start time in milliseconds
	Length int    `xml:"d,attr"` // Duration in milliseconds
}

// Convert milliseconds to WebVTT timestamp format: HH:MM:SS.mmm
func millisecondsToTimestamp(ms int) string {
	seconds := ms / 1000
	milliseconds := ms % 1000
	return fmt.Sprintf("%02d:%02d:%02d.%03d", seconds/3600, (seconds%3600)/60, seconds%60, milliseconds)
}

func writeString(s string, w http.ResponseWriter) {
	w.Write([]byte(s))
}

func Parse(inputData []byte, w http.ResponseWriter) error {
	var timedText TimedText
	err := xml.Unmarshal(inputData, &timedText)
	if err != nil {
		log.Println("Error unmarshalling XML:", err)
		return err
	}

	// Write the WebVTT header
	writeString("WEBVTT\n\n", w)

	// Loop through the paragraphs and write them to the WebVTT file
	for i, p := range timedText.Body.Paragraphs {
		startTime := millisecondsToTimestamp(p.Start)
		endTime := millisecondsToTimestamp(p.Start + p.Length)

		s := fmt.Sprintf("%d\n%s --> %s\n%s\n\n", i+1, startTime, endTime, p.Text)
		writeString(s, w)
	}

	return nil
}
