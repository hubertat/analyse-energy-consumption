package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
)

func LoadConfig(path string, target interface{}) error {

	configFile, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot read config file (%s) error: %v", path, err)
	}

	err = json.Unmarshal([]byte(configFile), target)
	if err != nil {
		return fmt.Errorf("unmarshal failed: %v", err)
	}
	return nil
}

func main() {
	influx := &Influx{}
	err := LoadConfig("influx.json", influx)
	if err != nil {
		panic(err)
	}

	now := time.Now()
	energy := EnergyUse{}
	energy.Start = time.Date(now.Year(), now.Month(), now.Day()-120, 0, 0, 0, 0, now.Location())
	energy.Stop = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	err = energy.GetFromInflux(influx)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Loaded %d days, from %s to %s\n\n", len(energy.Days), energy.Start.String(), energy.Stop.String())

	fmt.Print("Day: weekday\t: energy total/offpeak/peak\t : peak/off\n")
	for _, daily := range energy.Days {
		total := daily.TotalEnergy()
		peak := daily.PeakEnergy()
		off := daily.OffPeakEnergy()
		fmt.Printf("Day: %v\t:%d %d %d\t%f %f\n", daily.Weekday(), total, off, peak, float64(peak)/float64(total)*100, float64(off)/float64(total)*100)
	}

	fmt.Println(energy.GetAveragePeakPercentage())
}
