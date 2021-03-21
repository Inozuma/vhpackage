package vhpackage

import (
	"fmt"
	"io/ioutil"
)

type WorldPlayerData struct {
	SpawnPoint           Vector3
	HaveCustomSpawnPoint bool
	LogoutPoint          Vector3
	HaveLogoutPoint      bool
	DeathPoint           Vector3
	HaveDeathPoint       bool
	HomePoint            Vector3
	MapData              []byte
}

type PlayerStats struct {
	Kills  int
	Deaths int
	Crafts int
	Builds int
}

type PlayerProfile struct {
	Version            int
	PlayerName         string
	PlayerID           int64
	StartSeed          string
	OriginalSpawnPoint Vector3
	WorldData          map[int64]WorldPlayerData
	Stats              PlayerStats
	PlayerData         []byte
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
			haveMapData, err := pkg.ReadBool()
			if haveMapData {
				wpd.MapData, err = pkg.ReadByteArray()
				if err != nil {
					return err
				}
			}
		}

		p.WorldData[key] = wpd
	}

	// Player info
	p.PlayerName, err = pkg.ReadString()
	if err != nil {
		return err
	}
	p.PlayerID, err = pkg.ReadLong()
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
		p.PlayerData, err = pkg.ReadByteArray()
		if err != nil {
			return err
		}
	}

	return nil
}
