package vhpackage

import (
	"fmt"
	"io/ioutil"
	"os"
)

type PlayerProfile struct {
	Version            int
	Name               string
	ID                 int64
	StartSeed          string
	OriginalSpawnPoint Vector3
	WorldData          map[int64]WorldPlayerData
	Stats              PlayerStats
	Player             *Player
}

type WorldPlayerData struct {
	SpawnPoint           Vector3
	HaveCustomSpawnPoint bool
	LogoutPoint          Vector3
	HaveLogoutPoint      bool
	DeathPoint           Vector3
	HaveDeathPoint       bool
	HomePoint            Vector3
	Map                  *Map
}

type Map struct {
	Version                 int
	TextureSize             int
	Explored                []bool
	Pins                    []Pin
	PublicReferencePosition bool
}

type Pin struct {
	Name      string
	Position  Vector3
	Type      int // TODO: enum for pin type
	IsChecked bool
}

type PlayerStats struct {
	Kills  int
	Deaths int
	Crafts int
	Builds int
}

type Player struct {
	Version               int
	MaxHealth             float32
	Health                float32
	Stamina               float32
	FirstSpawn            bool
	TimeSinceDeath        float32
	GuardianPower         string
	GuardianPowerCooldown float32
	Inventory             []*Item
	KnownRecipes          []string
	KnownStations         map[string]int
	KnownMaterial         []string
	ShownTutorials        []string
	Uniques               []string
	Trophies              []string
	KnownBiomes           []int // TODO: enum for biome names
	KnownTexts            map[string]string
	Beard                 string
	Hair                  string
	SkinColor             Vector3
	HairColor             Vector3
	PlayerModel           int
	Foods                 []*Food
	Skills                []*Skill
}

type Item struct {
	Name        string
	Stack       int
	Durability  float32
	Position    Vector2i
	Equiped     bool
	Quality     int
	Variant     int
	CrafterID   int64
	CrafterName string
}

type Food struct {
	Name    string
	Health  float32
	Stamina float32
}

type Skill struct {
	Type        int // TODO: enum for skill type
	Level       float32
	Accumulator float32
}

func NewPlayerProfileFromFile(file string) (*PlayerProfile, error) {
	filedata, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	return NewPlayerProfileFromData(filedata)
}

func NewPlayerProfileFromData(data []byte) (*PlayerProfile, error) {
	pkg := NewZPackage(data)

	playerPkg, err := pkg.ReadPackage()
	if err != nil {
		return nil, err
	}

	_, err = pkg.ReadPackage()
	if err != nil {
		return nil, err
	}

	p := &PlayerProfile{}
	return p, p.readPlayerProfile(playerPkg)
}

func (p *PlayerProfile) readPlayerProfile(pkg *ZPackage) error {
	var err error

	p.Version, err = pkg.ReadInt()
	if err != nil {
		return fmt.Errorf("cannot read player version: %w", err)
	}

	// Player stats
	if p.Version >= 28 {
		p.Stats.Kills, err = pkg.ReadInt()
		if err != nil {
			return fmt.Errorf("cannot read player kills: %w", err)
		}
		p.Stats.Deaths, err = pkg.ReadInt()
		if err != nil {
			return fmt.Errorf("cannot read player deaths: %w", err)
		}
		p.Stats.Crafts, err = pkg.ReadInt()
		if err != nil {
			return fmt.Errorf("cannot read player crafts: %w", err)
		}
		p.Stats.Builds, err = pkg.ReadInt()
		if err != nil {
			return fmt.Errorf("cannot read player builds: %w", err)
		}
	}

	// World player data
	worldPlayerDataCount, err := pkg.ReadInt()
	if err != nil {
		return fmt.Errorf("cannot read player world data count: %w", err)
	}

	p.WorldData = make(map[int64]WorldPlayerData)
	for i := 0; i < worldPlayerDataCount; i++ {
		key, err := pkg.ReadLong()
		if err != nil {
			return fmt.Errorf("cannot read world player key: %w", err)
		}

		wpd := WorldPlayerData{}
		wpd.HaveCustomSpawnPoint, err = pkg.ReadBool()
		if err != nil {
			return err
		}
		wpd.SpawnPoint, err = pkg.ReadVector3()
		if err != nil {
			return err
		}
		wpd.HaveLogoutPoint, err = pkg.ReadBool()
		if err != nil {
			return err
		}
		wpd.LogoutPoint, err = pkg.ReadVector3()
		if err != nil {
			return err
		}

		if p.Version >= 30 {
			wpd.HaveDeathPoint, err = pkg.ReadBool()
			if err != nil {
				return err
			}
			wpd.DeathPoint, err = pkg.ReadVector3()
			if err != nil {
				return err
			}
		}

		wpd.HomePoint, err = pkg.ReadVector3()

		if p.Version >= 29 {
			haveMapData, _ := pkg.ReadBool()
			if haveMapData {
				mapPkg, err := pkg.ReadPackage()
				if err != nil {
					return err
				}
				wpd.Map, err = readMapData(mapPkg)
				if err != nil {
					return err
				}
			}
		}

		p.WorldData[key] = wpd
	}

	// Player info
	p.Name, err = pkg.ReadString()
	if err != nil {
		return err
	}
	p.ID, err = pkg.ReadLong()
	if err != nil {
		return err
	}
	p.StartSeed, err = pkg.ReadString()
	if err != nil {
		return err
	}
	havePlayerData, err := pkg.ReadBool()
	if err != nil {
		return err
	}
	if havePlayerData {
		playerPkg, err := pkg.ReadPackage()
		if err != nil {
			return err
		}

		p.Player, err = readPlayerData(playerPkg)
		if err != nil {
			return err
		}
	}

	return nil
}

