package main

import (
	"github.com/FloatTech/ZeroBot-Plugin/database"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
)

type zbpcfg struct {
	Z        zero.Config        `json:"zero"`
	W        []*driver.WSClient `json:"ws"`
	Database database.Config    `json:"database"`
}

var config zbpcfg
