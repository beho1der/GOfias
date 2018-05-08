package main

// An example streaming XML parser.

import (
	"encoding/xml"
	"fmt"
	"gopkg.in/mgo.v2"
	//	"gopkg.in/mgo.v2/bson"
	"os"
)

type Address struct {
	AOGUID     string `xml:"AOGUID,attr"`
	FORMALNAME string `xml:"FORMALNAME,attr"`
	OFFNAME    string `xml:"FOFFNAME,attr"`
	SHORTNAME  string `xml:"SHORTNAME,attr"`
	PARENTGUID string `xml:"PARENTGUID,attr"`
	CURRSTATUS int    `xml:"CURRSTATUS,attr"`
	LIVESTATUS int    `xml:"LIVESTATUS,attr"`
	REGIONCODE string `xml:"REGIONCODE,attr"`
	AOLEVEL    int    `xml:"AOLEVEL,attr"`
}

type Data struct {
	Doctypes []Address `xml:"NormativeDocumentType"`
}

type Mongo_Region struct {
	Region   string `bson:"region"`
	Code     string `bson:"regioncode"`
	Fullname string `bson:"fullname"`
	Sufix    string `bson:"sufix"`
	Parent   string `bson:"parent"`
	Id       string `bson:"id"`
}

func main() {

	session, err := mgo.Dial("mongodb://192.168.155.5")
	if err != nil {
		panic(err)
	}
	defer session.Close()
	// получаем коллекцию
	regionCollection := session.DB("fias").C("region")

	xmlFile, err := os.Open("/home/captain/Загрузки/fias_xml/AS_ADDROBJ.XML")
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
			if inElement == "Object" {
				var data Address
				// decode a whole chunk of following XML into the
				// variable p which is a Page (se above)
				decoder.DecodeElement(&data, &se)
				if data.CURRSTATUS == 0 && data.LIVESTATUS == 1 && data.AOLEVEL == 1 && data.REGIONCODE == "01" {
					fmt.Printf("%v", data)
					p1 := &Mongo_Region{Region: data.FORMALNAME, Code: data.REGIONCODE, Fullname: data.FORMALNAME, Sufix: data.SHORTNAME, Parent: data.PARENTGUID, Id: data.AOGUID}
					err = regionCollection.Insert(p1)
					if err != nil {
						fmt.Println(err)
					}
					// Do some stuff with the page.
					total++
				}
			}
		default:
		}

	}

	fmt.Printf("Total articles: %d \n", total)
	//fmt.Printf("%v", data)

}
