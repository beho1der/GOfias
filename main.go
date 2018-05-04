package main

// An example streaming XML parser.

import (
	"encoding/xml"
	"fmt"
	"gopkg.in/mgo.v2"
	//	"gopkg.in/mgo.v2/bson"
	"os"
)

type Doctype struct {
	NDTYPEID int    `xml:"NDTYPEID,attr"`
	NAME     string `xml:"NAME,attr"`
}

type Data struct {
	Doctypes []Doctype `xml:"NormativeDocumentType"`
}

type Cast struct {
	Ndtype    int    `bson:"ndtype"`
	Normative string `bson:"normative"`
}

func main() {

	session, err := mgo.Dial("mongodb://192.168.155.5")
	if err != nil {
		panic(err)
	}
	defer session.Close()
	// получаем коллекцию
	normaCollection := session.DB("fias").C("norma")

	xmlFile, err := os.Open("AS_NDOCTYPE.XML")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer xmlFile.Close()

	decoder := xml.NewDecoder(xmlFile)
	total := 0
	var inElement string
	for {
		// Read tokens from the XML document in a stream.
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		// Inspect the type of the token just read.
		switch se := t.(type) {
		case xml.StartElement:
			// If we just read a StartElement token
			inElement = se.Name.Local
			// ...and its name is "page"
			if inElement == "NormativeDocumentType" {
				var data Doctype
				// decode a whole chunk of following XML into the
				// variable p which is a Page (se above)
				decoder.DecodeElement(&data, &se)
				fmt.Printf("%v", data)
				p1 := &Cast{Ndtype: data.NDTYPEID, Normative: data.NAME}
				err = normaCollection.Insert(p1)
				if err != nil {
					fmt.Println(err)
				}
				// Do some stuff with the page.
				total++
			}
		default:
		}

	}

	fmt.Printf("Total articles: %d \n", total)
	//fmt.Printf("%v", data)

}