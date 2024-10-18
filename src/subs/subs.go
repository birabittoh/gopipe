package subs

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"strings"
)

type TimedText struct {
	XMLName xml.Name `xml:"timedtext"`
	Body    Body     `xml:"body"`
}

type Body struct {
	Paragraphs []Paragraph `xml:"p"`
}

type Paragraph struct {
	Start     int        `xml:"t,attr"`    // Start time in milliseconds
	Length    int        `xml:"d,attr"`    // Duration in milliseconds
	Text      string     `xml:",chardata"` // Direct text (for cases without <s> tags)
	Sentences []Sentence `xml:"s"`         // List of <s> tags (for cases with individual words/phrases)
}

type Sentence struct {
	Text string `xml:",chardata"` // Text inside the <s> tag
	Time int    `xml:"t,attr"`    // Optional start time (not always present)
}

func writeVTT(output *bytes.Buffer, i, startTime, endTime int, sentence string) (int, error) {
	return output.Write(
		[]byte(
			fmt.Sprintf(
				"%d\n%s --> %s\n%s\n\n",
				i,
				millisecondsToTimestamp(startTime),
				millisecondsToTimestamp(endTime),
				sentence,
			),
		))
}

// Convert milliseconds to WebVTT timestamp format: HH:MM:SS.mmm
func millisecondsToTimestamp(ms int) string {
	seconds := ms / 1000
	milliseconds := ms % 1000
	return fmt.Sprintf("%02d:%02d:%02d.%03d", seconds/3600, (seconds%3600)/60, seconds%60, milliseconds)
}

func Convert(reader io.Reader) (buffer bytes.Buffer, err error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return
	}

	var timedText TimedText
	err = xml.Unmarshal(content, &timedText)
	if err != nil {
		log.Println("Error unmarshalling XML:", err)
		return
	}

	n, err := buffer.Write([]byte("WEBVTT\n\n"))
	if err != nil {
		return
	}

	l := len(timedText.Body.Paragraphs) - 1
	var m int
	for i, p := range timedText.Body.Paragraphs {
		startTimeMS := p.Start
		endTimeMS := p.Start + p.Length

		if i < l {
			nextStartTimeMS := timedText.Body.Paragraphs[i+1].Start
			if nextStartTimeMS < endTimeMS {
				endTimeMS = nextStartTimeMS
			}
		}

		var sentence string
		if len(p.Sentences) > 0 {
			for _, s := range p.Sentences {
				sentence += s.Text
			}
		} else {
			sentence = p.Text
		}

		sentence = strings.TrimSpace(sentence)
		if sentence == "" {
			continue
		}

		m, err = writeVTT(&buffer, i+1, startTimeMS, endTimeMS, sentence)
		n += m
		if err != nil {
			return
		}
	}

	return
}
