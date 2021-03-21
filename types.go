package vhpackage

import "fmt"

// ZDO represents a data object.
type ZDO struct {
	// ZDO metadata
	UID           ZDOID `json:"uid"`
	OwnerRevision uint  `json:"owner_revision"`
	DataRevision  uint  `json:"data_revision"`
	Persistent    bool  `json:"persistent"`
	Owner         int64 `json:"owner"`
	TimeCreated   int64 `json:"time_created"`
	PGWVersion    int   `json:"pgw_version"`

	// ZDO type and location
	Type     int8       `json:"type"`
	Distant  bool       `json:"distant"`
	Prefab   int        `json:"prefab"`
	Sector   Vector2i   `json:"sector"`
	Position Vector3    `json:"position"`
	Rotation Quaternion `json:"rotation"`

	// ZDO properties
	Floats      map[int]float32    `json:"floats"`
	Vectors     map[int]Vector3    `json:"vectors"`
	Quaternions map[int]Quaternion `json:"quaternions"`
	Ints        map[int]int        `json:"ints"`
	Longs       map[int]int64      `json:"longs"`
	Strings     map[int]string     `json:"strings"`
}

func (zdo *ZDO) LoadZDO(pkg *ZPackage, version int) error {
	var err error

	zdo.OwnerRevision, err = pkg.ReadUInt()
	if err != nil {
		return fmt.Errorf("cannot read owner revision: %w", err)
	}
	zdo.DataRevision, err = pkg.ReadUInt()
	if err != nil {
		return fmt.Errorf("cannot read data revision: %w", err)
	}
	zdo.Persistent, err = pkg.ReadBool()
	if err != nil {
		return fmt.Errorf("cannot read persistent: %w", err)
	}
	zdo.Owner, err = pkg.ReadLong()
	if err != nil {
		return fmt.Errorf("cannot read owner: %w", err)
	}
	zdo.TimeCreated, err = pkg.ReadLong()
	if err != nil {
		return fmt.Errorf("cannot read time created: %w", err)
	}
	zdo.PGWVersion, err = pkg.ReadInt()
	if err != nil {
		return fmt.Errorf("cannot read PGW version: %w", err)
	}

	if version >= 16 && version < 24 {
		_, err = pkg.ReadInt()
		if err != nil {
			return fmt.Errorf("cannot skip int: %w", err)
		}
	}
	if version >= 23 {
		zdo.Type, err = pkg.ReadSByte()
		if err != nil {
			return fmt.Errorf("cannot read ZDO type: %w", err)
		}
	}
	if version >= 22 {
		zdo.Distant, err = pkg.ReadBool()
		if err != nil {
			return fmt.Errorf("cannot read distant: %w", err)
		}
	}
	if version < 13 {
		pkg.ReadChar()
		pkg.ReadChar()
	}
	if version >= 17 {
		zdo.Prefab, err = pkg.ReadInt()
		if err != nil {
			return fmt.Errorf("cannot read prefab: %w", err)
		}
	}

	zdo.Sector, err = pkg.ReadVector2i()
	if err != nil {
		return fmt.Errorf("cannot read sector: %w", err)
	}
	zdo.Position, err = pkg.ReadVector3()
	if err != nil {
		return fmt.Errorf("cannot read position: %w", err)
	}
	zdo.Rotation, err = pkg.ReadQuaternion()
	if err != nil {
		return fmt.Errorf("cannot read rotation: %w", err)
	}

	// Floats
	c, err := pkg.ReadChar()
	if err != nil {
		return fmt.Errorf("cannot read number of floats: %w", err)
	}
	num := int(c)
	if num > 0 {
		zdo.Floats = make(map[int]float32)
		for i := 0; i < num; i++ {
			key, err := pkg.ReadInt()
			if err != nil {
				return fmt.Errorf("cannot read float key: %w", err)
			}
			zdo.Floats[key], err = pkg.ReadSingle()
			if err != nil {
				return fmt.Errorf("cannot read float value: %w", err)
			}
		}
	}

	// Vector3s
	c, err = pkg.ReadChar()
	if err != nil {
		return fmt.Errorf("cannot read number of vector3s: %w", err)
	}
	num = int(c)
	if num > 0 {
		zdo.Vectors = make(map[int]Vector3)
		for i := 0; i < num; i++ {
			key, err := pkg.ReadInt()
			if err != nil {
				return fmt.Errorf("cannot read vector3 key: %w", err)
			}
			zdo.Vectors[key], err = pkg.ReadVector3()
			if err != nil {
				return fmt.Errorf("cannot read vector3 value: %w", err)
			}
		}
	}

	// Quaternions
	c, err = pkg.ReadChar()
	if err != nil {
		return fmt.Errorf("cannot read number of quaternions: %w", err)
	}
	num = int(c)
	if num > 0 {
		zdo.Quaternions = make(map[int]Quaternion)
		for i := 0; i < num; i++ {
			key, err := pkg.ReadInt()
			if err != nil {
				return fmt.Errorf("cannot read quaternion key: %w", err)
			}
			zdo.Quaternions[key], err = pkg.ReadQuaternion()
			if err != nil {
				return fmt.Errorf("cannot read quaternion value: %w", err)
			}
		}
	}

	// Ints
	c, err = pkg.ReadChar()
	if err != nil {
		return fmt.Errorf("cannot read number of ints: %w", err)
	}
	num = int(c)
	if num > 0 {
		zdo.Ints = make(map[int]int)
		for i := 0; i < num; i++ {
			key, err := pkg.ReadInt()
			if err != nil {
				return fmt.Errorf("cannot read int key: %w", err)
			}
			zdo.Ints[key], err = pkg.ReadInt()
			if err != nil {
				return fmt.Errorf("cannot read int value: %w", err)
			}
		}
	}

	// Longs
	c, err = pkg.ReadChar()
	if err != nil {
		return fmt.Errorf("cannot read number of longs: %w", err)
	}
	num = int(c)
	if num > 0 {
		zdo.Longs = make(map[int]int64)
		for i := 0; i < num; i++ {
			key, err := pkg.ReadInt()
			if err != nil {
				return fmt.Errorf("cannot read long key: %w", err)
			}
			zdo.Longs[key], err = pkg.ReadLong()
			if err != nil {
				return fmt.Errorf("cannot read long value: %w", err)
			}
		}
	}

	// Strings
	c, err = pkg.ReadChar()
	if err != nil {
		return fmt.Errorf("cannot read number of strings: %w", err)
	}
	num = int(c)
	if num > 0 {
		zdo.Strings = make(map[int]string)
		for i := 0; i < num; i++ {
			key, err := pkg.ReadInt()
			if err != nil {
				return fmt.Errorf("cannot read string key: %w", err)
			}
			zdo.Strings[key], err = pkg.ReadString()
			if err != nil {
				return fmt.Errorf("cannot read string value: %w", err)
			}
		}
	}

	return nil
}

// ZDOID represents data object ID.
type ZDOID struct {
	UserID int64  `json:"user_id"`
	ID     uint32 `json:"id"`
}

func (zid ZDOID) String() string {
	return fmt.Sprintf("%d:%d", zid.UserID, zid.ID)
}

// Vector3 represents Unity.Vector3 type.
type Vector3 struct {
	X, Y, Z float32
}

// Vector2i represents Unity.Vector2i type.
type Vector2i struct {
	X, Y int32
}

// Quaternion represents Unity.Quaternion type.
type Quaternion struct {
	X, Y, Z, W float32
}
