package renderer

import (
	"unicode/utf16"
	"unsafe"

	"github.com/LamkasDev/cimgui-go-vulkan/imgui"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/ime"
)

func (overlay *ImguiOverlay) DrawIme() {
	if !ime.GlobalImeDevice.IsDialogOpen || ime.GlobalImeDevice.DialogStatus != ime.ImeDialogStatusRunning {
		return
	}

	// For simplicity, we center it. We could use Param.PosX/PosY if needed.
	imgui.SetNextWindowPos(imgui.Vec2{X: 800/2 - 200, Y: 600/2 - 50})
	if imgui.BeginV("IME Input", nil, imgui.WindowFlagsAlwaysAutoResize|imgui.WindowFlagsNoSavedSettings|imgui.WindowFlagsNoTitleBar) {
		imgui.Text("PS4 On-Screen Keyboard")
		if imgui.InputTextWithHint("##ime_input", "Type here...", &ime.GlobalImeDevice.InputText, imgui.InputTextFlagsEnterReturnsTrue, nil) {
			// Encode as UTF-16 and null-terminate.
			newText := ime.GlobalImeDevice.InputText
			encoded := utf16.Encode([]rune(newText))
			encoded = append(encoded, 0)

			if ime.GlobalImeDevice.IsKeyboardOpen {
				for _, char := range encoded {
					if char == 0 {
						continue // Skip null terminator.
					}

					// OrbisImeKeycode layout:
					// offset 0: u16 keycode
					// offset 2: char16_t character
					// offset 4: u32 status
					// offset 12: int32 user_id
					// We'll set the character and let keycode be 0 for now.
					event := ime.ImeEvent{Id: ime.ImeEventIdKeyboardKeycodeDown}
					event.Param.Data[2] = byte(char)
					event.Param.Data[3] = byte(char >> 8)
					// status = CHARACTER_VALID | FROM_OSK
					event.Param.Data[4] = 0x0A

					userId := ime.GlobalImeDevice.KeyboardUserId
					event.Param.Data[12] = byte(userId)
					event.Param.Data[13] = byte(userId >> 8)
					event.Param.Data[14] = byte(userId >> 16)
					event.Param.Data[15] = byte(userId >> 24)
					ime.GlobalImeDevice.SendEvent(event)

					eventUp := event
					eventUp.Id = ime.ImeEventIdKeyboardKeycodeUp
					ime.GlobalImeDevice.SendEvent(eventUp)
				}

				// Send Enter key
				event := ime.ImeEvent{Id: ime.ImeEventIdKeyboardKeycodeDown}
				event.Param.Data[0] = 0x0D // VK_RETURN or similar
				event.Param.Data[2] = 0x0D
				// status = KEYCODE_VALID | CHARACTER_VALID | FROM_OSK
				event.Param.Data[4] = 0x0B

				userId := ime.GlobalImeDevice.KeyboardUserId
				event.Param.Data[12] = byte(userId)
				event.Param.Data[13] = byte(userId >> 8)
				event.Param.Data[14] = byte(userId >> 16)
				event.Param.Data[15] = byte(userId >> 24)
				ime.GlobalImeDevice.SendEvent(event)

				eventUp := event
				eventUp.Id = ime.ImeEventIdKeyboardKeycodeUp
				ime.GlobalImeDevice.SendEvent(eventUp)
			} else if ime.GlobalImeDevice.IsDialogOpen {
				// Write to guest memory
				if ime.GlobalImeDevice.DialogParam.InputTextBuffer != nil {
					maxLen := int(ime.GlobalImeDevice.DialogParam.MaxTextLength)
					destSlice := unsafe.Slice(ime.GlobalImeDevice.DialogParam.InputTextBuffer, maxLen)

					for i := 0; i < maxLen; i++ {
						if i < len(encoded) {
							destSlice[i] = encoded[i]
						} else {
							destSlice[i] = 0
						}
					}
					if maxLen > 0 {
						destSlice[maxLen-1] = 0
					}
				}

				ime.GlobalImeDevice.Mutex.Lock()
				ime.GlobalImeDevice.DialogStatus = ime.ImeDialogStatusFinished
				ime.GlobalImeDevice.DialogResult = ime.ImeDialogResult{EndStatus: ime.ImeDialogEndStatusOk}
				ime.GlobalImeDevice.Mutex.Unlock()
			} else {
				// Write to guest memory
				if ime.GlobalImeDevice.Param.InputTextBuffer != nil {
					maxLen := int(ime.GlobalImeDevice.Param.MaxTextLength)
					destSlice := unsafe.Slice(ime.GlobalImeDevice.Param.InputTextBuffer, maxLen)

					for i := 0; i < maxLen; i++ {
						if i < len(encoded) {
							destSlice[i] = encoded[i]
						} else {
							destSlice[i] = 0
						}
					}
					if maxLen > 0 {
						destSlice[maxLen-1] = 0
					}
				}

				// Construct the event
				event := ime.ImeEvent{Id: ime.ImeEventIdPressEnter}
				ime.GlobalImeDevice.SendEvent(event)
			}
		}
		imgui.End()
	}
}
