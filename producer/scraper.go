package producer

import (
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/nstehr/sunshine/model"
)

// Scraper is the type we can use to implement the Producer interface
type Scraper struct{}

// ProduceEmployee will use scraping to parse the specified link and will out Employee data to the out channel
func (Scraper) ProduceEmployee(link string, year int, category string, cpi float64, baseCPI float64, out chan model.Employee) {
	resp, err := http.Get(link)
	if err != nil {
		log.Fatal(err)
	}
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		log.Fatal(err)
	}
	employeeTable := doc.Find("table").First()
	employeeTable.Find("tbody tr").Each(func(i int, tableRow *goquery.Selection) {
		employee := model.Employee{}
		position := model.Position{Year: year, Category: category}
		tableRow.Find("td").Each(func(j int, data *goquery.Selection) {
			switch j {
			case 0:
				var employer string
				if data.Children().Length() == 1 {
					employer = data.Children().Eq(0).Text()
				} else {
					employer = data.Text()
				}
				employer = strings.Split(employer, "/")[0]
				employer = strings.TrimSpace(employer)
				position.Employer = employer
			case 1:
				employee.LastName = strings.TrimSpace(data.Text())
			case 2:
				employee.FirstName = strings.TrimSpace(data.Text())
			case 3:
				positionStr := data.Text()
				positionStr = strings.Split(positionStr, "/")[0]
				positionStr = strings.TrimSpace(positionStr)
				position.Title = positionStr
			case 4:
				salary := moneyStringToFloat(data.Text())
				position.Salary = salary
				position.AdjustedSalary = adjustMoney(salary, cpi, baseCPI)
			case 5:
				benefits := moneyStringToFloat(data.Text())
				position.Benefits = benefits
				position.AdjustedBenefits = adjustMoney(benefits, cpi, baseCPI)
			}

		})
		employee.P = position
		out <- employee
	})

}
