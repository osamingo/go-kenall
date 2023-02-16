package kenall

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
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

	holiday struct {
		Title         string `json:"title"`
		Date          string `json:"date"`
		DayOfWeek     int    `json:"day_of_week"`
		DayOfWeekText string `json:"day_of_week_text"`
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
	// A Corporation is a corporation associated with the corporate number defined by National Tax Agency Japan.
	Corporation struct {
		PublishedDate            string      `json:"published_date"`
		SequenceNumber           json.Number `json:"sequence_number"`
		CorporateNumber          string      `json:"corporate_number"`
		Process                  json.Number `json:"process"`
		Correct                  json.Number `json:"correct"`
		UpdateDate               string      `json:"update_date"`
		ChangeDate               string      `json:"change_date"`
		Name                     string      `json:"name"`
		NameImageID              NullString  `json:"name_image_id"`
		Kind                     string      `json:"kind"`
		PrefectureName           string      `json:"prefecture_name"`
		CityName                 string      `json:"city_name"`
		StreetNumber             string      `json:"street_number"`
		Town                     NullString  `json:"town"`
		KyotoStreet              NullString  `json:"kyoto_street"`
		BlockLotNum              NullString  `json:"block_lot_num"`
		Building                 NullString  `json:"building"`
		FloorRoom                NullString  `json:"floor_room"`
		AddressImageID           NullString  `json:"address_image_id"`
		JISX0402                 string      `json:"jisx0402"`
		PostCode                 string      `json:"post_code"`
		AddressOutside           string      `json:"address_outside"`
		AddressOutsideImageID    NullString  `json:"address_outside_image_id"`
		CloseDate                NullString  `json:"close_date"`
		CloseCause               NullString  `json:"close_cause"`
		SuccessorCorporateNumber NullString  `json:"successor_corporate_number"`
		ChangeCause              string      `json:"change_cause"`
		AssignmentDate           string      `json:"assignment_date"`
		EnName                   string      `json:"en_name"`
		EnPrefectureName         string      `json:"en_prefecture_name"`
		EnAddressLine            NullString  `json:"en_address_line"`
		EnAddressOutside         NullString  `json:"en_address_outside"`
		Furigana                 string      `json:"furigana"`
		Hihyoji                  string      `json:"hihyoji"`
	}
	// A RemoteAddress is an IP address from access point.
	RemoteAddress struct {
		Type    string      `json:"type"`
		Address string      `json:"address"`
		IPAddr  *net.IPAddr `json:"-"`
	}
	// A Holiday is Japan's holiday detail.
	Holiday struct {
		Title string `json:"title,omitempty"`
		time.Time
	}
	// A Query is data normalized to an address.
	Query struct {
		Q           NullString `json:"q"`
		T           NullString `json:"t"`
		Prefecture  NullString `json:"prefecture"`
		County      NullString `json:"county"`
		City        NullString `json:"city"`
		CityWard    NullString `json:"city_ward"`
		Town        NullString `json:"town"`
		KyotoStreet NullString `json:"kyoto_street"`
		BlockLotNum NullString `json:"block_lot_num"`
		Building    NullString `json:"building"`
		FloorRoom   NullString `json:"floor_room"`
	}
)

var (
	//nolint: gochecknoglobals
	nullLiteral = []byte("null")
	//nolint: gochecknoglobals, gomnd
	jst = time.FixedZone("Asia/Tokyo", 9*60*60)

	_ json.Unmarshaler = (*Version)(nil)
	_ json.Unmarshaler = (*NullString)(nil)
	_ json.Unmarshaler = (*RemoteAddress)(nil)
	_ json.Unmarshaler = (*Holiday)(nil)
	_ json.Unmarshaler = (*BusinessDay)(nil)

	_ json.Marshaler = (*Holiday)(nil)
	_ json.Marshaler = (*BusinessDay)(nil)

	_ net.Addr = (*RemoteAddress)(nil)
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

// UnmarshalJSON implements json.Unmarshaler interface.
func (ra *RemoteAddress) UnmarshalJSON(data []byte) error {
	type Alias RemoteAddress

	tmp := &struct{ *Alias }{Alias: (*Alias)(ra)}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return fmt.Errorf("kenall: failed to parse RemoteAddress: %w", err)
	}

	switch tmp.Type {
	case "v4":
		var err error
		if tmp.IPAddr, err = net.ResolveIPAddr("ip4", tmp.Address); err != nil {
			return fmt.Errorf("kenall: failde to resolve IP address: %w", err)
		}
	case "v6":
		var err error
		if tmp.IPAddr, err = net.ResolveIPAddr("ip6", tmp.Address); err != nil {
			return fmt.Errorf("kenall: failed to resolve IP address: %w", err)
		}
	default:
		//nolint: goerr113
		return errors.New("kenall: undefined type of RemoteAddress, type = " + tmp.Type)
	}

	return nil
}

// Network implements net.Addr interface.
func (ra *RemoteAddress) Network() string {
	return ra.IPAddr.Network()
}

// RemoteAddress implements net.Addr and fmt.Stringer interface.
func (ra *RemoteAddress) String() string {
	return ra.IPAddr.String()
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (h *Holiday) UnmarshalJSON(data []byte) error {
	var tmp holiday
	if err := json.Unmarshal(data, &tmp); err != nil {
		return fmt.Errorf("kenall: failed to parse Holiday: %w", err)
	}

	var err error
	if h.Time, err = time.ParseInLocation(RFC3339DateFormat, tmp.Date, jst); err != nil {
		return fmt.Errorf("kenall: failed to parse Holiday: %w", err)
	}

	h.Title = tmp.Title

	return nil
}

// MarshalJSON implements json.Marshaler interface.
func (h Holiday) MarshalJSON() ([]byte, error) {
	//nolint: wrapcheck
	return json.Marshal(&holiday{
		Title:         h.Title,
		Date:          h.Format(RFC3339DateFormat),
		DayOfWeek:     int(h.Weekday()),
		DayOfWeekText: strings.ToLower(h.Weekday().String()),
	})
}
