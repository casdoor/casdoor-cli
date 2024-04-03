package utils

import (
	"fmt"
	log "github.com/sirupsen/logrus"
)

func Colorize(colorFunc func(format string, a ...interface{}) string, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	coloredMessage := colorFunc(message)
	log.Info(coloredMessage)
}
