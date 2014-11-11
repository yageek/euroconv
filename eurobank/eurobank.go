package eurobank

import (
	"encoding/xml"
	"io"
	"net/http"
	"strconv"
	"time"
)

const (
	EuroBankDayRateURL = "http://www.ecb.europa.eu/stats/eurofxref/eurofxref-daily.xml"
)

type Currency struct {
	Rate float32
	Id   string
}

type DayRate struct {
	Day   time.Time
	Rates []Currency
}

func newRateFromXMLInput(r io.Reader) *DayRate {

	var dayRate *DayRate = nil

	decoder := xml.NewDecoder(r)
	for {
		token, _ := decoder.Token()
		if token == nil {
			break
		}
		switch xmlNode := token.(type) {

		case xml.StartElement:
			if xmlNode.Name.Local == "Cube" {

				if len(xmlNode.Attr) == 1 {

					dayTimeValue := xmlNode.Attr[0].Value
					dayTime, _ := time.Parse("2006-01-02", dayTimeValue)
					dayRate = &DayRate{Day: dayTime}
				} else if len(xmlNode.Attr) == 2 {

					rate, _ := strconv.ParseFloat(xmlNode.Attr[1].Value, 32)

					currency := Currency{Id: xmlNode.Attr[0].Value, Rate: float32(rate)}
					dayRate.Rates = append(dayRate.Rates, currency)
				}
			}
		}
	}
	return dayRate
}

func GetDayRate() (*DayRate, error) {

	resp, err := http.Get(EuroBankDayRateURL)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return newRateFromXMLInput(resp.Body), nil

}
