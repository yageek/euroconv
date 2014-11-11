package cache

import (
	"fmt"
	"net/url"
	"github.com/garyburd/redigo/redis"
	"github.com/yageek/euroconv/eurobank"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const lastRecordField = "LASTRECORD"

var cacheConn redis.Conn


func init() {

	redisFullURL := os.Getenv("REDIS_URL")
	log.Printf("Starting redis at %s\n", redisFullURL)

	redURL, _ := url.Parse(redisFullURL)
	address := redURL.Host
	password := ""

	userInfo := redURL.User
	if userInfo != nil {
		password, _ = userInfo.Password()
	}

	c, err := redis.Dial("tcp", address)

	if err != nil {
		log.Panicln("Cache is down: %v\n", err)
	}

	if password != "" {
		if _, err := c.Do("AUTH", password); err != nil {
			c.Close()
			log.Panicln(err)
		}
	}

	cacheConn = c

}


func SetDayRate(d *eurobank.DayRate) error {

	hash := fmt.Sprintf("dayrate:%d-%d-%d", d.Day.Year(), d.Day.Month(), d.Day.Day())
	args := make(map[string]string)

	for _, rate := range d.Rates {
		args[rate.Id] = strconv.FormatFloat(float64(rate.Rate), 'f', 4, 32)
	}
	_, err := cacheConn.Do("HMSET", redis.Args{}.Add(hash).AddFlat(args)...)
	if err != nil {
		return err
	}
	_, err = cacheConn.Do("SET", lastRecordField, hash)

	return err
}

func GetDayRate() *eurobank.DayRate {

	r, err := redis.String(cacheConn.Do("GET", lastRecordField))
	if err != nil {
		log.Println("Could not retrieve the last record ID:", err)
		return nil
	}

	if r == "" {
		log.Println("No previous record")
		return nil
	}

	currencies, err := redis.Strings(cacheConn.Do("HGETALL", r))
	if err != nil {
		log.Println("Could not retrieve dayrate from cache:", err)
		return nil
	}
	dayTime, _ := time.Parse("2006-01-02", strings.Split(r, ":")[0])
	dayRate := &eurobank.DayRate{Day: dayTime}

	for i := 0; i < len(currencies); i++ {

		rate, _ := strconv.ParseFloat(currencies[i+1], 32)
		currency := eurobank.Currency{Id: currencies[i], Rate: float32(rate)}
		dayRate.Rates = append(dayRate.Rates, currency)
		i++
	}
	return dayRate
}
