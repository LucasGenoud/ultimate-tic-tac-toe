//go:build js

package main

import (
	"encoding/json"
	"syscall/js"
)

// settingsKey is the local storage key of the settings, also read by wasm/index.html for its theme.
const settingsKey = "ultimate-tic-tac-toe.settings"

func readSettings() (s settings, ok bool) {
	defer func() {
		if recover() != nil { // local storage can be unavailable, for instance in private browsing
			ok = false
		}
	}()
	item := js.Global().Get("localStorage").Call("getItem", settingsKey)
	if item.Type() != js.TypeString {
		return s, false
	}
	return s, json.Unmarshal([]byte(item.String()), &s) == nil
}

func writeSettings(s settings) {
	defer func() { recover() }()
	data, _ := json.Marshal(s)
	js.Global().Get("localStorage").Call("setItem", settingsKey, string(data))
}

// prefersDark reports whether the browser asks for a dark theme.
func prefersDark() bool {
	match := js.Global().Call("matchMedia", "(prefers-color-scheme: dark)")
	return match.Truthy() && match.Get("matches").Bool()
}
