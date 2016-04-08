package control

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/jinzhu/gorm"
	. "satori/pawsimporter/paws"
	"strings"
)

const (
	TABLE               = "LibControlsV3"
	RISKCONTROL_TABLE   = "LibRiskControlLinks"
	ID                  = "lct_Id"
	RISKCONTROL_RISK_ID = "lrcl_RiskID"
	RISKCONTROL_CTRL_ID = "lrcl_ControlID"
)

type Control struct {
	DB      *gorm.DB
	Columns []Column
}

type ControlId struct {
	Id string `gorm:"column:id"`
}

func New(db *gorm.DB, columns []Column) *Control {
	ctrSvc := &Control{DB: db, Columns: columns}
	ctrSvc.addControlKey()

	return ctrSvc
}

func (a *Control) GetControlIds(rskId string) []ControlId {
	var ctrIds []ControlId
	idColumn := fmt.Sprintf("%s = ?", RISKCONTROL_RISK_ID)
	a.DB.Table(RISKCONTROL_TABLE).Select(fmt.Sprintf("convert(nvarchar(36), %s) as id", RISKCONTROL_CTRL_ID)).Where(idColumn, rskId).Scan(&ctrIds)
	return ctrIds
}

func (a *Control) addControlKey() {
	ctrIdColumn := Column{Import: false, Key: true, Index: -1, Name: ID}
	a.Columns = append(a.Columns, ctrIdColumn)
}

func (a *Control) setControlIdvalue(ctrId string, columnsData *[]Data) {
	ctrDataColumn := Data{Index: -1, Value: ctrId}
	*columnsData = append(*columnsData, ctrDataColumn)
}

func (a *Control) Update(columnsData []Data, rowIndex int, rskId string) string {
	ctrIds := a.GetControlIds(rskId)

	for _, ctrId := range ctrIds {
		a.setControlIdvalue(ctrId.Id, &columnsData)
		exist, id := a.exist(columnsData, rowIndex)
		if !exist {
			continue
		}

		updateString, err := a.getUpdateString(columnsData)
		if err != nil {
			color.Red("No fields available for control in row number %d", rowIndex)
			return ""
		}

		whereString, err := a.getWhereString(columnsData)
		a.DB.Exec(fmt.Sprintf("%s WHERE %s", updateString, whereString))
		color.Green("Control is updated at row number %d", rowIndex)
		return id
	}

	color.Blue("Control does not exist which is specified in row number %d", rowIndex)
	return ""
}

func (a *Control) getUpdateString(columnsData []Data) (string, error) {
	rawQuery := fmt.Sprintf("Update %s SET ", TABLE)
	fieldsExist := false
	for _, column := range a.Columns {
		if column.Key {
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

func (a *Control) getWhereString(columnsData []Data) (string, error) {
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

func (a *Control) safeSQLValue(input string) string {
	output := strings.Replace(input, "'", `''`, -1)
	return output
}

func (a *Control) safeSQLColumn(input string) string {
	output := strings.Replace(input, "'", `''`, -1)
	return output
}

func (a *Control) getRefValue(columnsData []Data) map[string]string {
	ref := make(map[string]string)
	for _, column := range a.Columns {
		if column.Key {
			for _, columnData := range columnsData {
				if columnData.Index == column.Index && column.Name != "" {
					ref[column.Name] = columnData.Value
				}
			}
		}
	}
	return ref
}

func (a *Control) exist(columnsData []Data, rowIndex int) (bool, string) {
	whereString, err := a.getWhereString(columnsData)

	if err != nil {
		color.Red("No reference fields for control in row number %d", rowIndex)
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
