package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/tidwall/gjson"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv" // autoregisters driver
)

type HueBridge struct {
	IP       string
	Username string
}

type Light struct {
	ID   string
	Name string
}

type MIDICalibration struct {
	LeftKey  uint8
	RightKey uint8
}

func main() {
	fmt.Println("🎹 HueMIDI - Control your Hue lights with MIDI!")
	fmt.Println("==========================================")

	// Discover Hue bridge
	bridge, err := discoverHueBridge()
	if err != nil {
		log.Fatal("Failed to discover Hue bridge:", err)
	}

	fmt.Printf("✅ Found Hue bridge at: %s\n", bridge.IP)

	// Authenticate with bridge
	err = authenticateWithBridge(bridge)
	if err != nil {
		log.Fatal("Failed to authenticate with bridge:", err)
	}

	// Get available lights
	lights, err := getLights(bridge)
	if err != nil {
		log.Fatal("Failed to get lights:", err)
	}

	// Let user select a light
	selectedLight, err := selectLight(lights)
	if err != nil {
		log.Fatal("Failed to select light:", err)
	}

	fmt.Printf("✅ Selected light: %s\n", selectedLight.Name)

	// Calibrate MIDI keyboard
	calibration, err := calibrateMIDIKeyboard()
	if err != nil {
		log.Fatal("Failed to calibrate MIDI keyboard:", err)
	}

	fmt.Printf("✅ MIDI keyboard calibrated: Left key %d, Right key %d\n", calibration.LeftKey, calibration.RightKey)

	// Start MIDI listener
	err = startMIDIListener(bridge, selectedLight, calibration)
	if err != nil {
		log.Fatal("Failed to start MIDI listener:", err)
	}
}

func discoverHueBridge() (*HueBridge, error) {
	fmt.Println("🔍 Discovering Hue bridge...")

	// Try the official discovery endpoint
	resp, err := http.Get("https://discovery.meethue.com/")
	if err != nil {
		return nil, fmt.Errorf("failed to discover bridge: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read discovery response: %v", err)
	}

	// Parse JSON response
	bridges := gjson.GetBytes(body, "#.internalipaddress")
	if !bridges.Exists() || len(bridges.Array()) == 0 {
		return nil, fmt.Errorf("no Hue bridges found")
	}

	// Use the first bridge found
	bridgeIP := bridges.Array()[0].String()

	return &HueBridge{IP: bridgeIP}, nil
}

func authenticateWithBridge(bridge *HueBridge) error {
	fmt.Println("🔐 Authenticating with Hue bridge...")

	// Check if we already have a username stored
	if username := os.Getenv("HUE_USERNAME"); username != "" {
		bridge.Username = username
		return nil
	}

	fmt.Println("Please press the link button on your Hue bridge, then press Enter...")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadLine()

	// Request username
	requestBody := `{"devicetype":"huemidi#cli"}`
	url := fmt.Sprintf("http://%s/api", bridge.IP)

	resp, err := http.Post(url, "application/json", strings.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("failed to authenticate: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read auth response: %v", err)
	}

	// Parse response
	username := gjson.GetBytes(body, "0.success.username")
	if !username.Exists() {
		errorMsg := gjson.GetBytes(body, "0.error.description")
		return fmt.Errorf("authentication failed: %s", errorMsg.String())
	}

	bridge.Username = username.String()
	fmt.Printf("✅ Authenticated! Username: %s\n", bridge.Username)
	fmt.Printf("💡 Set HUE_USERNAME=%s to skip this step next time\n", bridge.Username)

	return nil
}

func getLights(bridge *HueBridge) ([]Light, error) {
	fmt.Println("💡 Getting available lights...")

	url := fmt.Sprintf("http://%s/api/%s/lights", bridge.IP, bridge.Username)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get lights: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read lights response: %v", err)
	}

	var lights []Light
	result := gjson.ParseBytes(body)

	result.ForEach(func(key, value gjson.Result) bool {
		name := value.Get("name").String()
		if name != "" {
			lights = append(lights, Light{
				ID:   key.String(),
				Name: name,
			})
		}
		return true
	})

	if len(lights) == 0 {
		return nil, fmt.Errorf("no lights found")
	}

	return lights, nil
}

func selectLight(lights []Light) (*Light, error) {
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "▶ {{ .Name | cyan }}",
		Inactive: "  {{ .Name | white }}",
		Selected: "✅ {{ .Name | green }}",
	}

	prompt := promptui.Select{
		Label:     "Select a light to control",
		Items:     lights,
		Templates: templates,
	}

	i, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}

	return &lights[i], nil
}

