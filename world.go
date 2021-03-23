package vhpackage

import (
	"fmt"
	"io/ioutil"
)

type LocationInstance struct {
	Name      string  `json:"name"`
	Position  Vector3 `json:"position"`
	Generated bool    `json:"generated"`
}

type RandomEvent struct {
	Text     string  `json:"text"`
	Time     float32 `json:"time"`
	Position Vector3 `json:"position"`
}

type WorldMetadata struct {
	Version         int    `json:"version"`
	Name            string `json:"name"`
	SeedName        string `json:"seed_name"`
	Seed            int    `json:"seed"`
	UID             int64  `json:"uid"`
	WorldGenVersion int    `json:"world_gen_version"`
}

// World represents world data.
type World struct {
	Metadata *WorldMetadata `json:"metadata"`

	Version int     `json:"version"`
	NetTime float64 `json:"net_time"`

	// ZDO
	ZDOs     []*ZDO           `json:"zdos"`
	DeadZDOs map[string]int64 `json:"dead_zdos"`

	// ZoneSystem
	GeneratedZones     []Vector2i         `json:"generated_zones"`
	LocationVersion    int                `json:"location_version"`
	PGWVersion         int                `json:"pgw_version"`
	LocationsGenerated bool               `json:"locations_generated"`
	GlobalKeys         []string           `json:"global_keys"`
	LocationInstances  []LocationInstance `json:"location_instances"`

	// RandEventSystem
	EventTimer float32      `json:"event_timer"`
	Event      *RandomEvent `json:"event,omitempty"` // Only version < 25
}

func NewWorldFromFile(metaPath, dbPath string) (*World, error) {
	w := &World{}

	if metaPath != "" {
		if err := w.loadMetadata(metaPath); err != nil {
			return nil, err
		}
	}

	if dbPath != "" {
		if err := w.loadData(dbPath); err != nil {
			return nil, err
		}
	}

	return w, nil
}

func (w *World) loadMetadata(file string) error {
	filedata, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	pkg := NewZPackage(filedata)
	pkg, err = pkg.ReadPackage()
	if err != nil {
		return err
	}

	return w.readMetadata(pkg)
}

func (w *World) readMetadata(pkg *ZPackage) error {
	// Read world version.
	version, err := pkg.ReadInt()
	if err != nil {
		return fmt.Errorf("Failed to read world version: %w", err)
	}

	name, err := pkg.ReadString()
	if err != nil {
		return fmt.Errorf("Failed to read world name: %w", err)
	}

	seedName, err := pkg.ReadString()
	if err != nil {
		return fmt.Errorf("Failed to read world seed name: %w", err)
	}

	seed, err := pkg.ReadInt()
	if err != nil {
		return fmt.Errorf("Failed to read world seed: %w", err)
	}

	uid, err := pkg.ReadLong()
	if err != nil {
		return fmt.Errorf("Failed to read world UID: %w", err)
	}

	// Only read world generation version if world version is >= 26.
	var genVersion int
	if version >= 26 {
		v, err := pkg.ReadInt()
		if err != nil {
			return fmt.Errorf("Failed to read world generation version: %w", err)
		}
		genVersion = v
	}

	w.Metadata = &WorldMetadata{
		Version:         version,
		Name:            name,
		SeedName:        seedName,
		Seed:            seed,
		UID:             uid,
		WorldGenVersion: genVersion,
	}

	return nil
}

func (w *World) loadData(file string) error {
	filedata, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	pkg := NewZPackage(filedata)
	return w.readData(pkg)
}

func (w *World) readData(pkg *ZPackage) error {
	// World version
	version, err := pkg.ReadInt()
	if err != nil {
		return fmt.Errorf("cannot read world version: %w", err)
	}
	w.Version = version
	// World uptime
	if version >= 4 {
		w.NetTime, err = pkg.ReadDouble()
		if err != nil {
			return fmt.Errorf("cannot read net time: %w", err)
		}
	}

	// ZDOMan
	if err := w.readZDOMan(pkg); err != nil {
		return fmt.Errorf("cannot read ZDOMan section: %w", err)
	}

	// ZoneSystem
	if err := w.readZoneSystem(pkg); err != nil {
		return fmt.Errorf("cannot read ZoneSystem section: %w", err)
	}

	// RandEventSystem
	if err := w.readRandEventSystem(pkg); err != nil {
		return fmt.Errorf("cannot read RandEventSystem section: %w", err)
	}

	return nil
}

