package producer

import (
	"log"
	"strconv"
	"strings"

	"github.com/nstehr/sunshine/model"
)

//EmployeeProducer defines an interface for taking in a link to data, and then producing a model.Employee
type EmployeeProducer interface {
	ProduceEmployee(link string, year int, category string, cpi float64, baseCPI float64, out chan model.Employee)
}

func adjustMoney(money float64, cpi float64, baseCPI float64) float64 {
	adjusted := (money / cpi) * baseCPI
	return adjusted
}

func moneyStringToFloat(moneyStr string) float64 {
	if moneyStr == "" {
		return 0
	}
	m := strings.Trim(moneyStr, "$")
	m = strings.Replace(m, ",", "", -1)

	money, err := strconv.ParseFloat(m, 64)
	if err != nil {
		log.Println(err)
		return 0
	}
	return money
}
