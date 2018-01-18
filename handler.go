package main

import {
    "fmt"
    "github.com/ratorx/chumenu-go/facebook"
}

struct EventHandler {
    commandPrefix string
}

func (e *EventHandler) HandleEvent(m []facebook.MessagingEvent) {
    for i := range m {
        r := m[i].Recipient.String()
        cfg.sendClient.SendMessage(r, "Default", facebook.Response, facebook.QuickReply{"a"})
    }
}
