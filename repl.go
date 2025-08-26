package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// A config struct to hold application state.
type config struct {
	NextLocationAreasURL string
	PrevLocationAreasURL *string // Use a pointer for nullable fields
}

// The main REPL function, now takes a pointer to the config.
func startRepl(cfg *config) {
	reader := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")
		reader.Scan()

		words := cleanInput(reader.Text())
		if len(words) == 0 {
			continue
		}

		commandName := words[0]
		command, ok := getCommands()[commandName]
		if !ok {
			fmt.Println("Unknown command")
			continue
		}

		// Pass the config to the command's callback.
		err := command.callback(cfg)
		if err != nil {
			fmt.Println(err)
		}
	}
}

// A struct for each CLI command. The callback now accepts the config.
type cliCommand struct {
	name        string
	description string
	callback    func(*config) error
}

// A map of all available commands.
func getCommands() map[string]cliCommand {
	return map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"map": {
			name:        "map",
			description: "Displays the next 20 location areas",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Displays the previous 20 location areas",
			callback:    commandMapb,
		},
	}
}

// A struct to unmarshal the JSON response from the PokeAPI.
type locationAreaResponse struct {
	Count    int     `json:"count"`
	Next     string  `json:"next"`
	Previous *string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

// Callback for the 'map' command.
func commandMap(cfg *config) error {
	res, err := http.Get(cfg.NextLocationAreasURL)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var response locationAreaResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}

	// Update the config with the new URLs from the response.
	cfg.NextLocationAreasURL = response.Next
	cfg.PrevLocationAreasURL = response.Previous

	for _, result := range response.Results {
		fmt.Println(result.Name)
	}
	return nil
}

// Callback for the 'mapb' (map back) command.
func commandMapb(cfg *config) error {
	if cfg.PrevLocationAreasURL == nil {
		return fmt.Errorf("you're on the first page")
	}

	res, err := http.Get(*cfg.PrevLocationAreasURL)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var response locationAreaResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}

	cfg.NextLocationAreasURL = response.Next
	cfg.PrevLocationAreasURL = response.Previous

	for _, result := range response.Results {
		fmt.Println(result.Name)
	}
	return nil
}

// Updated 'help' command callback.
func commandHelp(cfg *config) error {
	fmt.Println("\nWelcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println()
	for _, cmd := range getCommands() {
		fmt.Printf("%s: %s\n", cmd.name, cmd.description)
	}
	fmt.Println()
	return nil
}

// Updated 'exit' command callback.
func commandExit(cfg *config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

// Input cleaning function.
func cleanInput(text string) []string {
	output := strings.ToLower(text)
	words := strings.Fields(output)
	return words
}
