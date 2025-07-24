package scrape

import (
	"fmt"
	"log/slog"
	"net/http"
	"scuffed-v2/internal/util"
	"strings"
)

const (
	CamecoRequestBody = `{
	   "request": {
	       "__type": "WebDataRequest:http://COM.AXYS.COMMON.WEB.CONTRACTS",
	       "Key": "METAR",
	       "DataSourceKey": "7e7dbc35-1d26-4b85-8f7e-077ad7bad794",
	       "Query": "SELECT TOP 100 PERCENT * FROM (SELECT TOP 1000 * FROM avWX_%s_METAR ORDER BY DataTimeStamp DESC) a WHERE DataTimeStamp >= DATEADD(DAY, -1, GETUTCDATE()) ORDER BY DataTimeStamp DESC"
	   }
	}`
)

type CamecoResponse struct {
	D struct {
		Type             string        `json:"__type"`
		AccessType       interface{}   `json:"AccessType"`
		Key              string        `json:"Key"`
		ModifyDateString string        `json:"ModifyDateString"`
		ModifyUser       interface{}   `json:"ModifyUser"`
		Properties       []interface{} `json:"Properties"`
		ColumnCount      int           `json:"ColumnCount"`
		Columns          []struct {
			Type       string `json:"__type"`
			ColumnName string `json:"ColumnName"`
			DataType   string `json:"DataType"`
			Ordinal    int    `json:"Ordinal"`
		} `json:"Columns"`
		RowCount int `json:"RowCount"`
		Rows     []struct {
			Type    string `json:"__type"`
			RowData string `json:"RowData"`
			RowID   int    `json:"RowID"`
		} `json:"Rows"`
	} `json:"d"`
}

// GetCamecoWeatherReport returns the metar readouts for the specified site
func GetCamecoWeatherReport(site string) (*WeatherReport, error) {
	var body CamecoResponse

	var camecoRequestBody = strings.NewReader(fmt.Sprintf(CamecoRequestBody, site))

	req, err := http.NewRequest("POST", "https://smartweb.axys-aps.com/svc/WebDataService.svc/WebData/GetWebDataResponse", camecoRequestBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Keep-Alive", "timeout=3")
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	err = util.RequestAndParse(req, &body)
	if err != nil {
		return nil, err
	}

	report, err := ProcessCamecoMetarResponse(body, site)
	if err != nil {
		return nil, err
	}

	return report, nil
}

func ProcessCamecoMetarResponse(mr CamecoResponse, site string) (*WeatherReport, error) {
	res := WeatherReport{
		Airport: site,
	}

	for i, row := range mr.D.Rows {
		if i == 5 {
			break
		}

		metar := strings.Split(row.RowData, ",")
		if len(metar) > 1 {
			res.Metar = append(res.Metar, metar[1])
		} else {
			slog.Error("unexpected cameco response row data items count",
				slog.Int("expected", 2),
				slog.Int("actual", len(mr.D.Rows)),
			)
			return &res, nil
		}
	}

	return &res, nil
}
