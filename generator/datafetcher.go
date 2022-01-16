package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type WeekData struct {
	Weeks []Week `json:Weeks`
}

type Week struct {
	WeekNumber int   `json:WeekNumber`
	Days       []Day `json:Days`
}

type Day struct {
	DayMenuDate string    `json:DayMenuDate`
	DayMenus    []DayMenu `json:DayMenus`
}

type DayMenu struct {
	DayMenuName         string `json:DayMenuName`
	MenuAlternativeName string `json:MenuAlternativeName`
}

type DayOut struct {
	Menu string `json:value`
	Date string `json:date`
	Vego bool   `json:vego`
}

func main() {
	menu := fetchMenu()
	data := parseMenuData(menu)

	res, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}

	ioutil.WriteFile("out.json", res, 0777)
}

func parseMenuData(menu []byte) []DayOut {
	data := WeekData{}
	err := json.Unmarshal(menu, &data)
	if err != nil {
		log.Fatal(err)
	}

	parsed := []DayOut{}

	for _, week := range data.Weeks {
		for _, day := range week.Days {
			dayOut := DayOut{
				Vego: false,
				Date: day.DayMenuDate,
			}

			menu := []byte{}

			for _, dayMenu := range day.DayMenus {
				menu = append(menu, (dayMenu.MenuAlternativeName + ": " + dayMenu.DayMenuName + "\n")...)
			}
			// Remove \n at the end
			menu = menu[0 : len(menu)-1]
			dayOut.Menu = string(menu)

			// Guess if vegetarian
			if len(day.DayMenus) == 1 && day.DayMenus[0].MenuAlternativeName == "Dagens gröna" {
				dayOut.Vego = true
			}

			parsed = append(parsed, dayOut)
		}
	}

	return parsed
}

func fetchMenu() []byte {
	resp, err := http.Get("https://mpi.mashie.com/public/menu/uppsala+kommun/00e57ed5?country=se")
	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	startDelim := "var weekData = "
	start := strings.Index(string(body), startDelim) + len(startDelim)
	end := strings.Index(string(body), "\n</script>")
	menu := body[start:end]

	// Change date format from JS to JSON compatible
	rgx := regexp.MustCompile(`new Date\(([0-9]+)\)`)
	rpl := []byte(`"${1}"`)
	menu = rgx.ReplaceAll(menu, rpl)

	return menu

	/*json, err := ioutil.ReadFile("in.json")
	if err != nil {
		log.Fatal(err)
	}

	return json*/
}