func readMapData(pkg *ZPackage) (*Map, error) {
	m := &Map{}
	var err error
	m.Version, err = pkg.ReadInt()
	if err != nil {
		return nil, err
	}
	m.TextureSize, err = pkg.ReadInt()
	if err != nil {
		return nil, err
	}
	m.Explored = make([]bool, m.TextureSize*m.TextureSize)
	err = pkg.read(&m.Explored)
	if err != nil {
		return nil, err
	}

	// pins
	if m.Version >= 2 {
		pinCount, err := pkg.ReadInt()
		if err != nil {
			fmt.Fprintln(os.Stderr, "MAP PINS", pinCount)
			return nil, err
		}
		m.Pins = make([]Pin, pinCount)
		for i := 0; i < pinCount; i++ {
			m.Pins[i], err = readPin(pkg)
			if err != nil {
				return nil, err
			}
		}
	}

	// public pos ref
	if m.Version >= 4 {
		m.PublicReferencePosition, err = pkg.ReadBool()
		if err != nil {
			return nil, err
		}
	}

	return m, nil
}

func readPin(pkg *ZPackage) (Pin, error) {
	pin := Pin{}

	pin.Name, _ = pkg.ReadString()
	pin.Position, _ = pkg.ReadVector3()
	pin.Type, _ = pkg.ReadInt()
	pin.IsChecked, _ = pkg.ReadBool()

	return pin, nil
}

