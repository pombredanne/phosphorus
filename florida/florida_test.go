package florida

import (
	"os"
	"testing"
	"encoding/json"
)

var sampleRecords []string

func init() {
	file, _ := os.Open("sample_records.json")
	decoder := json.NewDecoder(file)
	decoder.Decode(&sampleRecords)
}

var RECORD_DATA = "OKA\t115753900\tMeehan\t\tMary\tJ\tN\t6164   Holloway Rd \t \tBaker\t\t325318132\t\t\t\t\t\t\t\tF\t5\t01/11/1942\t01/11/2008\tDEM\t01\t0\t01.1\t\tACT\t1\t3\t2\t3\t3\t555\t8675309\t\n"

func TestSingleRecord(t *testing.T) {
	record, _ := ParseRecord(RECORD_DATA)

	if record.lastName != "Meehan" ||
		record.firstName != "Mary" ||
		record.middleName != "J" {
		t.Errorf("Error reading name.")
	}

	if record.birthMonth != "January" ||
		record.birthDay != 11 ||
		record.birthYear != 1942 {
		t.Errorf("Error parsing birthday: %q", record)
	}

	if record.gender != Female {
		t.Errorf("Error parsing gender.")
	}

	if record.race != WhiteNotHispanic {
		t.Errorf("Error parsing race: %d != ", record.race, WhiteNotHispanic)
	}

	if record.party != Democratic {
		t.Errorf("Error parsing party: %d != ", record.party, Democratic)
	}

	if record.telephone.areaCode != [3]byte{'5','5','5'} ||
		record.telephone.number != [7]byte{'8','6','7','5','3','0','9'} {
		t.Errorf("Error parsing phone: %d", record.telephone.areaCode)
	}
}

/*
If you don't copy the struct members into a buffer before passing them
to Metaphone, it will read past the end of the field. (SWIG's fault?
Works as designed?) Gross.
*/
func TestMetaphoneStringBounds(t *testing.T) {
	record, _ := ParseRecord(RECORD_DATA)
	f := record.Fields()
	if f[0] != "MHN" || f[1] != "MR" {
		t.Errorf("Incorrect Metaphone encoding: %s", f)
	}
}

func TestMultipleRecords(t *testing.T) {
	for _, line := range sampleRecords {
		_, err := ParseRecord(line)

		if err != nil {
			t.Errorf("Failed parsing record: %q", line)
		}
	}
}

var goodNumbers = []string{
	"   +1 312 867 5309",
	"3128675309",
	"(312) 867-5309",
	"31 28 67 53 09",
	"312/867.5309",
	"1 312 867 5309"}

var correctArea = [3]byte{'3','1','2'}
var correctNumber = [7]byte{'8','6','7','5','3','0','9'}

func TestParsePhoneGood(t *testing.T) {
	for _, num := range goodNumbers {
		p, err := parsePhone(num)
		if err != nil {
			t.Errorf("%q: %s", num, err.Error())
		} else if p.areaCode != correctArea {
			t.Errorf("%q: %q != %q", num, p.areaCode, correctArea)
		} else if p.number != correctNumber {
			t.Errorf("%q: %q != %q", num, p.number, correctNumber)
		}
	}
}

var badNumbers = []string{
	"1-800-ABCDEFG",
	"312X8675309",
	"911",
	"8675309",
	"+++13128675309",
	"+3128675309",
	")312) 867-5309",
	"312-86-75-309",
	"(312)) 867-5309",
	"+1113128675309",
	"+93128675309",
	"+044 020 5555 5555"}

func TestParsePhoneBad(t *testing.T) {
	for _, num := range badNumbers {
		_, err := parsePhone(num)
		if err == nil {
			t.Errorf("Parsed bad number: %q", num)
		}
	}
}
