package xlsximporter

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/jinzhu/gorm"
	"github.com/tealeg/xlsx"
	"satori/pawsimporter/area"
	"satori/pawsimporter/control"
	"satori/pawsimporter/objective"
	. "satori/pawsimporter/paws"
	"satori/pawsimporter/risk"
	"satori/pawsimporter/test"
	"strings"
)

type Importer struct {
	DB               *gorm.DB
	DataFile         string
	xlsxFile         *xlsx.File
	areaColumns      []Column
	objectiveColumns []Column
	riskColumns      []Column
	controlColumns   []Column
	testColumns      []Column
}

func (i *Importer) Begin() {
	fmt.Println("Begin....")
	if readble, err := i.isExcelFileReadable(); !readble {
		color.Red("Sorry, unable to read given file: %s, error: %s", i.DataFile, err)
		return
	}
	err := i.readXlsx()
	if err != nil {
		color.Red("Sorry, error occurred %s", err)
		return
	}
	i.startImporting()
}

func (i *Importer) readXlsx() error {
	xlFile, err := xlsx.OpenFile(i.DataFile)
	if err != nil {
		color.Red("Unknown error occurred while reading data file: %s", err)
		return err
	}
	i.xlsxFile = xlFile
	// supports only one excel sheet
	sheetIndex := 1
	for _, sheet := range xlFile.Sheets {
		if sheetIndex != 1 {
			continue
		}

		color.Blue("Setting index for getting area fields: %d", 1)
		nextIndex, err := i.getColumns(sheet.Rows, "area", 1)
		if err != nil {
			color.Red("Unable to read area columns: %s", err)
			return err
		}

		color.Blue("Setting index for getting objective fields: %d", nextIndex)
		nextIndex, err = i.getColumns(sheet.Rows, "objective", nextIndex)
		if err != nil {
			color.Red("Unable to read objective columns: %s", err)
			return err
		}

		color.Blue("Setting index for getting risk fields: %d", nextIndex)
		nextIndex, err = i.getColumns(sheet.Rows, "risk", nextIndex)
		if err != nil {
			color.Red("Unable to read risk columns: %s", err)
			return err
		}

		color.Blue("Setting index for getting control fields: %d", nextIndex)
		nextIndex, err = i.getColumns(sheet.Rows, "control", nextIndex)
		if err != nil {
			color.Red("Unable to read control columns: %s", err)
			return err
		}

		color.Blue("Setting index for getting test fields: %d", nextIndex)
		nextIndex, err = i.getColumns(sheet.Rows, "test", nextIndex)
		if err != nil {
			color.Red("Unable to read test columns: %s", err)
			return err
		}

		sheetIndex += 1
	}
	return nil

}

func (i *Importer) startImporting() {
	areaService := &area.Area{}
	areaService.DB = i.DB
	areaService.Columns = i.areaColumns

	objectiveService := &objective.Objective{}
	objectiveService.DB = i.DB
	objectiveService.Columns = i.objectiveColumns

	riskService := &risk.Risk{}
	riskService.DB = i.DB
	riskService.Columns = i.riskColumns

	controlService := &control.Control{}
	controlService.DB = i.DB
	controlService.Columns = i.controlColumns

	testService := &test.Test{}
	testService.DB = i.DB
	testService.Columns = i.testColumns

	rowLength := len(i.xlsxFile.Sheets[0].Rows)
	for rowIndex, row := range i.xlsxFile.Sheets[0].Rows[5:rowLength] {
		areaId := areaService.Update(i.getAreaValues(row), rowIndex+6)
		if areaId == "" {
			continue
		}

		objId := objectiveService.Update(i.getObjectiveValues(row), rowIndex+6)
		if objId != "" {
			riskId := riskService.Update(i.getRiskValues(row), rowIndex+6, objId)
			if riskId != "" {
				ctrId := controlService.Update(i.getControlValues(row), rowIndex+6, riskId)
				if ctrId != "" {
					testService.Update(i.getTestValues(row), rowIndex+6, areaId, riskId, ctrId)
				}
			} else {
				color.Blue("Unabel to find Risk in row number %d", rowIndex+6)
			}
		} else {
			color.Blue("Unabel to find Objective in row number %d", rowIndex+6)
		}

	}
}

