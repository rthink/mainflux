package Utils

import (
	influxdata "github.com/influxdata/influxdb/client/v2"
	"log"
)

func ConnInflux() influxdata.Client {
	cli, err := influxdata.NewHTTPClient(influxdata.HTTPConfig{
		Addr:     "http://118.31.19.149:8086", /*"http://127.0.0.1:8086"*/
		Username: "mainflux",                  /*"admin"*/
		Password: "mainflux",                  /*""*/
	})
	if err != nil {
		log.Fatal(err)
	}
	return cli
}

// query
func QueryDB(cli influxdata.Client, cmd string) (res []influxdata.Result, err error) {
	q := influxdata.Query{
		Command:  cmd,
		Database: "mainflux",
	}
	if response, err := cli.Query(q); err == nil {
		if response.Error() != nil {
			return res, response.Error()
		}
		res = response.Results
	} else {
		return res, err
	}
	return res, nil
}