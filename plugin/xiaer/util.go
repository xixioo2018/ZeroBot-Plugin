package xiaer

import (
	"fmt"
	"strconv"

	zero "github.com/wdvxdr1123/ZeroBot"
)

func GetFirstAt(ctx *zero.Ctx) (bool, int64) {
	for _, singleMessage := range ctx.Event.Message {
		if singleMessage.Type == "at" {
			fmt.Println(singleMessage.Data["qq"])
			parseInt, err := strconv.ParseInt(singleMessage.Data["qq"], 10, 64)
			if err != nil {
				continue
			}
			return true, parseInt
		}
	}
	return false, 0
}

func CardNameInGroup(ctx *zero.Ctx, userId int64) string {
	return ctx.GetGroupMemberInfo(ctx.Event.GroupID, userId, false).Get("nickname").String()
}
