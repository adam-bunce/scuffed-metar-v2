package scrape

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
//func GetCamecoWeatherReport(site string) (map[string]*WeatherReport, error) {
//
//}
