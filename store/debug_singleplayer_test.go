package store_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/neper-stars/houston/store"
)

func TestDebugSingleplayer(t *testing.T) {
	data, err := os.ReadFile("../testdata/scenario-singleplayer/2483/Game.m1")
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	gs := store.New()
	err = gs.AddFile("Game.m1", data)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	player, ok := gs.Player(0)
	if !ok {
		t.Fatal("Player not found")
	}

	fmt.Println("=== Player Info ===")
	fmt.Printf("PlayerNumber: %d\n", player.PlayerNumber)
	fmt.Printf("NamePlural: %s\n", player.NamePlural)
	fmt.Printf("PRT: %d\n", player.PRT)

	fmt.Println("\n=== Tech Levels ===")
	fmt.Printf("Energy: %d\n", player.Tech.Energy)
	fmt.Printf("Weapons: %d\n", player.Tech.Weapons)
	fmt.Printf("Propulsion: %d\n", player.Tech.Propulsion)
	fmt.Printf("Construction: %d\n", player.Tech.Construction)
	fmt.Printf("Electronics: %d\n", player.Tech.Electronics)
	fmt.Printf("Biotech: %d\n", player.Tech.Biotech)
	fmt.Printf("Sum: %d\n", player.Tech.Energy+player.Tech.Weapons+player.Tech.Propulsion+
		player.Tech.Construction+player.Tech.Electronics+player.Tech.Biotech)

	ownedPlanets := gs.PlanetsByOwner(0)
	fmt.Printf("\n=== Owned Planets: %d ===\n", len(ownedPlanets))

	totalPop := int64(0)
	for i, planet := range ownedPlanets {
		totalPop += planet.Population
		fmt.Printf("Planet %d: Pop=%d (file=%d), Factories=%d, HasStarbase=%v\n",
			i, planet.Population, planet.Population/100, planet.Factories, planet.HasStarbase)
	}
	fmt.Printf("Total Population: %d (file units: %d)\n", totalPop, totalPop/100)

	// Score calculation
	sc := gs.CalculateScore(0)
	fmt.Println("\n=== Score Calculation ===")
	fmt.Printf("PlanetCount: %d\n", sc.PlanetCount)
	fmt.Printf("PlanetPopScore: %d\n", sc.PlanetPopScore)
	fmt.Printf("TotalResources: %d\n", sc.TotalResources)
	fmt.Printf("ResourceScore: %d\n", sc.ResourceScore)
	fmt.Printf("StarbaseCount: %d\n", sc.StarbaseCount)
	fmt.Printf("StarbaseScore: %d\n", sc.StarbaseScore)
	fmt.Printf("TechScore: %d\n", sc.TechScore)
	fmt.Printf("UnarmedShips: %d\n", sc.UnarmedShips)
	fmt.Printf("EscortShips: %d\n", sc.EscortShips)
	fmt.Printf("CapitalShips: %d\n", sc.CapitalShips)
	fmt.Printf("ShipScore: %d\n", sc.ShipScore)
	fmt.Printf("Total Score: %d\n", sc.Score)

	fmt.Println("\n=== Expected Values (from scores.png) ===")
	fmt.Println("Planets: 11, Starbases: 1, Unarmed: 6, Escort: 2, Capital: 0")
	fmt.Println("Tech: 76, Resources: 17k, Score: 838")

	// Calculate expected breakdown
	fmt.Println("\n=== Expected Breakdown ===")
	expResourceScore := 17000 / 30
	expStarbaseScore := 1 * 3
	expTechScore := 76
	// Ship score with 11 planets: unarmed=min(6,11)=6, escort=min(2,11)=2
	// shipScore = 6/2 + 2 = 3 + 2 = 5
	expShipScore := 6/2 + 2
	expPlanetPopScore := 838 - expResourceScore - expStarbaseScore - expTechScore - expShipScore
	fmt.Printf("ResourceScore (17000/30): %d\n", expResourceScore)
	fmt.Printf("StarbaseScore (1*3): %d\n", expStarbaseScore)
	fmt.Printf("TechScore: %d\n", expTechScore)
	fmt.Printf("ShipScore (6/2 + 2): %d\n", expShipScore)
	fmt.Printf("Implied PlanetPopScore: %d\n", expPlanetPopScore)
}
