package handlers

import "fmt"

var allowedColours = map[string]bool{
	"grey":   true,
	"red":    true,
	"orange": true,
	"gold":   true,
	"green":  true,
	"blue":   true,
	"purple": true,
	"brown":  true,
}

var allowedIcons = map[string]bool{
	"position-marker": true,
	"castle":          true,
	"crossed-swords":  true,
	"skull":           true,
	"treasure-map":    true,
	"campfire":        true,
	"forest-camp":     true,
	"mountain-cave":   true,
	"village":         true,
	"temple-gate":     true,
	"sailboat":        true,
	"crown":           true,
	"dragon-head":     true,
	"tombstone":       true,
	"bridge":          true,
	"mine-entrance":   true,
	"tower-flag":      true,
	"cauldron":        true,
	"wood-cabin":      true,
	"portal":          true,
}

// ValidatePinColour returns an error if the colour is not in the allowed set.
func ValidatePinColour(colour string) error {
	if !allowedColours[colour] {
		return fmt.Errorf("invalid colour %q: must be one of grey, red, orange, gold, green, blue, purple, brown", colour)
	}
	return nil
}

// ValidatePinIcon returns an error if the icon is not in the allowed set.
func ValidatePinIcon(icon string) error {
	if !allowedIcons[icon] {
		return fmt.Errorf("invalid icon %q: must be one of position-marker, castle, crossed-swords, skull, treasure-map, campfire, forest-camp, mountain-cave, village, temple-gate, sailboat, crown, dragon-head, tombstone, bridge, mine-entrance, tower-flag, cauldron, wood-cabin, portal", icon)
	}
	return nil
}
