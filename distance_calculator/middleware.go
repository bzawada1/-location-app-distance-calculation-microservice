package main

import (
	"time"

	"github.com/bzawada1/location-app-obu-service/types"
	"github.com/sirupsen/logrus"
)

type LogMiddleware struct {
	next CalculatorServicer
}

func NewLogMiddleware(next CalculatorServicer) CalculatorServicer {
	return &LogMiddleware{
		next: next,
	}
}

func (l *LogMiddleware) CalculateDistance(data types.OBUData) (dist float64, err error) {
	defer func(start time.Time) {
		logrus.WithFields(logrus.Fields{
			"obuID": data.OBUID,
			"err":   err,
			"dist":  dist,
			"took":  time.Since(start),
		}).Info("calculating the distance")
	}(time.Now())
	dist, err = l.next.CalculateDistance(data)
	return
}
