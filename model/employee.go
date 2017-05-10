package model

type Employee struct {
	FirstName string
	LastName  string
	P         Position
}

type Position struct {
	Title            string
	Employer         string
	Category         string
	Year             int
	Salary           float64
	AdjustedSalary   float64
	Benefits         float64
	AdjustedBenefits float64
}
