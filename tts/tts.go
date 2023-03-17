package tts

import (
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/micmonay/keybd_event"
)

func Say(i string) error {
	// Init COM
	e := ole.CoInitialize(0)
	defer ole.CoUninitialize()

	// Process error
	if e != nil {
		return e
	}

	// Create object
	unknown, e := oleutil.CreateObject("SAPI.SpVoice")

	// Process error
	if e != nil {
		return e
	}

	// Get voice
	voice, e := unknown.QueryInterface(ole.IID_IDispatch)

	// Process error
	if e != nil {
		return e
	}

	// Speak
	_, e = oleutil.CallMethod(voice, "Speak", i)

	// Return
	return e
}

func VoiceInputToggle() {

	kb, err := keybd_event.NewKeyBonding()

	if err != nil {
		panic(err)
	}

	// Select keys to be pressed
	kb.HasSuper(true)
	kb.SetKeys(keybd_event.VK_H)

	// Press the selected keys
	err = kb.Launching()
	if err != nil {
		panic(err)
	}
}
