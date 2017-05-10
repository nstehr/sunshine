package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"sync"

	mgo "gopkg.in/mgo.v2"

	"github.com/nstehr/sunshine/model"
	"github.com/nstehr/sunshine/producer"
)

const cpi96 = 88.9
const seperator = "_"

type deptData struct {
	Link  string
	Name  string
	Pages int
}

type employeeData struct {
	Year     int
	Cpi      float64
	Dept     []deptData
	AllDepts string
	Reader   string
}

func main() {

	session, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	log.Println("Sunshine list")
	file, err := ioutil.ReadFile("./employees.json")
	if err != nil {
		log.Fatal(err)
	}

	allEmployees := make(map[string][]model.Position)

	//set up a goroutine that will listen a store employees in a map
	//we should probably write to mongo directly instead of updating
	//a huge map and writing the map all at once
	out := make(chan model.Employee)
	go func() {
		for s := range out {
			key := strings.ToLower(s.LastName) + seperator + strings.ToLower(s.FirstName)
			postitions, ok := allEmployees[key]
			if ok {
				postitions = append(postitions, s.P)
			} else {
				postitions = []model.Position{s.P}
			}
			allEmployees[key] = postitions
		}
	}()

	var employees []employeeData
	err = json.Unmarshal(file, &employees)
	if err != nil {
		log.Fatal(err)
	}

	//go through are input data and parse out data
	//our input is setup as a list of links to parse the
	//data out of, and list is keyed by year
	for _, v := range employees {

		year := v.Year
		cpi := v.Cpi
		prod, err := getEmployeeProducer(v.Reader)
		if err != nil {
			log.Println(err)
			continue
		}

		var wg sync.WaitGroup
		//for each department for a given year
		//we will parse out the data.
		//we are using a goroutine per department, within a year
		for _, d := range v.Dept {
			deptName := d.Name
			link := d.Link
			pages := d.Pages
			wg.Add(1)
			go func(l string, y int, dN string, p int, c float64, bc float64, ch chan model.Employee) {
				defer wg.Done()
				//if there are multiple pages, retrieve data for each one
				if pages > 0 {
					for i := 1; i <= pages; i++ {
						pagedLink := fmt.Sprintf("%s&page=%d", l, i)
						prod.ProduceEmployee(pagedLink, y, dN, c, bc, ch)
					}
				} else {
					prod.ProduceEmployee(l, y, dN, c, bc, ch)
				}

			}(link, year, deptName, pages, cpi, cpi96, out)

		}
		wg.Wait()
	}

	close(out)

	//after we've parsed all the data, send it to the DB
	for k, v := range allEmployees {
		s := strings.Split(k, seperator)
		firstName := s[1]
		lastName := s[0]

		//anonymous struct to instert
		data := struct {
			FirstName string
			LastName  string
			Positions []model.Position
		}{
			firstName,
			lastName,
			v,
		}
		c := session.DB("test").C("people")
		err = c.Insert(&data)
		if err != nil {
			log.Fatal(err)
		}

	}

}

//we have different types of employee producers depending on the data
//i.e whether we need to scrape the data or just parse JSON
func getEmployeeProducer(reader string) (producer.EmployeeProducer, error) {
	if reader == "parse" {
		return producer.Parser{Simple: false}, nil
	}
	if reader == "scrape" {
		return producer.Scraper{}, nil
	}
	if reader == "parse-simple" {
		return producer.Parser{Simple: true}, nil
	}
	return nil, fmt.Errorf("no reader found for type: %s", reader)

}
