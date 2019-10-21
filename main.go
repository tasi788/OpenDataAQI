package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// AQI parse json to struct.
type AQI []struct {
	SiteName    string  `json:"SiteName"`
	County      string  `json:"County"`
	AQI         string  `json:"AQI"`
	Pollutant   string  `json:"Pollutant"`
	Status      string  `json:"Status"`
	SO2         string  `json:"SO2"`
	CO          string  `json:"CO"`
	CO8Hr       string  `json:"CO_8hr"`
	O3          string  `json:"O3"`
	O38Hr       string  `json:"O3_8hr"`
	PM10        string  `json:"PM10"`
	PM25        string  `json:"PM2.5"`
	NO2         string  `json:"NO2"`
	NOx         string  `json:"NOx"`
	NO          string  `json:"NO"`
	WindSpeed   string  `json:"WindSpeed"`
	WindDirec   string  `json:"WindDirec"`
	PublishTime EpaDate `json:"PublishTime"`
	PM25AVG     string  `json:"PM2.5_AVG"`
	PM10AVG     string  `json:"PM10_AVG"`
	SO2AVG      string  `json:"SO2_AVG"`
	Longitude   string  `json:"Longitude"`
	Latitude    string  `json:"Latitude"`
	SiteID      string  `json:"SiteId"`
}

// EpaDate struct to time
type EpaDate struct {
	time.Time
}

// UnmarshalJSON loads specfic datetime
func (sd *EpaDate) UnmarshalJSON(input []byte) error {
	strInput := string(input)
	strInput = strings.Trim(strInput, `"`)
	location, _ := time.LoadLocation("Asia/Taipei")
	layout := "2006-01-02 15:04"
	newTime, err := time.ParseInLocation(layout, strInput, location)
	if err != nil {
		return err
	}
	sd.Time = newTime
	return nil
}

func main() {
	// init mongodb connection
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))

	if err != nil {
		log.Fatal(err)
	}

	fetchStatus, Data := fetchData()
	log.Println(fetchStatus)

	if fetchStatus != true {
		log.Println("Fetch Error.")
		return
	}
	collection := client.Database("Opendata").Collection("Air")
	collection.Drop(ctx)

	for perSite := range Data {
		collection := client.Database("Opendata").Collection("Air")
		Site := Data[perSite]

		filter := bson.D{{"publishtime", Site.PublishTime}, {"sitename", Site.SiteName}}
		update := bson.D{{"$set",
			Site,
		}}
		upsert := options.Update().SetUpsert(true)
		updateresult, err := collection.UpdateOne(ctx, filter, update, upsert)
		if err != nil {
			log.Fatal(err)
		} else {
			log.Println(updateresult)
		}

	}

}

func fetchData() (bool, AQI) {
	url := "https://opendata.epa.gov.tw/api/v1/AQI?format=json"
	resp, err := http.Get(url)

	var Aqi AQI
	// except err or can't process
	if err != nil {
		log.Println(err)
		return false, Aqi
	}
	if resp.StatusCode != 200 {
		log.Println(err)
		return false, Aqi
	}
	defer resp.Body.Close()

	loads, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(loads, &Aqi)
	if err != nil {
		log.Println(err)
		return false, Aqi
	}
	return true, Aqi
}
