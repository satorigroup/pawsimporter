package area

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/jinzhu/gorm"
	"github.com/twinj/uuid"
	. "satori/pawsimporter/paws"
	"strings"
)

var TABLE string = GetPAWSInfo("AREA_TABLE", "LibAreas")

var ID string = GetPAWSInfo("AREA_ID_FIELD", "laa_ID")

type Area struct {
	DB      *gorm.DB
	Columns []Column
}

type SqlResult struct {
	LaaArearef string `gorm:"column:laa_AreaRef"`
}

func New(db *gorm.DB, columns []Column) *Area {
	areaSvc := &Area{DB: db, Columns: columns}
	areaSvc.setColumnType()
	return areaSvc
}

func (a *Area) Update(columnsData []Data, rowIndex int) string {
	exist, areaId := a.exist(columnsData, rowIndex)
	insertSQL := ""
	if !exist {
		insertSQL, areaId = a.getInsertString(columnsData)
		a.DB.Exec(insertSQL)
		color.Blue("Inserting a new Area for row number %d", rowIndex)
	}
	updateString, err := a.getUpdateString(columnsData)
	if err != nil {
		color.Red("No fields available for area in row number %d", rowIndex)
	} else {
		whereString, _ := a.getWhereString(columnsData)
		updateResult := a.DB.Exec(fmt.Sprintf("%s WHERE %s", updateString, whereString))

		if updateResult.Error != nil {
			color.Red("error occurred at row no %d while updating area : %s", rowIndex, updateResult.Error)
			return ""
		}
		color.Green("Area is updated at row number %d", rowIndex)
	}

	return areaId
}

func (a *Area) getInsertString(columnsData []Data) (string, string) {
	rawQuery := fmt.Sprintf("INSERT INTO %s ", TABLE)

	fieldString := ""
	valueString := ""

	for _, column := range a.Columns {
		if !column.Import {
			continue
		}
		fieldString += fmt.Sprintf("%s, ", column.Name)
		for _, columnData := range columnsData {
			if column.Index != columnData.Index {
				continue
			}
			valueString += fmt.Sprintf("'%s', ", a.safeSQLValue(columnData.Value))
		}
	}
	fieldString = strings.TrimRight(fieldString, ", ")
	valueString = strings.TrimRight(valueString, ", ")
	guid := uuid.NewV4().String()
	rawQuery += fmt.Sprintf("(%s, %s) VALUES ('%s', %s)", ID, fieldString, guid, valueString)

	return rawQuery, guid
}
func (a *Area) getUpdateString(columnsData []Data) (string, error) {
	rawQuery := fmt.Sprintf("Update %s SET ", TABLE)
	fieldsExist := false
	for _, column := range a.Columns {
		if column.Key || !column.Import {
			continue
		}
		for _, columnData := range columnsData {
			if column.Index != columnData.Index {
				continue
			}
			if !fieldsExist {
				fieldsExist = true
			}
			rawQuery += fmt.Sprintf(" %s = '%s', ", a.safeSQLColumn(column.Name), a.safeSQLValue(columnData.Value))
		}
	}
	if !fieldsExist {
		return "", errors.New("no fields exist")
	}

	rawQuery = strings.TrimRight(rawQuery, ", ")
	return rawQuery, nil
}

func (a *Area) getWhereString(columnsData []Data) (string, error) {
	refs := a.getRefValue(columnsData)
	rawQuery := ""

	if len(refs) < 1 {
		return "", errors.New("there is no reference fields")
	}

	i := 1
	for key, value := range refs {
		if i == len(refs) {
			rawQuery += fmt.Sprintf(" %s = '%s'", a.safeSQLColumn(key), a.safeSQLValue(value))

		} else {
			rawQuery += fmt.Sprintf(" %s = '%s' AND ", a.safeSQLColumn(key), a.safeSQLValue(value))

		}
		i++
	}

	return rawQuery, nil
}

func (a *Area) safeSQLValue(input string) string {
	output := strings.Replace(input, "'", `''`, -1)
	return output
}

func (a *Area) safeSQLColumn(input string) string {
	output := strings.Replace(input, "'", `''`, -1)
	return output
}

func (a *Area) setColumnType() {
	type SqlColumnInfo struct {
		ColumnName string `gorm:"column:COLUMN_NAME"`
		DataType   string `gorm:"column:DATA_TYPE"`
	}
	var sqlInfo []SqlColumnInfo
	a.DB.Table("INFORMATION_SCHEMA.COLUMNS").Select("COLUMN_NAME, DATA_TYPE").Where("TABLE_NAME = ?", TABLE).Scan(&sqlInfo)
	for _, cInfo := range sqlInfo {
		for cIndex, column := range a.Columns {
			if strings.ToLower(column.Name) == strings.ToLower(cInfo.ColumnName) {
				a.Columns[cIndex].SqlType = cInfo.DataType

			}
		}
	}
}

func (a *Area) getRefValue(columnsData []Data) map[string]string {
	ref := make(map[string]string)
	for _, column := range a.Columns {
		if column.Key {
			for _, columnData := range columnsData {
				if columnData.Index == column.Index {
					ref[column.Name] = columnData.Value
				}
			}
		}
	}
	return ref
}

func (a *Area) exist(columnsData []Data, rowIndex int) (bool, string) {
	whereString, err := a.getWhereString(columnsData)

	if err != nil {
		color.Red("No reference fields for area in row number %d", rowIndex)
		return false, ""
	}

	id := ""

	row := a.DB.Table(TABLE).Select(fmt.Sprintf("convert(nvarchar(36), %s) as id", ID)).Where(whereString).Row()
	row.Scan(&id)
	if id != "" {
		return true, id
	}

	return false, ""
}
