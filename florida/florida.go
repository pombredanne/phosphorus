package florida

import (
	"bytes"
	"fmt"
	// "runtime/pprof"
	"encoding/gob"
	"strconv"
	"strings"
	"time"
	"unicode"
	"willstclair.com/phosphorus/metaphone3"
)

type Gender int

const (
	Male Gender = iota
	Female
	GenderUnknown
)

var GenderKey = map[string]Gender{
	"M": Male,
	"F": Female,
	"U": GenderUnknown,
}

type Party int

const (
	Democratic Party = iota
	Republican
)

var PartyKey = map[string]Party{
	"DEM": Democratic,
	"REP": Republican,
}

type Race int

const (
	AmericanIndianOrAlaskanNative Race = 1 + iota
	AsianOrPacificIslander
	BlackNotHispanic
	Hispanic
	WhiteNotHispanic
	Other
	MultiRacial
	RaceUnknown
)

var RaceKey = map[int]Race{
	1: AmericanIndianOrAlaskanNative,
	2: AsianOrPacificIslander,
	3: BlackNotHispanic,
	4: Hispanic,
	5: WhiteNotHispanic,
	6: Other,
	7: MultiRacial,
	8: RaceUnknown,
}

type NANPNumber struct {
	areaCode [3]byte
	number   [7]byte
}

type VoterRecord struct {
	lastName   string
	firstName  string
	middleName string
	birthMonth string
	birthDay   int
	birthYear  int
	city       string
	gender     Gender
	party      Party
	race       Race
	telephone  NANPNumber
}

var mp metaphone3.Metaphone3

func init() {
	mp = metaphone3.NewMetaphone3()
	// mp.SetKeyLength(5)

	gob.RegisterName("Gender", Male)
	gob.RegisterName("Party", Democratic)
	gob.RegisterName("Race", AmericanIndianOrAlaskanNative)
	// gob.RegisterName("NANPNumber", NANPNumber{[3]byte{'3','1','2'},[7]byte{'8','6','7','5','3','0','9'}})
}

func (r VoterRecord) Fields() []interface{} {
	var phFirst string
	buf := new(bytes.Buffer)

	if len(r.firstName) > 0 {
		buf.WriteString(r.firstName)
		mp.SetWord(buf.String())
		buf.Reset()
		mp.Encode()
		phFirst = mp.GetMetaph()
	}

	var phLast string
	if len(r.lastName) > 0 {
		buf.WriteString(r.lastName)
		mp.SetWord(buf.String())
		buf.Reset()
		mp.Encode()
		phLast = mp.GetMetaph()
	}

	// TODO: missing fields should not be weighted at all.
	var mInitial string
	if len(r.middleName) > 0 {
		mInitial = string(r.middleName[0])
	}

	return []interface{}{
		phLast, phFirst, mInitial, r.birthMonth,
		r.birthDay, r.birthYear, r.city, r.gender, r.party,
		r.race} //, r.telephone.areaCode, r.telephone.number}
}

func ParseRecord(data string) (*VoterRecord, error) {
	parsedRecord := &VoterRecord{}

	fields := strings.Split(data, "\t")

	birthdate, err := time.Parse("01/02/2006", fields[21])
	if err != nil {
		return parsedRecord, err
	}

	race_id, err := strconv.ParseInt(fields[20], 0, 32)
	if err != nil {
		race_id = 8
	}

	// gross
	parsedRecord.lastName = fields[2]
	parsedRecord.firstName = fields[4]
	parsedRecord.middleName = fields[5]
	parsedRecord.birthMonth = birthdate.Month().String()
	parsedRecord.birthDay = birthdate.Day()
	parsedRecord.birthYear = birthdate.Year()
	parsedRecord.city = fields[9]
	parsedRecord.party = PartyKey[fields[23]]
	parsedRecord.gender = GenderKey[fields[19]]
	parsedRecord.race = Race(int(race_id))

	phone, err := parsePhone(fields[34] + fields[35])
	if err == nil {
		parsedRecord.telephone = phone
	}

	return parsedRecord, nil
}

func parsePhone(rawNumber string) (NANPNumber, error) {
	type parserState int
	const (
		Start = iota
		ExpectCountryCode
		ExpectArea
		BeginArea
		Area
		ExpectNumber
		BeginNumber
		Number
		SeenMidNumberSymbol
		End
	)

	var parsedNumber NANPNumber
	var state parserState = Start
	var i = 0
	var err error = nil

loop:
	for _, r := range rawNumber {
		if unicode.IsSpace(r) {
			continue
		}

		switch state {
		case ExpectCountryCode:
			if r == '1' {
				state = ExpectArea
				continue
			}
			err = fmt.Errorf("expected '1' to follow prefix")
			break loop
		case Start:
			if r == '1' {
				state = BeginArea
				continue
			} else if r == '+' {
				state = ExpectCountryCode
				continue
			}
			fallthrough
		case ExpectArea:
			if r == '(' {
				state = BeginArea
				continue
			}
			fallthrough
		case BeginArea:
			if 50 <= r && r <= 57 {
				parsedNumber.areaCode[i] = byte(r)
				i++
				state = Area
				continue
			}
			err = fmt.Errorf("expected 2-9, got %q", r)
			break loop
		case Area:
			if 48 <= r && r <= 57 {
				parsedNumber.areaCode[i] = byte(r)
				if i == 2 {
					i = 0
					state = ExpectNumber
				} else {
					i++
				}
				continue
			}
			err = fmt.Errorf("expected digit, got %q", r)
			break loop
		case ExpectNumber:
			if (r == ')') || (r == '/') || (r == '-') || (r == '.') {
				state = BeginNumber
				continue
			}
			fallthrough
		case BeginNumber:
			if 50 <= r && r <= 57 {
				parsedNumber.number[i] = byte(r)
				i++
				state = Number
				continue
			}
			err = fmt.Errorf("expected 2-9, got %q", r)
			break loop
		case Number:
			if (r == '.') || (r == '-') {
				state = SeenMidNumberSymbol
				continue
			}
			fallthrough
		case SeenMidNumberSymbol:
			if 48 <= r && r <= 57 {
				parsedNumber.number[i] = byte(r)
				if i == 6 {
					goto success
				} else {
					i++
				}
				continue
			}
			err = fmt.Errorf("expected digit, got %q", r)
			break loop
		}
	}

	if err == nil {
		err = fmt.Errorf("parse ended prematurely")
	}

success:
	return parsedNumber, err
}
