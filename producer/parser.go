package producer

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/nstehr/sunshine/model"
)

type transformer func(*json.RawMessage) (model.Employee, error)

// Parser is the type we can use to implement the Producer interface
type Parser struct {
	Simple bool
}

//we have two types of JSON we can parse, one that is straightforward (Simple)
//and one that uses a nested structure of columns (complexJSON)
type simpleJSON struct {
	Sector    string `json:"Sector"`
	LastName  string `json:"Last Name"`
	FirstName string `json:"First Name"`
	Salary    string `json:"Salary Paid"`
	Benefit   string `json:"Taxable Benefits"`
	Employer  string `json:"Employer"`
	Title     string `json:"Job Title"`
}

type complexJSON struct {
	Sector    column `json:"sector"`
	LastName  column `json:"last_name"`
	FirstName column `json:"first_name"`
	Salary    column `json:"salary_paid"`
	Benefit   column `json:"taxable_benefits"`
	Employer  column `json:"employer"`
	Title     column `json:"job_title"`
	Position  column `json:"position"`
}

type column struct {
	ColumnName string
	Content    string
}

// ProduceEmployee will parse actual JSON from the server and output an Employee onto the out channel
func (p Parser) ProduceEmployee(link string, year int, category string, cpi float64, baseCPI float64, out chan model.Employee) {
	var myClient = &http.Client{}

	r, err := myClient.Get(link)
	if err != nil {
		log.Println("Error Retrieving data", err)
	}
	defer r.Body.Close()

	body, _ := ioutil.ReadAll(r.Body)
	var data []*json.RawMessage
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Println("Error parsing data", err)
	}
	var transformFunc transformer
	if p.Simple {
		transformFunc = transformSimple
	} else {
		transformFunc = transform
	}
	for _, raw := range data {
		e, err := transformFunc(raw)
		if err != nil {
			log.Println("Error transforming data", err)
		}

		e.P.AdjustedBenefits = adjustMoney(e.P.Benefits, cpi, baseCPI)
		e.P.AdjustedSalary = adjustMoney(e.P.Salary, cpi, baseCPI)
		e.P.Year = year
		out <- e
	}
}

func transformSimple(raw *json.RawMessage) (model.Employee, error) {
	var data simpleJSON
	err := json.Unmarshal(*raw, &data)
	if err != nil {
		return model.Employee{}, err
	}
	e := model.Employee{}
	p := model.Position{}
	e.FirstName = data.FirstName
	e.LastName = data.LastName

	p.Benefits = moneyStringToFloat(data.Benefit)
	p.Category = data.Sector
	p.Employer = data.Employer
	p.Salary = moneyStringToFloat(data.Salary)
	p.Title = data.Title
	e.P = p
	return e, nil
}

func transform(raw *json.RawMessage) (model.Employee, error) {
	var data *complexJSON
	err := json.Unmarshal(*raw, &data)
	if err != nil {
		return model.Employee{}, err
	}

	e := model.Employee{}
	p := model.Position{}
	e.FirstName = data.FirstName.Content
	e.LastName = data.LastName.Content
	p.Category = data.Sector.Content
	p.Benefits = moneyStringToFloat(data.Benefit.Content)
	p.Employer = data.Employer.Content
	p.Salary = moneyStringToFloat(data.Salary.Content)
	if data.Title.Content != "" {
		p.Title = data.Title.Content
	} else {
		p.Title = data.Position.Content
	}

	e.P = p
	return e, nil
}
