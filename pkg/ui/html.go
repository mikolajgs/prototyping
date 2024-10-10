package ui

import (
	"embed"
	"fmt"
	"html"
)

//go:embed html/*
var htmlDir embed.FS

const MsgSuccess = 1
const MsgFailure = 2

func (c *Controller) getMsgHTML(msgType int, msg string) string {
	if msgType == 0 {
		return ""
	}
	return fmt.Sprintf("<div>%d: %s</div>", msgType, html.EscapeString(msg))
}
