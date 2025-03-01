package export

import (
	"discord-uptime-checker/constants"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"strconv"
)

var reg *prometheus.Registry

func ServeMetrics() (*prometheus.Registry, func() error, error) {
	if reg != nil {
		return reg, nil, nil
	}

	reg = prometheus.NewRegistry()

	g := gin.Default()
	g.GET("/metrics", gin.WrapH(promhttp.HandlerFor(
		reg,
		promhttp.HandlerOpts{
			Registry: reg,
		},
	)))

	portNum := 8080
	if constants.Port != "" {
		portNum, err := strconv.Atoi(constants.Port)
		if err != nil {
			return nil, nil, err
		}
		if portNum < 1 || portNum > 65535 {
			return nil, nil, fmt.Errorf("invalid port number: %d", portNum)
		}
	}

	return reg, func() error {
		address := ":" + strconv.Itoa(portNum)
		log.Printf("Exporting metrics on %v", address)
		return g.Run(address)
	}, nil
}
