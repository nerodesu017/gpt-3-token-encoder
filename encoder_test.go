package gpt3tokenencoder

import (
	"testing"
)

func compareArrays(arr1 []int32, arr2 []int32) bool {
	if len(arr1) != len(arr2) {
		return false
	}

	for i := range arr1 {
		if arr1[i] != arr2[i] {
			return false
		}
	}
	return true
}

func TestEncodeEmptyString(t *testing.T) {
	str := ""
	gotEncoded := Encode(str)
	if len(gotEncoded) != 0 {
		t.Errorf(`Encode(%s) = %v wanted, got %v`, str, gotEncoded, []int32{})
	}
	gotDecoded := Decode(Encode(str))
	if gotDecoded != str {
		t.Errorf(`Decode(Encode(%s)) = %s`, str, gotDecoded)
	}
}

func TestEncodeSpace(t *testing.T) {
	str := " "
	gotEncoded := Encode(str)
	if !compareArrays(gotEncoded, []int32{220}) {
		t.Errorf(`Encode(%s) = %v`, str, gotEncoded)
	}
	gotDecoded := Decode(Encode(str))
	if gotDecoded != str {
		t.Errorf(`Decode(Encode(%s)) = %s`, str, gotDecoded)
	}
}

func TestEncodeTab(t *testing.T) {
	str := "\t"
	gotEncoded := Encode(str)
	if !compareArrays(gotEncoded, []int32{197}) {
		t.Errorf(`Encode(%s) = %v`, str, gotEncoded)
	}
	gotDecoded := Decode(Encode(str))
	if gotDecoded != str {
		t.Errorf(`Decode(Encode(%s)) = %s`, str, gotDecoded)
	}
}

func TestEncodeSimpleText(t *testing.T) {
	str := "This is some text"
	gotEncoded := Encode(str)
	if !compareArrays(gotEncoded, []int32{1212, 318, 617, 2420}) {
		t.Errorf(`Encode(%s) = %v`, str, gotEncoded)
	}
	gotDecoded := Decode(Encode(str))
	if gotDecoded != str {
		t.Errorf(`Decode(Encode(%s)) = %s`, str, gotDecoded)
	}

}

func TestEncodeMultiTokenWord(t *testing.T) {
	str := "indivisible"
	gotEncoded := Encode(str)
	if !compareArrays(gotEncoded, []int32{521, 452, 12843}) {
		t.Errorf(`Encode(%s) = %v`, str, gotEncoded)
	}
	gotDecoded := Decode(Encode(str))
	if gotDecoded != str {
		t.Errorf(`Decode(Encode(%s)) = %s`, str, gotDecoded)
	}
}

func TestEncodeEmojis(t *testing.T) {
	str := "hello 👋 world 🌍"
	gotEncoded := Encode(str)
	if !compareArrays(gotEncoded, []int32{31373, 50169, 233, 995, 12520, 234, 235}) {
		t.Errorf(`Encode(%s) = %v`, str, gotEncoded)
	}
	gotDecoded := Decode(Encode(str))
	if gotDecoded != str {
		t.Errorf(`Decode(Encode(%s)) = %s`, str, gotDecoded)
	}
}
