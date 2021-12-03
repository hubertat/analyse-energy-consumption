package main

import (
	"fmt"
	"strconv"
	"time"
)

type DailyUse struct {
	EnergyHourly [24]uint64

	Day time.Time
}

func (du *DailyUse) DayIndex() int {
	if du.Day.IsZero() {
		return 0
	}
	return du.Day.YearDay() + 1000*du.Day.Year()
}

func (du *DailyUse) Weekday() time.Weekday {
	return du.Day.Weekday()
}

func (du *DailyUse) getZone(hour int) int {
	switch du.Weekday() {
	case time.Sunday:
		return 0
	case time.Saturday:
		return 0
	default:
		if du.Day.Month() > 3 && du.Day.Month() < 10 {
			if hour > 6 && hour < 13 {
				return 1
			}
			if hour > 18 && hour < 22 {
				return 2
			}
			return 0
		} else {
			if hour > 6 && hour < 13 {
				return 1
			}
			if hour > 15 && hour < 21 {
				return 2
			}
			return 0
		}
	}
}

func (du *DailyUse) TotalEnergy() (sum uint64) {
	for _, hourly := range du.EnergyHourly {
		sum += hourly
	}

	return
}

func (du *DailyUse) EnergyZones() (sum [3]uint64) {
	for hour, hourly := range du.EnergyHourly {
		sum[du.getZone(hour)] += hourly
	}
	return
}

type EnergyUse struct {
	Days []*DailyUse

	Start time.Time
	Stop  time.Time
}

func (eu *EnergyUse) GetFromInflux(influx *Influx) error {
	query, err := influx.energyByHourQuery(eu.Start, eu.Stop)
	if err != nil {
		return fmt.Errorf("preparing influx query failed: %v", err)
	}

	tableResult, err := influx.readQuery(query)
	if err != nil {
		return fmt.Errorf("running influx query failed: %v", err)
	}

	dailyMap := map[int]*DailyUse{}

	for tableResult.Next() {
		if tableResult.Err() != nil {
			return fmt.Errorf("parsing table result failed: %v", tableResult.Err())
		}
		energyValue, err := strconv.ParseUint(fmt.Sprint(tableResult.Record().Value()), 10, 64)
		if err != nil {
			return fmt.Errorf("parsing Record().Value() to float failed: %v", err)
		}
		when := tableResult.Record().Time()
		dayIndex := when.YearDay() + when.Year()*1000
		daily, exist := dailyMap[dayIndex]
		if exist {
			daily.EnergyHourly[when.Hour()] = energyValue
		} else {
			dailyMap[dayIndex] = &DailyUse{Day: time.Date(when.Year(), when.Month(), when.Day(), 0, 0, 0, 0, when.Location())}
			dailyMap[dayIndex].EnergyHourly[when.Hour()] = energyValue
			eu.Days = append(eu.Days, dailyMap[dayIndex])
		}
	}
	return nil
}

func (eu *EnergyUse) GetAveragePercentages() (percentage [3]float64) {
	var sum [3]uint64
	for _, daily := range eu.Days {
		zones := daily.EnergyZones()
		sum[0] += zones[0]
		sum[1] += zones[1]
		sum[2] += zones[2]
	}
	totalSum := sum[0] + sum[1] + sum[2]
	percentage[0] = float64(sum[0]) / float64(totalSum) * 100
	percentage[1] = float64(sum[1]) / float64(totalSum) * 100
	percentage[2] = float64(sum[2]) / float64(totalSum) * 100
	return
}
