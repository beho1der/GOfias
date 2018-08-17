package main

// An example streaming XML parser.

import (
	"encoding/xml"
	"fmt"
	"gopkg.in/mgo.v2"
	"io/ioutil"
	"os"
	"regexp"
	"unsafe"
)

var m map[string][]string

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

type House struct {
	AOGUID   string `xml:"AOGUID,attr"`
	ENDDATE  string `xml:"ENDDATE,attr"`
	HOUSENUM string `xml:"HOUSENUM,attr"`
	BUILDNUM string `xml:"BUILDNUM,attr"`
	STRUCNUM string `xml:"STRUCNUM,attr"`
}

type Data struct {
	Doctypes []Address `xml:"NormativeDocumentType"`
}

type Mongo_Region struct {
	Region   string `bson:"region"`
	Code     string `bson:"regioncode"`
	Fullname string `bson:"fullname"`
	Sufix    string `bson:"sufix"`
	Id       string `bson:"_id"`
}

type Mongo_City struct {
	City     string `bson:"city"`
	Fullname string `bson:"fullname"`
	Sufix    string `bson:"sufix"`
	Parent   string `bson:"parent"`
	Id       string `bson:"_id"`
}

type Mongo_Street struct {
	Street   string `bson:"street"`
	Fullname string `bson:"fullname"`
	Sufix    string `bson:"sufix"`
	Parent   string `bson:"parent"`
	Id       string `bson:"_id"`
}

type Mongo_House struct {
	Houseid  string `bson:"houseid"`
	Number   string `bson:"number"`
	Strnum   string `bson:"strnum"`
	Buildnum string `bson:"buildnum"`
}

// читает директорию и находит совпадения по имени,возвращает полное имя файла
func readDir(patch, filter string) string {
	re := regexp.MustCompile(filter)
	files, err := ioutil.ReadDir(patch)
	if err != nil {
		fmt.Println(err)
	}
	for _, f := range files {
		t := re.FindStringSubmatch(f.Name())
		if len(t) >= 1 {
			return f.Name()
		}

	}
	return "Отсутсвует файл"
}

// функция поиска в массиве нужного значение
func findAO(s []int, f int) bool {
	for i := range s {
		if s[i] == f {
			return true
		}
	}
	return false
}

func createHouse(s *mgo.Session, v string) {
	session := s.Copy()
	defer session.Close()
	streetCollection := session.DB("fias").C("street")
	patch := "/home/captain/Загрузки/fias_xml/"
	name := readDir(patch, "AS_HOUSE")
	xmlFile, err := os.Open(patch + name)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer xmlFile.Close()
	decoder := xml.NewDecoder(xmlFile)
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
			if inElement == "House" {
				var data House
				decoder.DecodeElement(&data, &se)
				if data.CURRSTATUS == 0 && data.LIVESTATUS == 1 && findAO(cityCode, data.AOLEVEL) == true && data.REGIONCODE == v {
					//fmt.Printf("%v", data)
					p := &Mongo_City{City: data.FORMALNAME, Fullname: data.FORMALNAME, Sufix: data.SHORTNAME, Parent: data.PARENTGUID, Id: data.AOGUID}
					err = cityCollection.Insert(p)
					if err != nil {
						fmt.Println(err)
					}
					city++
				}
			}
		}
	}
}

func createAddrobj(s *mgo.Session, v string) {
	session := s.Copy()
	defer session.Close()
	regionCollection := session.DB("fias").C("region")
	cityCollection := session.DB("fias").C("city")
	streetCollection := session.DB("fias").C("street")
	m = make(map[string][]string)
	patch := "/home/captain/Загрузки/fias_xml/"
	name := readDir(patch, "AS_ADDROBJ")
	xmlFile, err := os.Open(patch + name)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer xmlFile.Close()
	decoder := xml.NewDecoder(xmlFile)
	var city, street, region = 0, 0, 0
	cityCode := []int{4, 5, 6, 90}
	streetCode := []int{7, 91}
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
				// Декодируем xml и разбираем его в зависимости от полей
				// AOLEVEL = 1 определяем регион
				decoder.DecodeElement(&data, &se)
				// находим Регион
				if data.CURRSTATUS == 0 && data.LIVESTATUS == 1 && data.AOLEVEL == 1 && data.REGIONCODE == v {
					p := &Mongo_Region{Region: data.FORMALNAME, Code: data.REGIONCODE, Fullname: data.FORMALNAME, Sufix: data.SHORTNAME, Id: data.AOGUID}
					err = regionCollection.Insert(p)
					if err != nil {
						fmt.Println(err)
					}
					region++
				}
				// AOLEVEL = 4,5,6,90 определяем город,населенный пункт,СНТ
				if data.CURRSTATUS == 0 && data.LIVESTATUS == 1 && findAO(cityCode, data.AOLEVEL) == true && data.REGIONCODE == v {
					//fmt.Printf("%v", data)
					p := &Mongo_City{City: data.FORMALNAME, Fullname: data.FORMALNAME, Sufix: data.SHORTNAME, Parent: data.PARENTGUID, Id: data.AOGUID}
					err = cityCollection.Insert(p)
					if err != nil {
						fmt.Println(err)
					}
					city++
				}
				// находим Улицу
				if data.CURRSTATUS == 0 && data.LIVESTATUS == 1 && findAO(streetCode, data.AOLEVEL) == true && data.REGIONCODE == v {
					//fmt.Printf("%v", data)
					p := &Mongo_Street{Street: data.FORMALNAME, Fullname: data.FORMALNAME, Sufix: data.SHORTNAME, Parent: data.PARENTGUID, Id: data.AOGUID}
					err = streetCollection.Insert(p)
					if err != nil {
						fmt.Println(err)
					}
					// добавляем AOGUID в массив для поиска домов
					m[v] = append(m[v], data.AOGUID)
					street++
				}
			}

		default:
		}
	}
}

func main() {
	session, err := mgo.Dial("mongodb://192.168.155.5")
	if err != nil {
		panic(err)
	}
	defer session.Close()
	createAddrobj(session, "01")
	fmt.Println("Real size of m:", unsafe.Sizeof(m)+unsafe.Sizeof([1000]int32{}))

}

//fmt.Printf("Total articles: %d \n", total)
