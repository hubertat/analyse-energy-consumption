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
	energy.Start = time.Date(now.Year(), now.Month(), now.Day()-200, 0, 0, 0, 0, now.Location())
	energy.Stop = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	err = energy.GetFromInflux(influx)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Loaded %d days, from %s to %s\n\n", len(energy.Days), energy.Start.String(), energy.Stop.String())

	fmt.Print("Day: weekday\t: energy offpeak/peak1/peak2\t : peak1[%]/peak2[%]\n")
	for _, daily := range energy.Days {
		zones := daily.EnergyZones()
		sum := zones[0] + zones[1] + zones[2]
		fmt.Printf("Day: %v\t:%d %d %d\t%f %f\n", daily.Weekday(), zones[0], zones[1], zones[2], float64(zones[1])/float64(sum)*100, float64(zones[2])/float64(sum)*100)
	}

	fmt.Println("średnice procentowe udziały stref: [poza szczyt, szczyt 1, szczyt 1]")
	fmt.Println(energy.GetAveragePercentages())
}
