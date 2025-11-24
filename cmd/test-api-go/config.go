package main

type Config struct {
	JobCron            string
	ConnectionString   string
	CurrencyApiBaseUrl string
	Currencies         string
	DaysLookBack       int
	AppPort            int
}
