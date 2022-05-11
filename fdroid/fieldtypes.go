package fdroid

import (
	"encoding/hex"
	"strconv"
	"time"
)

type HexVal []byte

func (hv *HexVal) String() string {
	return hex.EncodeToString(*hv)
}

func (hv *HexVal) UnmarshalText(text []byte) error {
	b, err := hex.DecodeString(string(text))
	if err != nil {
		return err
	}
	*hv = b
	return nil
}

type UnixDate struct {
	time.Time
}

func (ud *UnixDate) String() string {
	return ud.Format("2006-01-02")
}

func (ud *UnixDate) UnmarshalJSON(data []byte) error {
	msec, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return err
	}
	t := time.Unix(msec/1000, 0).UTC()
	*ud = UnixDate{t}
	return nil
}
