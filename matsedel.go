package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"
)

// TODO allow updating web push subscriptions
type SubscriptionUpdate struct {
	OldEndpoint string
	WebSub      *webpush.Subscription
}

type DayMenuRaw struct {
	Menu string
	Date string
	Vego bool
}

type DayMenu struct {
	Menu string
	Date time.Time
	Vego bool
}

var subs []webpush.Subscription
var days []DayMenu

var vapidPublic string
var vapidPrivate string

func main() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	http.HandleFunc("/vapid/", respondToGetVapidPublic)
	http.HandleFunc("/subscribepush/", respondToSubscribePush)

	initWebPush()
	readMenuData()
	go awaitNextVegoDay()

	fmt.Println("Listening on :5619...")
	err := http.ListenAndServe(":5619", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func awaitNextVegoDay() {
	now := time.Now()
	if now.Hour() > 15 {
		now.Add(24 * time.Hour)
	}

	for _, day := range days {
		if !day.Vego {
			continue
		}

		warnPoint := day.Date

		// Go to 17:00 UTC the day before, which is 15:00 in Swedish summer time
		warnPoint = warnPoint.Add(17 * time.Hour)

		if warnPoint.After(now) {
			warnAtPoint(warnPoint)
			return
		}
	}
}

func warnAtPoint(point time.Time) {
	waitDur := point.Sub(time.Now())

	fmt.Println("Will warn at ", point)

	t := time.NewTimer(waitDur)
	<-t.C

	warnAllUsers()

	// Wait for next warning
	awaitNextVegoDay()
}

func warnAllUsers() {
	for _, sub := range subs {
		sendNotificationToUser([]byte("WARNING: Vegetarian lunch tomorrow"), &sub)
	}
}

func readMenuData() {
	data, err := ioutil.ReadFile("data.json")
	if err != nil {
		log.Fatal(err)
	}

	var raw []DayMenuRaw
	err = json.Unmarshal(data, &raw)
	if err != nil {
		log.Fatal(err)
	}

	for _, rd := range raw {
		// Base 0 lol
		ms, err := strconv.ParseInt(rd.Date, 0, 64)
		if err != nil {
			fmt.Println(err)
		}
		dt := time.UnixMilli(ms)

		dm := DayMenu{
			Vego: rd.Vego,
			Menu: rd.Menu,
			Date: dt,
		}
		days = append(days, dm)
	}
}

func respondToGetVapidPublic(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, vapidPublic)
}

func initWebPush() {
	// Read vapid keys or generate if they don't exist
	fileName := "vapid keys"
	stored, err := os.ReadFile(fileName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			vapidPrivate, vapidPublic, err = webpush.GenerateVAPIDKeys()
			if err != nil {
				log.Fatal(err)
			}
			data := []byte(vapidPrivate + "\n" + vapidPublic)
			os.WriteFile(fileName, data, 0666)
		} else {
			log.Fatal(err)
		}
	} else {
		keys := strings.Split(string(stored), "\n")
		vapidPrivate = keys[0]
		vapidPublic = keys[1]
	}

	// Read push subscribers from disk
	content, err := os.ReadFile("webpushsubs")
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Fatal(err)
		} else {
			fmt.Println("webpushsubs file does not exist. I'll skip reding from it.")
		}
	}

	wpsubs := strings.Split(string(content), "\n")
	for _, wpsub := range wpsubs[0 : len(wpsubs)-1] {
		sub := &webpush.Subscription{}
		json.Unmarshal([]byte(wpsub), sub)
		subs = append(subs, *sub)
	}
}

func respondToSubscribePush(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		serveInternalError(w, r)
		return
	}

	// Store subscription on disk
	fileName := "webpushsubs"
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}
	defer file.Close()
	_, err = file.Write([]byte(string(body) + "\n"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}

	// Also store it in RAM
	sub := &webpush.Subscription{}
	json.Unmarshal(body, sub)
	subs = append(subs, *sub)

	sendNotificationToUser([]byte("if you see this, WOW!!"), sub)

	// Respond with success
	fmt.Fprint(w, "▼・ᴥ・▼")
}

func serveInternalError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprint(w, "Internal server error")
}

func sendNotificationToUser(message []byte, sub *webpush.Subscription) {
	resp, err := webpush.SendNotification(message, sub, &webpush.Options{
		Subscriber:      "ebinbellini@airmail.cc",
		VAPIDPublicKey:  vapidPublic,
		VAPIDPrivateKey: vapidPrivate,
		TTL:             30,
	})
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
}
