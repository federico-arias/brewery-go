package main

import (
	"io/ioutil"
	"net/http"
)

func main() {
	for {
		res, err := http.Get("https://api.ipify.org")
		if err != nil {
			fmt.Println("Couldn't get IP: ", err.Error())
		}
		ip, err := ioutil.ReadAll(res.Body)
		defer res.Body.Close()
		if err != nil {
			fmt.Println("failure to read response body: ", err.Error())
		}
		_, err = m.Set("ip", string(ip), 0, 0, 0)
		if err != nil {
			fmt.Println("failure to store IP in memcached: ", err.Error())
		}
		time.Sleep(3 * time.Hour)
	}
}
