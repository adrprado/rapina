package reports

import (
	"database/sql"
	"fmt"

	"github.com/adrprado/rapina/parsers"
)

type stItems struct {
	hash    uint32
	cdConta string
	dsConta string
}

//
// statementItems returns all statements codes and descriptions, e.g.:
// [1 Ativo Total, 1.01 Ativo Circulante, ...]
//
func statementItems(db *sql.DB, company string) (items []stItems, err error) {
	selectItems := fmt.Sprintf(`
	SELECT DISTINCT
		CD_CONTA, DS_CONTA
	FROM
		bpa
	WHERE
		DENOM_CIA LIKE "%s%%"
		AND ORDEM_EXERC LIKE "_LTIMO"

	UNION SELECT DISTINCT
		CD_CONTA, DS_CONTA
	FROM
		bpp
	WHERE
		DENOM_CIA LIKE "%s%%"
		AND ORDEM_EXERC LIKE "_LTIMO"

	UNION SELECT DISTINCT
		CD_CONTA, DS_CONTA
	FROM
		dre
	WHERE
		DENOM_CIA LIKE "%s%%"
		AND ORDEM_EXERC LIKE "_LTIMO"

	ORDER BY
		CD_CONTA, DS_CONTA
	;`, company, company, company)

	rows, err := db.Query(selectItems)
	if err != nil {
		panic(err)
	}

	var item stItems
	for rows.Next() {
		rows.Scan(&item.cdConta, &item.dsConta)
		item.hash = parsers.GetHash(item.cdConta + item.dsConta)
		items = append(items, item)
	}

	// genericPrint(rows)

	return
}

type statement struct {
	date     string
	denomCia string
	escala   string
	cdConta  string
	dsConta  string
	vlConta  float32
}

//
// financialReport
//
func financialReport(db *sql.DB, company string, year int) (statements map[uint32]float32, err error) {

	selectReport := fmt.Sprintf(`
	SELECT
		strftime('%%Y-%%m-%%d', DT_FIM_EXERC, 'unixepoch') AS DT,
		DENOM_CIA,
		ESCALA_MOEDA,
		CD_CONTA,
		DS_CONTA,
		VL_CONTA
	FROM
		bpa
	WHERE
		DENOM_CIA LIKE "%s%%"
		AND ORDEM_EXERC LIKE "_LTIMO"
		AND DT = "%d-12-31"

	UNION SELECT
		strftime('%%Y-%%m-%%d', DT_FIM_EXERC, 'unixepoch') AS DT,
		DENOM_CIA,
		ESCALA_MOEDA,
		CD_CONTA,
		DS_CONTA,
		VL_CONTA
	FROM
		bpp
	WHERE
		DENOM_CIA LIKE "%s%%"
		AND ORDEM_EXERC LIKE "_LTIMO"
		AND DT = "%d-12-31"

	UNION SELECT
		strftime('%%Y-%%m-%%d', DT_FIM_EXERC, 'unixepoch') AS DT,
		DENOM_CIA,
		ESCALA_DRE AS ESCALA_MOEDA,
		CD_CONTA,
		DS_CONTA,
		VL_CONTA
	FROM
		dre
	WHERE
		DENOM_CIA LIKE "%s%%"
		AND ORDEM_EXERC LIKE "_LTIMO"
		AND DT = "%d-12-31"

	ORDER BY
		DT, CD_CONTA
	;`, company, year, company, year, company, year)

	statements = make(map[uint32]float32)
	st := statement{}

	rows, err := db.Query(selectReport)
	for rows.Next() {
		rows.Scan(
			&st.date,
			&st.denomCia,
			&st.escala,
			&st.cdConta,
			&st.dsConta,
			&st.vlConta,
		)

		statements[parsers.GetHash(st.cdConta+st.dsConta)] = st.vlConta
	}

	return
}

//
// genericPrint prints the entire row
//
func genericPrint(rows *sql.Rows) (err error) {
	limit := 0
	cols, _ := rows.Columns()
	for rows.Next() {
		// Create a slice of interface{}'s to represent each column,
		// and a second slice to contain pointers to each item in the columns slice.
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		// Scan the result into the column pointers...
		if err := rows.Scan(columnPointers...); err != nil {
			return err
		}

		// Create our map, and retrieve the value for each column from the pointers slice,
		// storing it in the map with the name of the column as the key.
		// m := make(map[string]interface{})
		for i := range cols {
			val := columnPointers[i].(*interface{})
			// m[colName] = *val
			// fmt.Println(colName, *val)

			switch (*val).(type) {
			default:
				fmt.Print(*val, ";")
			case []uint8:
				y := *val
				var x = y.([]uint8)
				fmt.Print(string(x[:]), ";")
			}
		}
		fmt.Println()

		// Outputs: map[columnName:value columnName2:value2 columnName3:value3 ...]
		// fmt.Println(m)
		limit++
		if limit >= 4000 {
			break
		}
	}

	return
}
