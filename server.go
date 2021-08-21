package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"
)

type SubscriptionUpdate struct {
	OldEndpoint string
	WebSub      *webpush.Subscription
}

// NotificationContents Contains all the information that is sent through web push
type NotificationContents struct {
	Title  string
	Text   string
	Image  string
	Name   string // The name of the reciever
	Action string
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

		// Go to 15:00 day before
		warnPoint = warnPoint.Add(-9 * time.Hour)

		if warnPoint.After(now) {
			warnAtPoint(warnPoint)
			return
		}
	}

}

func warnAtPoint(point time.Time) {
	waitDur := point.Sub(time.Now())

	t := time.NewTimer(waitDur)
	<-t.C

	warnAllUsers()
}

func warnAllUsers() {
	for _, sub := range subs {
		sendNotificationToUser([]byte("WARNING: Vegetarian lunch tomorrow"), &sub)
	}

	// Wait for next warning
	awaitNextVegoDay()
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
		format := "2006-01-02"
		cut := string([]byte(rd.Date)[0:len(format)])

		dt, err := time.Parse(format, cut)
		if err != nil {
			fmt.Println(err)
		}

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
	var err error
	vapidPrivate, vapidPublic, err = webpush.GenerateVAPIDKeys()
	if err != nil {
		fmt.Println(err)
	}
}

func respondToSubscribePush(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		serveInternalError(w, r)
		return
	}

	// Store subscription
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
