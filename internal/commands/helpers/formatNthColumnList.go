package helpers

import "github.com/jedib0t/go-pretty/v6/table"

func FormatNthColumnList[T interface{}](columnsAmount int, dataArr []T, headers table.Row, formatRowData func(data T, idx int) table.Row) string {
	rowCount := len(dataArr) / columnsAmount

	if len(dataArr)%columnsAmount != 0 {
		rowCount++
	}

	dataArrPtr := table.Row{}
	for _, v := range dataArr {
		dataArrPtr = append(dataArrPtr, &v)
	}

	for len(dataArrPtr)%columnsAmount != 0 {
		dataArrPtr = append(dataArrPtr, nil)
	}

	t := table.NewWriter()

	var duplicatedHeaders table.Row
	for i := 0; i < columnsAmount; i++ {
		duplicatedHeaders = append(duplicatedHeaders, headers...)
	}
	t.AppendHeader(duplicatedHeaders)

	for i := 0; i < rowCount; i++ {
		var rowData table.Row
		for j := 0; j < columnsAmount; j++ {
			if dataArrPtr[i+rowCount*j] == nil {
				for k := 0; k < len(headers); k++ {
					rowData = append(rowData, "")
				}
			} else {
				rowData = append(rowData, formatRowData(dataArr[i+rowCount*j], i+rowCount*j)...)
			}
		}
		t.AppendRow(rowData)
	}

	return t.Render()
}
