package kenall

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

type (
	// A Version is the version-controlled date of the retrieved data.
	Version time.Time
	// A NullString represents a string that may be null.
	NullString struct {
		String string
		Valid  bool // Valid is true if String is not NULL
	}
)

type (
	// An Address is an address associated with the postal code defined by JP POST.
	Address struct {
		JISX0402           string `json:"jisx0402"`
		OldCode            string `json:"old_code"`
		PostalCode         string `json:"postal_code"`
		PrefectureKana     string `json:"prefecture_kana"`
		CityKana           string `json:"city_kana"`
		TownKana           string `json:"town_kana"`
		TownKanaRaw        string `json:"town_kana_raw"`
		Prefecture         string `json:"prefecture"`
		City               string `json:"city"`
		Town               string `json:"town"`
		Koaza              string `json:"koaza"`
		KyotoStreet        string `json:"kyoto_street"`
		Building           string `json:"building"`
		Floor              string `json:"floor"`
		TownPartial        bool   `json:"town_partial"`
		TownAddressedKoaza bool   `json:"town_addressed_koaza"`
		TownChome          bool   `json:"town_chome"`
		TownMulti          bool   `json:"town_multi"`
		TownRaw            string `json:"town_raw"`
		Corporation        struct {
			Name        string      `json:"name"`
			NameKana    string      `json:"name_kana"`
			BlockLot    string      `json:"block_lot"`
			BlockLotNum NullString  `json:"block_lot_num"`
			PostOffice  string      `json:"post_office"`
			CodeType    json.Number `json:"code_type"`
		} `json:"corporation"`
	}
	// A City is a city associated with the prefecture code defined by JIS X 0401.
	City struct {
		JISX0402       string `json:"jisx0402"`
		PrefectureCode string `json:"prefecture_code"`
		CityCode       string `json:"city_code"`
		PrefectureKana string `json:"prefecture_kana"`
		CityKana       string `json:"city_kana"`
		Prefecture     string `json:"prefecture"`
		City           string `json:"city"`
	}
)

var (
	// nolint: gochecknoglobals
	nullLiteral = []byte("null")

	_ json.Unmarshaler = (*Version)(nil)
	_ json.Unmarshaler = (*NullString)(nil)
)

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *Version) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, nullLiteral) {
		return nil
	}

	t, err := time.Parse(`"`+RFC3339DateFormat+`"`, string(data))
	if err != nil {
		return fmt.Errorf("kenall: failed to parse date with RFC3339 Date: %w", err)
	}

	*v = Version(t)

	return nil
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (ns *NullString) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, nullLiteral) {
		return nil
	}

	if err := json.Unmarshal(data, &ns.String); err != nil {
		return fmt.Errorf("kenall: failed to parse NullString: %w", err)
	}

	ns.Valid = true

	return nil
}
