package xlsximporter

import (
	"errors"
	"github.com/fatih/color"
	"github.com/jinzhu/gorm"
	"github.com/tealeg/xlsx"
	"strings"
)

const (
	TABLE_INFO_ROW  = iota // 0
	IMPORT_INFO_ROW        // 1
	KEY_INFO_ROW           // 2
	ALIAS_INFO_ROW         // 3
	FIELD_INFO_ROW         // 4
	START_DATA_ROW         // 5
)

type Column struct {
	Name, Alias string
	Key, Import bool
	Index       int
}

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
	if readble, err := i.isExcelFileReadable(); !readble {
		color.Red("Sorry, unable to read given file: %s, error: %s", i.DataFile, err)
		return
	}
	err := i.readXlsx()
	if err != nil {
		color.Red("Sorry, error occurred %s", err)
		return
	}
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
		i.riskColumns = append(i.riskColumns, column)
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
