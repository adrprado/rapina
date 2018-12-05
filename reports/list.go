package reports

import (
	"database/sql"
	"fmt"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

//
// ListCompanies shows all available companies
//
func ListCompanies(db *sql.DB) (err error) {
	com, err := companies(db)

	if err != nil {
		fmt.Println("[x] Falha:", err)
		return
	}

	// Sort accents correctly
	cl := collate.New(language.BrazilianPortuguese, collate.Loose)
	cl.SortStrings(com)

	for _, c := range com {
		fmt.Println(c)
	}

	return
}
