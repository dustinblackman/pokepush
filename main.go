package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/buger/jsonparser"
	"github.com/kellydunn/golang-geo"
	"github.com/parnurzeal/gorequest"
	"github.com/robfig/cron"
	"github.com/x-cray/logrus-prefixed-formatter"
	"github.com/xconstruct/go-pushbullet"
)

// Places contains the name, point, and polygon variables for every location defined in configs.
type Places struct {
	Name    string
	Point   *geo.Point
	Polygon *geo.Polygon
}

// User contains the the Pushbullet key and subscribed locations for a user.
type User struct {
	PushBulletKey string
	Places        []string
}

const (
	pointRange = 0.04 // Kilometers
)

var (
	degrees = []float64{0, 45, 90, 135, 180, 225, 270, 315}
	places  = []*Places{}
	usedIDs = map[string]int64{}
	users   = []User{}
	version = "dev"
)

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// createPlace creates a Places based of name, lat, and lon.
func createPlace(name string, lat, lon float64) *Places {
	point := geo.NewPoint(lat, lon)
	points := []*geo.Point{}

	for _, degree := range degrees {
		points = append(points, point.PointAtDistanceAndBearing(pointRange, degree))
	}
	polygon := geo.NewPolygon(points)
	return &Places{name, point, polygon}
}

// createMessage creates the message for all found pokemon
func createMessage(pokemon []string) string {
	l := len(pokemon)
	if l == 1 {
		return pokemon[0]
	}

	wordsForSentence := make([]string, l)
	copy(wordsForSentence, pokemon)
	wordsForSentence[l-1] = "and " + wordsForSentence[l-1]
	return strings.Join(wordsForSentence, ", ")
}

// sendToDevices interates over every user to verify if they are subscribed to a location that's detected a pokemon, and then sends a message.
func sendToDevices(place string, pokemon []string) {
	for _, user := range users {
		if stringInSlice(place, user.Places) {
			pb := pushbullet.New(user.PushBulletKey)
			devices, _ := pb.Devices()
			for _, device := range devices {
				pb.PushNote(device.Iden, "Pokemon detected at "+place, createMessage(pokemon))
			}
		}
	}
}

// scanPokevision hits the pokevision API for all subscribed locations, and calls sendToDevices when found.
// TODO: Error handling
func scanPokevision(place *Places) []string {
	lat := strconv.FormatFloat(place.Point.Lat(), 'f', -1, 64)
	lon := strconv.FormatFloat(place.Point.Lng(), 'f', -1, 64)

	_, body, _ := gorequest.New().Get(fmt.Sprintf("https://pokevision.com/map/scan/%s/%s", lat, lon)).End()
	jobID, _ := jsonparser.GetString([]byte(body), "jobId")
	time.Sleep(5000 * time.Millisecond)
	// TODO Handle jobs in progress
	_, body, _ = gorequest.New().Get(fmt.Sprintf("https://pokevision.com/map/data/%s/%s/%s", lat, lon, jobID)).End()

	available := []string{}

	pokemon, _, _, _ := jsonparser.Get([]byte(body), "pokemon")
	jsonparser.ArrayEach(pokemon, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		pokemonIDInt, _ := jsonparser.GetInt(value, "pokemonId")
		pokemonID := strconv.FormatInt(pokemonIDInt, 10)
		id, _ := jsonparser.GetString(value, "id")
		expirationTime, _ := jsonparser.GetInt(value, "expiration_time")
		lat, _ := jsonparser.GetFloat(value, "latitude")
		lon, _ := jsonparser.GetFloat(value, "longitude")
		point := geo.NewPoint(lat, lon)

		if _, ok := usedIDs[id]; !ok && place.Polygon.Contains(point) {
			available = append(available, pokemonIDs[pokemonID])
			usedIDs[id] = expirationTime
		}
	})

	fmt.Println(available)

	uniqMap := map[string]bool{}
	uniqAvailable := []string{}
	for _, name := range available {
		uniqMap[name] = true
	}
	for key := range uniqMap {
		uniqAvailable = append(uniqAvailable, key)
	}

	return uniqAvailable
}

// searchPlaces checks all subscribed locations.
func searchPlaces() {
	logrus.Info("Searching places")
	for _, place := range places {
		availablePokemon := scanPokevision(place)
		if len(availablePokemon) >= 1 {
			logrus.Info("Found pokemon at " + place.Name)
			logrus.Info(availablePokemon)
			sendToDevices(place.Name, availablePokemon)
		}
	}
}

// cleanExpiredIDs clears the expired pokemon IDs cache
func cleanExpiredIDs() {
	for id, expireTime := range usedIDs {
		if time.Unix(expireTime, 0).Before(time.Now()) {
			delete(usedIDs, id)
		}
	}
}

// loadConfig loads the config.json file.
// TODO: Error handling
func loadConfig() {
	config, _ := ioutil.ReadFile("./config.json")
	logrus.SetFormatter(new(prefixed.TextFormatter))

	logtype, _ := jsonparser.GetString(config, "log")
	if logtype == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}

	logrus.Info("Pokepush v" + version)

	placesJSON, _, _, _ := jsonparser.Get(config, "places")
	jsonparser.ArrayEach(placesJSON, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		name, _ := jsonparser.GetString(value, "name")
		lat, _ := jsonparser.GetFloat(value, "lat")
		lon, _ := jsonparser.GetFloat(value, "lon")
		places = append(places, createPlace(name, lat, lon))
	})

	usersJSON, _, _, _ := jsonparser.Get(config, "users")
	jsonparser.ArrayEach(usersJSON, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		pushbulletKey, _ := jsonparser.GetString(value, "pushbullet_key")
		userPlacesJSON, _, _, _ := jsonparser.Get(value, "places")

		userPlaces := []string{}
		jsonparser.ArrayEach(userPlacesJSON, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			place, _ := jsonparser.ParseString(value)
			userPlaces = append(userPlaces, place)
		})

		users = append(users, User{pushbulletKey, userPlaces})
	})
}

func main() {
	loadConfig()
	c := cron.New()
	c.AddFunc("@every 2m", searchPlaces)
	c.AddFunc("@every 10m", cleanExpiredIDs)
	searchPlaces()

	logrus.Info("Starting cron")
	c.Start()

	select {} // Block forever.
}
