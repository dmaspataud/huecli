package main

import (
	hue "GoHue"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

var (
	configFile = strings.Join([]string{os.Getenv("HOME"), "/.config/huecli"}, "")
	colorList  = map[string][2]float32{
		"DEFAULT": [2]float32{0.4571, 0.4097},
		"RED":     [2]float32{0.6915, 0.3083},
		"GREEN":   [2]float32{0.0139, 0.7502},
		"BLUE":    [2]float32{0.1096, 0.0868},
		"PURPLE":  [2]float32{0.1611, 0.0138},
		"ORANGE":  [2]float32{0.5752, 0.4242},
		"YELLOW":  [2]float32{0.5125, 0.4866},
	}
	confTemplate = []byte("BridgeIP =\nBridgeToken =\n")
	usage        = `Usage: huecli [option] [args]
	
Options:

  status : Give a status of current lights bound to the bridge.
  color : Set the color of the targeted light.
  brightness : Set the brightness of the targeted lights.
  on : Switch the targeted lights on.
  off : Switch the targeted lights off.`
)

// Config structure contain decoded conf.toml data.
type Config struct {
	BridgeIP    string
	BridgeToken string
}

func main() {

	config := loadConf(configFile)
	bridge, err := hue.NewBridge(config.BridgeIP)
	if err != nil {
		fmt.Println("Could not connect : ", err)
	}
	if err := bridge.Login(config.BridgeToken); err != nil {
		fmt.Println("Could not authenticate with Hue Bridge :", err)
	}

	if len(os.Args) < 2 {
		fmt.Println(usage)
		os.Exit(0)
	}
	if os.Args[1] == "off" && len(os.Args) >= 3 {
		switchOff(parseLights(os.Args[2:], bridge))
	}
	if os.Args[1] == "on" && len(os.Args) >= 3 {
		switchOn(parseLights(os.Args[2:], bridge))
	}
	if os.Args[1] == "color" && len(os.Args) >= 4 {
		inputColor := strings.ToUpper(os.Args[2])
		if _, ok := colorList[inputColor]; ok {
			setColor(parseLights(os.Args[3:], bridge), colorList[inputColor])
		}
	}
	if os.Args[1] == "brightness" && len(os.Args) >= 4 {
		// os.Args[2] <= 100 && os.Args[2] >= 0
		inputBrightness, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Printf("Could not interpret brightness value %v : %v.\n", os.Args[2], err)
		}
		setBrightness(parseLights(os.Args[3:], bridge), inputBrightness)
	}
	if os.Args[1] == "status" {
		getStatus(bridge)
	}
}

// function to load configuration
func loadConf(path string) Config {
	var data Config
	// If file does no exist, create a template file.
	if _, err := os.Stat(configFile); err != nil {
		fmt.Println("Configuration file does not currently exist. Creating a template.")
		file, err := os.Create(configFile)
		if err != nil {
			fmt.Println("Could not create configuration file.")
		}
		defer file.Close()
		ioutil.WriteFile(configFile, confTemplate, 0644)
	}
	// Else, try to decode and return configuration data.
	if _, err := toml.DecodeFile(path, &data); err != nil {
		fmt.Println("Could not decode configuration file : ", err)
	}
	return data
}

// function to turn light off and print message. (arg : light)
func switchOff(target []hue.Light) {
	for _, eachLight := range target {
		eachLight.Off()
	}
}

// function to turn light on and print message. (arg : light)
func switchOn(target []hue.Light) {
	for _, eachLight := range target {
		eachLight.On()
	}
}

// function to change color and print message (arg : light, color)
func setColor(target []hue.Light, color [2]float32) {
	for _, eachLight := range target {
		err := eachLight.SetColor(&color) // TODO : handle error
		if err != nil {
			fmt.Printf("Could not change %v color : %v.\n", eachLight.Name, err)
		}
	}
}

// function to change luminosity and print message (arg : light, power)
func setBrightness(target []hue.Light, percent int) {
	for _, eachLight := range target {
		err := eachLight.SetBrightness(percent)
		if err != nil {
			fmt.Printf("Could not change %v brightness : %v\n", eachLight.Name, err)
		}
	}
}

// function to show current lights status
func getStatus(bridge *hue.Bridge) {
	allLights := getLights(bridge)
	fmt.Printf("%-15s %-15s\n", "LIGHT", "STATE")
	for _, eachLight := range allLights {
		if eachLight.State.On == true {
			fmt.Printf("%-15v %-15v\n", eachLight.Name, "\x1b[32;1mON\x1b[0m")
		} else {
			fmt.Printf("%-15v %-15v\n", eachLight.Name, "\x1b[31;1mOFF\x1b[0m")
		}

	}
}

func getLights(bridge *hue.Bridge) []hue.Light {
	allLights, err := bridge.GetAllLights()
	if err != nil {
		fmt.Println("Could not get light list. : ", err)
	}
	return allLights
}

func parseLights(inputLights []string, bridge *hue.Bridge) []hue.Light {
	allLights := getLights(bridge)
	results := make([]hue.Light, 0)
	for _, eachInput := range inputLights {
		for _, eachLight := range allLights {
			if eachInput == eachLight.Name {
				results = append(results, eachLight)
			}
		}
	}
	return results
}