func calibrateMIDIKeyboard() (*MIDICalibration, error) {
	fmt.Println("🎹 Calibrating MIDI keyboard...")

	defer midi.CloseDriver()

	ins := midi.GetInPorts()
	if len(ins) == 0 {
		return nil, fmt.Errorf("no MIDI input devices found")
	}

	fmt.Printf("Found MIDI devices:\n")
	for i, in := range ins {
		fmt.Printf("  %d: %s\n", i, in.String())
	}

	// Use the first available MIDI device
	in := ins[0]
	fmt.Printf("Using MIDI device: %s\n", in.String())

	stop, err := midi.ListenTo(in, func(msg midi.Message, timestampms int32) {
		// We'll handle this in the calibration process
	}, midi.UseSysEx())
	if err != nil {
		return nil, fmt.Errorf("failed to listen to MIDI device: %v", err)
	}
	defer stop()

	calibration := &MIDICalibration{}

	// Calibrate left key
	fmt.Println("Press the LEFT-MOST key on your MIDI keyboard...")
	leftKey, err := waitForMIDIKey(in)
	if err != nil {
		return nil, fmt.Errorf("failed to get left key: %v", err)
	}
	calibration.LeftKey = leftKey
	fmt.Printf("✅ Left key: %d\n", leftKey)

	// Calibrate right key
	fmt.Println("Press the RIGHT-MOST key on your MIDI keyboard...")
	rightKey, err := waitForMIDIKey(in)
	if err != nil {
		return nil, fmt.Errorf("failed to get right key: %v", err)
	}
	calibration.RightKey = rightKey
	fmt.Printf("✅ Right key: %d\n", rightKey)

	if calibration.LeftKey >= calibration.RightKey {
		return nil, fmt.Errorf("left key (%d) should be less than right key (%d)", calibration.LeftKey, calibration.RightKey)
	}

	return calibration, nil
}

func waitForMIDIKey(in drivers.In) (uint8, error) {
	keyChan := make(chan uint8, 1)

	stop, err := midi.ListenTo(in, func(msg midi.Message, timestampms int32) {
		var channel, key, vel uint8

		if msg.GetNoteOn(&channel, &key, &vel) {
			select {
			case keyChan <- key:
			default:
			}
		}
	}, midi.UseSysEx())
	if err != nil {
		return 0, err
	}
	defer stop()

	select {
	case key := <-keyChan:
		return key, nil
	case <-time.After(30 * time.Second):
		return 0, fmt.Errorf("timeout waiting for MIDI key press")
	}
}

func startMIDIListener(bridge *HueBridge, light *Light, calibration *MIDICalibration) error {
	fmt.Println("🎵 Starting MIDI listener... Press keys to control brightness!")
	fmt.Printf("   Left key (%d) = 0%% brightness\n", calibration.LeftKey)
	fmt.Printf("   Right key (%d) = 100%% brightness\n", calibration.RightKey)
	fmt.Println("   Press Ctrl+C to exit")

	ins := midi.GetInPorts()
	if len(ins) == 0 {
		return fmt.Errorf("no MIDI input devices found")
	}

	in := ins[0]

	stop, err := midi.ListenTo(in, func(msg midi.Message, timestampms int32) {
		var channel, key, vel uint8

		if msg.GetNoteOn(&channel, &key, &vel) {
			brightness := calculateBrightness(key, calibration)
			err := setLightBrightness(bridge, light, brightness)
			if err != nil {
				fmt.Printf("❌ Failed to set brightness: %v\n", err)
			} else {
				fmt.Printf("🎹 Key %d → %d%% brightness\n", key, brightness)
			}
		}
	}, midi.UseSysEx())
	if err != nil {
		return fmt.Errorf("failed to listen to MIDI device: %v", err)
	}
	defer stop()

	// Keep the program running
	fmt.Println("Press Enter to exit...")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadLine()

	return nil
}

func calculateBrightness(key uint8, calibration *MIDICalibration) int {
	if key <= calibration.LeftKey {
		return 0
	}
	if key >= calibration.RightKey {
		return 100
	}

	// Linear interpolation between left and right keys
	keyRange := float64(calibration.RightKey - calibration.LeftKey)
	keyPosition := float64(key - calibration.LeftKey)
	brightness := int((keyPosition / keyRange) * 100)

	if brightness < 0 {
		brightness = 0
	}
	if brightness > 100 {
		brightness = 100
	}

	return brightness
}

func setLightBrightness(bridge *HueBridge, light *Light, brightness int) error {
	// Convert percentage to Hue brightness scale (0-254)
	hueBrightness := int(float64(brightness) * 254.0 / 100.0)
	if hueBrightness < 1 && brightness > 0 {
		hueBrightness = 1 // Minimum brightness when not off
	}

	var requestBody string
	if brightness == 0 {
		requestBody = `{"on":false}`
	} else {
		requestBody = fmt.Sprintf(`{"on":true,"bri":%d}`, hueBrightness)
	}

	url := fmt.Sprintf("http://%s/api/%s/lights/%s/state", bridge.IP, bridge.Username, light.ID)

	req, err := http.NewRequest("PUT", url, strings.NewReader(requestBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
