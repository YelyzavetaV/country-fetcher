package models

type Country struct {
	Name       string `json:"name"`
	Population int    `json:"population"`
	Region     string `json:"region"`
	Capital    string `json:"capital"`
}

type Region struct {
	Name            string
	Countries       []Country
	TotalPopulation int
	AvgPopulation   float64
}