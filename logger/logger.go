/*
 * Copyright (c) 2024 Fabien CHEVALIER
 * All rights reserved.
 */

package logger

import (
	"fmt"
	log "github.com/sirupsen/logrus"
)

type PlainFormatter struct {
}

// Format returns a formatted log entry message.
//
// Takes a log entry and returns the formatted message as a byte slice along with any error that occurs.
func (f *PlainFormatter) Format(entry *log.Entry) ([]byte, error) {
	return []byte(fmt.Sprintf("%s\n", entry.Message)), nil
}

// ToggleDebug toggles the debug mode in the application.
//
// cmd *cobra.Command, args []string
func ToggleDebug(debug bool) {
	if debug {
		log.Info("Debug logs enabled")
		log.SetLevel(log.DebugLevel)
		log.SetFormatter(&log.TextFormatter{})
	} else {
		plainFormatter := new(PlainFormatter)
		log.SetFormatter(plainFormatter)
	}
}