func readPlayerData(pkg *ZPackage) (*Player, error) {
	p := &Player{}

	var err error
	p.Version, err = pkg.ReadInt()
	if err != nil {
		return nil, err
	}

	if p.Version >= 7 {
		p.MaxHealth, err = pkg.ReadSingle()
		if err != nil {
			return nil, err
		}
	}
	p.Health, err = pkg.ReadSingle()
	if err != nil {
		return nil, err
	}

	if p.Version >= 10 {
		p.Stamina, err = pkg.ReadSingle()
		if err != nil {
			return nil, err
		}
	}
	if p.Version >= 8 {
		p.FirstSpawn, err = pkg.ReadBool()
		if err != nil {
			return nil, err
		}
	}
	if p.Version >= 20 {
		p.TimeSinceDeath, err = pkg.ReadSingle()
		if err != nil {
			return nil, err
		}
	}
	if p.Version >= 23 {
		p.GuardianPower, err = pkg.ReadString()
		if err != nil {
			return nil, err
		}
	}
	if p.Version >= 24 {
		p.GuardianPowerCooldown, err = pkg.ReadSingle()
		if err != nil {
			return nil, err
		}
	}
	if p.Version == 2 {
		pkg.ReadZDOID()
	}

	// inventory
	p.Inventory, err = readInventory(pkg)
	if err != nil {
		return nil, fmt.Errorf("cannot read player inventory: %w", err)
	}

	// known recipes
	err = pkg.ReadIntoList(&p.KnownRecipes)
	if err != nil {
		return nil, err
	}

	// known stations
	if p.Version < 15 {
		// old version, skip part
		l := []string{}
		pkg.ReadIntoList(&l)
	} else {
		p.KnownStations = make(map[string]int)
		knownStationsCount, err := pkg.ReadInt()
		if err != nil {
			return nil, err
		}
		for i := 0; i < knownStationsCount; i++ {
			stationName, err := pkg.ReadString()
			if err != nil {
				return nil, err
			}
			stationLevel, err := pkg.ReadInt()
			if err != nil {
				return nil, err
			}
			p.KnownStations[stationName] = stationLevel
		}
	}

	// known material
	err = pkg.ReadIntoList(&p.KnownMaterial)
	if err != nil {
		return nil, err
	}

	// shown tutorials
	if p.Version < 19 || p.Version >= 21 {
		err = pkg.ReadIntoList(&p.ShownTutorials)
		if err != nil {
			return nil, err
		}
	}

	// uniques
	if p.Version >= 6 {
		err = pkg.ReadIntoList(&p.Uniques)
		if err != nil {
			return nil, err
		}
	}

	// trophies
	if p.Version >= 9 {
		err = pkg.ReadIntoList(&p.Trophies)
		if err != nil {
			return nil, err
		}
	}

	// known biomes
	if p.Version >= 18 {
		err = pkg.ReadIntoList(&p.KnownBiomes)
		if err != nil {
			return nil, err
		}
	}

	// known texts
	if p.Version >= 22 {
		knownTextCount, err := pkg.ReadInt()
		if err != nil {
			return nil, err
		}
		p.KnownTexts = make(map[string]string)
		for i := 0; i < knownTextCount; i++ {
			key, err := pkg.ReadString()
			if err != nil {
				return nil, err
			}
			value, err := pkg.ReadString()
			if err != nil {
				return nil, err
			}
			p.KnownTexts[key] = value
		}
	}

	// beard and hair
	if p.Version >= 4 {
		p.Beard, err = pkg.ReadString()
		if err != nil {
			return nil, err
		}
		p.Hair, err = pkg.ReadString()
		if err != nil {
			return nil, err
		}
	}

	// skin and hair color
	if p.Version >= 5 {
		p.SkinColor, err = pkg.ReadVector3()
		if err != nil {
			return nil, err
		}
		p.HairColor, err = pkg.ReadVector3()
		if err != nil {
			return nil, err
		}
	}

	// player model
	if p.Version >= 11 {
		p.PlayerModel, err = pkg.ReadInt()
		if err != nil {
			return nil, err
		}
	}

	// food consumed
	if p.Version >= 12 {
		foodCount, err := pkg.ReadInt()
		if err != nil {
			return nil, err
		}

		for i := 0; i < foodCount; i++ {
			if p.Version >= 14 {
				food, err := readFood(pkg, p.Version)
				if err != nil {
					return nil, err
				}
				p.Foods = append(p.Foods, food)
			} else {
				// old version, skip data
				pkg.ReadString()
				pkg.ReadSingle()
				pkg.ReadSingle()
				pkg.ReadSingle()
				pkg.ReadSingle()
				pkg.ReadSingle()
				pkg.ReadSingle()
				if p.Version >= 13 {
					pkg.ReadSingle()
				}
			}
		}
	}

	// skills
	if p.Version >= 17 {
		p.Skills, err = readSkills(pkg)
		if err != nil {
			return nil, err
		}
	}

	return p, nil
}

func readInventory(pkg *ZPackage) ([]*Item, error) {
	version, err := pkg.ReadInt()
	if err != nil {
		return nil, err
	}
	count, err := pkg.ReadInt()
	if err != nil {
		return nil, err
	}

	inventory := make([]*Item, count)
	for i := 0; i < count; i++ {
		item, err := readInventoryItem(pkg, version)
		if err != nil {
			return nil, err
		}

		inventory[i] = item
	}

	return inventory, nil
}

func readInventoryItem(pkg *ZPackage, version int) (*Item, error) {
	item := &Item{}

	item.Name, _ = pkg.ReadString()
	item.Stack, _ = pkg.ReadInt()
	item.Durability, _ = pkg.ReadSingle()
	item.Position, _ = pkg.ReadVector2i()
	item.Equiped, _ = pkg.ReadBool()
	item.Quality = 1
	if version >= 101 {
		item.Quality, _ = pkg.ReadInt()
	}
	if version >= 102 {
		item.Variant, _ = pkg.ReadInt()
	}
	if version >= 103 {
		item.CrafterID, _ = pkg.ReadLong()
		item.CrafterName, _ = pkg.ReadString()
	}

	return item, nil
}

func readFood(pkg *ZPackage, version int) (*Food, error) {
	food := &Food{}

	food.Name, _ = pkg.ReadString()
	food.Health, _ = pkg.ReadSingle()

	if version >= 16 {
		food.Stamina, _ = pkg.ReadSingle()
	}

	return food, nil
}

func readSkills(pkg *ZPackage) ([]*Skill, error) {
	version, err := pkg.ReadInt()
	if err != nil {
		return nil, err
	}

	count, err := pkg.ReadInt()
	if err != nil {
		return nil, err
	}

	skills := make([]*Skill, count)
	for i := 0; i < count; i++ {
		skill := &Skill{}

		skill.Type, err = pkg.ReadInt()
		if err != nil {
			return nil, err
		}
		skill.Level, err = pkg.ReadSingle()
		if err != nil {
			return nil, err
		}
		if version >= 2 {
			skill.Accumulator, err = pkg.ReadSingle()
			if err != nil {
				return nil, err
			}
		}

		skills[i] = skill
	}

	return skills, nil
}
