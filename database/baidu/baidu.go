package baidu

import "github.com/FloatTech/ZeroBot-Plugin/database"

func GetAkSk() (string, string) {
	return database.DefaultConfig.Baidu.Ak, database.DefaultConfig.Baidu.Sk
}
