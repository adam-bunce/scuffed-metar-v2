package scrape

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"log/slog"
	"math"
	"time"
)

type MQTTReportLogTopicMessage struct {
	History []string `json:"history"`
}

func GetMesotechWeatherReport(site string) (*WeatherReport, error) {
	opts := MQTT.NewClientOptions().
		AddBroker("wss://mqtt.awos.live:8083/").
		SetTLSConfig(&tls.Config{InsecureSkipVerify: true}).
		SetClientID("$54a3d43a-4adc-418b-8d56-3785c96c3a16").
		SetUsername("AWA_Web_wVVdDr").
		SetPassword("Po&X58vexCkq;Wyp").
		SetConnectTimeout(3 * time.Second)

	mqttClient := MQTT.NewClient(opts)

	return ProcessMesotechMetarResponse(mqttClient, site)
}

func ProcessMesotechMetarResponse(client MQTT.Client, site string) (*WeatherReport, error) {
	res := WeatherReport{
		Airport: site,
	}

	token := client.Connect()
	token.Wait()
	if token.Error() != nil {
		return nil, token.Error()
	}

	topic := fmt.Sprintf("AWA/%s/Archives/ReportLog", site)
	subToken := client.Subscribe(topic, 0, func(client MQTT.Client, msg MQTT.Message) {
		var dest *MQTTReportLogTopicMessage

		err := json.NewDecoder(bytes.NewReader(msg.Payload())).Decode(&dest)
		if err != nil {
			slog.Error(err.Error())
			return
		}

		res.Metar = dest.History[:int(math.Min(float64(len(dest.History)), 5))]
	})

	subToken.Wait()
	if subToken.Error() != nil {
		return nil, subToken.Error()
	}

	client.Unsubscribe(topic)
	client.Disconnect(250)

	return &res, nil
}
