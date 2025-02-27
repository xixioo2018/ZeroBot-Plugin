package omikuji

import (
	sql "github.com/FloatTech/sqlite"
)

type kuji struct {
	ID   uint8  `db:"id"`
	Text string `db:"text"`
}

var db sql.Sqlite

// 返回一个解签
func getKujiByBango(id uint8) string {
	var s kuji
	err := db.Find("kuji", &s, "WHERE id = ?", id)
	if err != nil {
		return err.Error()
	}
	return s.Text
}