func (i *Importer) getAreaValues(row *xlsx.Row) []Data {
	var columnsData []Data

	for cellIndex, cell := range row.Cells {
		for _, areaColumn := range i.areaColumns {
			if areaColumn.Index == cellIndex {
				columnsData = append(columnsData, Data{cell.String(), cellIndex})
			}
		}
	}
	return columnsData
}

func (i *Importer) getObjectiveValues(row *xlsx.Row) []Data {
	var columnsData []Data

	for cellIndex, cell := range row.Cells {
		for _, objectiveColumn := range i.objectiveColumns {
			if objectiveColumn.Index == cellIndex {

				columnsData = append(columnsData, Data{cell.String(), cellIndex})
			}
		}
	}
	return columnsData
}

func (i *Importer) getRiskValues(row *xlsx.Row) []Data {
	var columnsData []Data

	for cellIndex, cell := range row.Cells {
		for _, riskColumn := range i.riskColumns {
			if riskColumn.Index == cellIndex {
				columnsData = append(columnsData, Data{cell.String(), cellIndex})
			}
		}
	}
	return columnsData
}

func (i *Importer) getControlValues(row *xlsx.Row) []Data {
	var columnsData []Data

	for cellIndex, cell := range row.Cells {
		for _, controlColumn := range i.controlColumns {
			if controlColumn.Index == cellIndex {
				columnsData = append(columnsData, Data{cell.String(), cellIndex})
			}
		}
	}

	return columnsData
}

func (i *Importer) getTestValues(row *xlsx.Row) []Data {
	var columnsData []Data

	for cellIndex, cell := range row.Cells {
		for _, testColumn := range i.testColumns {
			if testColumn.Index == cellIndex {
				columnsData = append(columnsData, Data{cell.String(), cellIndex})
			}
		}
	}
	return columnsData
}
func (i *Importer) importObjective() {

}

func (i *Importer) importRisk(objectiveId string) {

}

func (i *Importer) importControl(riskId string) {

}

func (i *Importer) importTest(parentId, parentType string) {

}

func (i *Importer) getColumns(rows []*xlsx.Row, tableName string, cellIndex int) (int, error) {
	if !i.CellReadble(4, 0, rows) {
		return 0, errors.New("looks like you did not follow the expected format.")
	}

	rowIndex := 0
	color.Blue("Getting columns for %s", tableName)
	for _, row := range rows[4:5] {
		for _, cell := range row.Cells[cellIndex:len(row.Cells)] {
			column := Column{}
			tableInfo := strings.ToLower(rows[0].Cells[cellIndex].String())
			tableInfo = strings.TrimSpace(tableInfo)

			importInfo := strings.ToLower(rows[1].Cells[cellIndex].String())
			importInfo = strings.TrimSpace(importInfo)

			if importInfo == "y" || importInfo == "yes" {
				column.Import = true
			}

			keyInfo := strings.ToLower(rows[2].Cells[cellIndex].String())
			keyInfo = strings.TrimSpace(keyInfo)
			if keyInfo == "y" || keyInfo == "yes" {
				column.Key = true
			}

			column.Index = cellIndex
			column.Name = cell.String()
			if i.CellReadble(3, cellIndex, rows) {
				column.Alias = rows[3].Cells[cellIndex].String()
			}

			if tableInfo != tableName && tableInfo != "" {
				return cellIndex, nil
			}

			if strings.TrimSpace(column.Name) != "" {
				color.Green("  %s", cell.String())
				i.addColumns(column, tableName)
			}
			cellIndex += 1
		}
		rowIndex += 1
	}

	return 0, nil
}

func (i *Importer) addColumns(column Column, tableName string) {
	if tableName == "area" {
		i.areaColumns = append(i.areaColumns, column)
		return
	}

	if tableName == "objective" {
		i.objectiveColumns = append(i.objectiveColumns, column)
		return
	}

	if tableName == "risk" {
		i.riskColumns = append(i.riskColumns, column)
		return
	}

	if tableName == "control" {
		i.controlColumns = append(i.controlColumns, column)
		return
	}

	if tableName == "test" {
		i.testColumns = append(i.testColumns, column)
		return
	}

}

func (i *Importer) CellReadble(rowIndex, cellIndex int, rows []*xlsx.Row) bool {
	readable := false
	if rowIndex < len(rows) && cellIndex < len(rows[rowIndex].Cells) {
		return true
	}
	return readable
}

func (i *Importer) isExcelFileReadable() (bool, error) {
	_, err := xlsx.OpenFile(i.DataFile)
	if err != nil {
		return false, err
	}

	return true, nil
}
