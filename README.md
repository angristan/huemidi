# HueMIDI 🎹💡

A Go CLI tool that allows you to control the brightness of your Philips Hue light bulbs using a MIDI piano keyboard. The leftmost key corresponds to 0% brightness, and the rightmost key corresponds to 100% brightness.

## Features

- 🔍 **Auto-discovery**: Automatically discovers your local Hue bridge
- 🎯 **Light Selection**: Choose from available light bulbs using a fuzzy finder
- 🎹 **MIDI Calibration**: Calibrate your keyboard by pressing the leftmost and rightmost keys
- 🌈 **Real-time Control**: Control brightness in real-time by pressing keys on your MIDI keyboard

## Prerequisites

- Go 1.21 or later
- A Philips Hue bridge on your local network
- A MIDI keyboard connected to your computer

## Installation

1. Clone this repository:

   ```bash
   git clone <your-repo-url>
   cd huemidi
   ```

2. Install dependencies:

   ```bash
   go mod tidy
   ```

3. Build the application:

   ```bash
   make build
   # or manually: go build -o huemidi
   ```

## Usage

1. **Connect your MIDI keyboard** to your computer
2. **Ensure your Hue bridge** is on the same network as your computer
3. **Run the application**:

   ```bash
   ./huemidi
   # or: make run
   ```

4. **Follow the on-screen instructions**:
   - The app will auto-discover your Hue bridge
   - Press the link button on your Hue bridge when prompted
   - Select a light bulb from the list using arrow keys and Enter
   - Calibrate your MIDI keyboard by pressing the leftmost and rightmost keys
   - Start playing! Press keys to control the brightness

## Environment Variables

- `HUE_USERNAME`: Set this to skip the authentication step on subsequent runs

Example:

```bash
export HUE_USERNAME=your_username_here
./huemidi
```

## How It Works

1. **Discovery**: Uses the official Hue discovery API to find your bridge
2. **Authentication**: Creates a new user on the bridge (requires pressing the link button)
3. **Light Selection**: Fetches available lights and presents them in a user-friendly list
4. **MIDI Calibration**: Maps your keyboard range to 0-100% brightness
5. **Real-time Control**: Listens for MIDI note-on events and maps key positions to brightness levels

## Dependencies

- `github.com/manifoldco/promptui` - Interactive prompts and selection
- `gitlab.com/gomidi/midi/v2` - MIDI input handling
- `github.com/tidwall/gjson` - JSON parsing for Hue API responses

## Troubleshooting

### MIDI Issues

- **No MIDI devices found**: Ensure your MIDI keyboard is connected and recognized by your system
- **Permission denied**: On Linux, you may need to add your user to the `audio` group

### Hue Bridge Issues

- **Bridge not found**: Ensure the bridge is on the same network and accessible
- **Authentication failed**: Make sure to press the link button within 30 seconds of the prompt

### General Issues

- **Build errors**: Ensure you have Go 1.21+ and run `go mod tidy`
- **Network issues**: Check firewall settings that might block bridge discovery

## License

MIT License
