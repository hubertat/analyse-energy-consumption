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

func (du *DailyUse) isPeak(hour int) bool {
	switch du.Weekday() {
	case time.Sunday:
		return false
	case time.Saturday:
		return false
	default:
		return (hour > 6 && hour < 13) || (hour > 15 && hour < 22)
	}
}

func (du *DailyUse) TotalEnergy() (sum uint64) {
	for _, hourly := range du.EnergyHourly {
		sum += hourly
	}

	return
}

func (du *DailyUse) OffPeakEnergy() (sum uint64) {
	for hour, hourly := range du.EnergyHourly {
		if !du.isPeak(hour) {
			sum += hourly
		}
	}
	return
}

func (du *DailyUse) PeakEnergy() (sum uint64) {
	for hour, hourly := range du.EnergyHourly {
		if du.isPeak(hour) {
			sum += hourly
		}
	}
	return
}

func (du *DailyUse) GetPeakPercentage() float64 {
	return float64(du.PeakEnergy()) / float64(du.TotalEnergy())
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

func (eu *EnergyUse) GetAveragePeakPercentage() float64 {
	var totalEnergy uint64
	var peakEnergy uint64
	for _, daily := range eu.Days {
		totalEnergy += daily.TotalEnergy()
		peakEnergy += daily.PeakEnergy()
	}
	return float64(peakEnergy) / float64(totalEnergy) * 100
}