func (w *World) readZDOMan(pkg *ZPackage) error {
	// ZDOMan metadata
	_, err := pkg.ReadLong() // ID, Skipping this value
	if err != nil {
		return fmt.Errorf("cannot skip value: %w", err)
	}
	_, err = pkg.ReadUInt() // Next UID, skipping
	if err != nil {
		return fmt.Errorf("cannot skip value: %w", err)
	}

	// ZDOs
	zdoCount, err := pkg.ReadInt()
	if err != nil {
		return fmt.Errorf("cannot read zdo count: %w", err)
	}
	for i := 0; i < zdoCount; i++ {
		zdo := &ZDO{}
		zdo.UID, err = pkg.ReadZDOID()
		if err != nil {
			return fmt.Errorf("(ZDO #%d) cannot read ZDOID: %w", i, err)
		}

		zdoPkg, err := pkg.ReadPackage()
		if err != nil {
			return fmt.Errorf("(ZDO #%d) cannot read ZDO: %w", i, err)
		}

		err = zdo.LoadZDO(zdoPkg, w.Version)
		if err != nil {
			return fmt.Errorf("(ZDO #%d) cannot load ZDO: %w", i, err)
		}

		w.ZDOs = append(w.ZDOs, zdo)
	}

	// Dead ZDOs
	w.DeadZDOs = make(map[string]int64)
	deadZdoCount, err := pkg.ReadInt()
	if err != nil {
		return fmt.Errorf("cannot read dead zdo count: %w", err)
	}
	for i := 0; i < deadZdoCount; i++ {
		key, err := pkg.ReadZDOID()
		if err != nil {
			return err
		}
		value, err := pkg.ReadLong()
		if err != nil {
			return err
		}

		w.DeadZDOs[key.String()] = value
	}

	return nil
}

func (w *World) readZoneSystem(pkg *ZPackage) error {
	generatedZoneCount, err := pkg.ReadInt()
	if err != nil {
		return fmt.Errorf("cannot read generated zone count: %w", err)
	}
	for i := 0; i < generatedZoneCount; i++ {
		z, err := pkg.ReadVector2i()
		if err != nil {
			return err
		}

		w.GeneratedZones = append(w.GeneratedZones, z)
	}

	if w.Version < 13 {
		return nil
	}

	w.PGWVersion, err = pkg.ReadInt()
	if err != nil {
		return fmt.Errorf("cannot read world PGW version: %w", err)
	}

	if w.Version >= 21 {
		w.LocationVersion, err = pkg.ReadInt()
		if err != nil {
			return fmt.Errorf("cannot read location version: %w", err)
		}
	}

	if w.Version >= 14 {
		globalKeysCount, err := pkg.ReadInt()
		if err != nil {
			return fmt.Errorf("cannot read global keys count: %w", err)
		}

		for i := 0; i < globalKeysCount; i++ {
			globalKey, err := pkg.ReadString()
			if err != nil {
				return fmt.Errorf("cannot read global key: %w", err)
			}

			w.GlobalKeys = append(w.GlobalKeys, globalKey)
		}
	}

	if w.Version < 18 {
		return nil
	}

	if w.Version >= 20 {
		w.LocationsGenerated, err = pkg.ReadBool()
		if err != nil {
			return fmt.Errorf("cannot read locations generated: %w", err)
		}
	}

	locationInstancesCount, err := pkg.ReadInt()
	if err != nil {
		return fmt.Errorf("cannot read location instances count: %w", err)
	}

	for i := 0; i < locationInstancesCount; i++ {
		loc := LocationInstance{}

		loc.Name, err = pkg.ReadString()
		if err != nil {
			return fmt.Errorf("cannot read location name: %w", err)
		}

		loc.Position, err = pkg.ReadVector3()
		if err != nil {
			return fmt.Errorf("cannot read location position: %w", err)
		}

		if w.Version >= 19 {
			loc.Generated, err = pkg.ReadBool()
			if err != nil {
				return fmt.Errorf("cannot read location generated: %w", err)
			}
		}
	}

	return nil
}

func (w *World) readRandEventSystem(pkg *ZPackage) error {
	var err error

	w.EventTimer, err = pkg.ReadSingle()
	if err != nil {
		return fmt.Errorf("cannot read event timer: %w", err)
	}

	if w.Version < 25 {
		return nil
	}

	evt := &RandomEvent{}
	evt.Text, err = pkg.ReadString()
	if err != nil {
		return fmt.Errorf("cannot read random event text: %w", err)
	}
	evt.Time, err = pkg.ReadSingle()
	if err != nil {
		return fmt.Errorf("cannot read random event time: %w", err)
	}
	evt.Position, err = pkg.ReadVector3()
	if err != nil {
		return fmt.Errorf("cannot read random event position: %w", err)
	}
	w.Event = evt

	return nil
}
