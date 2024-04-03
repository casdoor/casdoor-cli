package utils

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"os"
)

func PrintTable(inputMap map[string]interface{}) {
	table := tablewriter.NewWriter(os.Stdout)

	var headers []string
	var values []string

	for key, value := range inputMap {
		headers = append(headers, key)
		values = append(values, fmt.Sprintf("%v", value))
	}

	if len(headers) > 0 {
		table.SetHeader(headers)

		headerColors := make([]tablewriter.Colors, len(headers))
		for i := 0; i < len(headers); i++ {
			headerColors[i] = tablewriter.Colors{tablewriter.Bold}
		}
		table.SetHeaderColor(headerColors...)

		table.SetAutoMergeCells(false)
		table.SetRowLine(false)
		table.SetAlignment(tablewriter.ALIGN_LEFT)

		table.Append(values)
	}

	table.Render()
}

func PrintTables(items []map[string]interface{}) {
	table := tablewriter.NewWriter(os.Stdout)

	var headers []string

	if len(items) > 0 {
		for key := range items[0] {
			headers = append(headers, key)
		}

		table.SetHeader(headers)

		headerColors := make([]tablewriter.Colors, len(headers))
		for i := 0; i < len(headers); i++ {
			headerColors[i] = tablewriter.Colors{tablewriter.Bold}
		}
		table.SetHeaderColor(headerColors...)

		table.SetAutoMergeCells(false)
		table.SetRowLine(false)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
	}

	for _, item := range items {
		var values []string
		for _, header := range headers {
			values = append(values, fmt.Sprintf("%v", item[header]))
		}
		table.Append(values)
	}

	table.Render()
}
