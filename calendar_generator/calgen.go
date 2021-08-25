package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"time"
)

type DayMenuRaw struct {
	Menu string
	Date string
	Vego bool
}

type DayMenu struct {
	Menu       string
	Date       time.Time
	Vego       bool
	DayOfMonth int
}

type MenuOut struct {
	Weeks []Week
}

type Week struct {
	Days []DayMenu
}

func main() {
	menuData := readMenuData()
	weeks := menuDataWeeks(menuData)
	out := MenuOut{Weeks: weeks}

	tmpl, err := template.ParseFiles("./template.html")
	if err != nil {
		log.Fatal(err)
	}

	cal, err := os.Create("./calendar.html")
	if err != nil {
		log.Fatal(err)
	}

	err = tmpl.ExecuteTemplate(cal, "kalender", out)
	if err != nil {
		log.Fatal(err)
	}
}

func readMenuData() []DayMenu {
	data, err := ioutil.ReadFile("data.json")
	if err != nil {
		log.Fatal(err)
	}

	var days []DayMenu

	var raw []DayMenuRaw
	err = json.Unmarshal(data, &raw)
	if err != nil {
		log.Fatal(err)
	}

	for _, rd := range raw {
		format := "2006-01-02"
		cut := string([]byte(rd.Date)[0:len(format)])

		dt, err := time.Parse(format, cut)
		if err != nil {
			fmt.Println(err)
		}

		dm := DayMenu{
			Vego:       rd.Vego,
			Menu:       rd.Menu,
			Date:       dt,
			DayOfMonth: dt.Day(),
		}
		days = append(days, dm)
	}

	return days
}

func menuDataWeeks(menus []DayMenu) []Week {
	// Orders the menu data by weeks

	weeks := []Week{}

	week := []DayMenu{}
	for _, menu := range menus {
		// NOT TESTED LOL BECAUSE MENU STARTS ON A MONDAY
		if len(week) == 0 {
			if menu.Date.Weekday() != time.Monday {
				// Can only be tuesday-friday

				// If it's wednesday, daysToAdd will be 2, representing monday
				// and tuesday
				daysToAdd := int(menu.Date.Weekday() - time.Monday)
				for i := daysToAdd; i > 0; i-- {
					// Go back some days
					dummyTime := menu.Date.Add(-time.Duration(i) * 24 * time.Hour)
					// Add empty days
					week = append(week, DayMenu{
						DayOfMonth: int(dummyTime.Day()),
					})
				}
			}
		}

		week = append(week, menu)

		if len(week) == 5 {
			lastTime := week[len(week)-1].Date
			for i := 1; i <= 2; i++ {
				// Go back some days
				dummyTime := lastTime.Add(time.Duration(i) * 24 * time.Hour)
				// Add empty days
				week = append(week, DayMenu{
					DayOfMonth: int(dummyTime.Day()),
				})
			}

			// Insert this finished week
			weeks = append(weeks, Week{Days: week})
			week = []DayMenu{}
		}
	}

	return weeks
}
