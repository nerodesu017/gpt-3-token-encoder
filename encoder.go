package gpt3tokenencoder

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"regexp"
	"strings"

	"github.com/dlclark/regexp2"
)

const joinValue = "_-_"

var encoder map[string]int32
var decoder map[int32]string
var bpe_file string
var pat regexp2.Regexp
var lines []string
var bpe_merges [][]string
var byte_encoder map[int32]string
var byte_decoder map[string]int32
var bpe_ranks map[string]int32
var cache map[string]string

func init() {
	// load encoder
	encoderFile, err := os.Open("encoder.json")
	if err != nil {
		log.Fatal(err)
	}
	defer encoderFile.Close()

	byteValue, _ := ioutil.ReadAll(encoderFile)

	json.Unmarshal(byteValue, &encoder)

	// fill the decoder
	decoder = make(map[int32]string, len(encoder))
	for i := range encoder {
		decoder[encoder[i]] = i
	}

	// load bpe_file
	dat, err := os.ReadFile("vocab.bpe")
	if err != nil {
		log.Fatal(err)
	}

	bpe_file = string(dat)

	// set the regexp
	pat = *regexp2.MustCompile(`'s|'t|'re|'ve|'m|'ll|'d| ?\p{L}+| ?\p{N}+| ?[^\s\p{L}\p{N}]+|\s+(?!\S)|\s+`, 0x0400)

	// lines
	lines = strings.Split(bpe_file, "\n")

	// fill bpe_merges
	bpe_merges_regexp := regexp.MustCompile(`(\s+)`)
	bpe_merges = make([][]string, 0)
	for _, v := range lines[1 : len(lines)-1] {
		values := bpe_merges_regexp.Split(v, -1)

		innerArray := make([]string, 0)
		for _, v := range values {
			if len(strings.Trim(v, "")) > 0 {
				innerArray = append(innerArray, v)
			}
		}
		bpe_merges = append(bpe_merges, innerArray)
	}

	// byte_encoder and byte_decoder
	byte_encoder = bytes_to_unicode()
	byte_decoder = make(map[string]int32, len(byte_encoder))
	for x := range byte_encoder {
		byte_decoder[byte_encoder[x]] = x
	}

	// bpe_ranks
	bpe_ranks = dictZip(bpe_merges, _range(0, int32(len(bpe_merges))))

	// cache
	cache = make(map[string]string)
}

func _range(x, y int32) []int32 {
	arr := make([]int32, y-x+1)
	for i := range arr {
		arr[i] = x + int32(i)
	}
	return arr
}

func ord(char string) int32 {
	return []rune(char)[0]
}

func encodeStr(str string) []rune {
	return []rune(str)
}

func decodeStr(arr []rune) (outString string) {
	for _, v := range arr {
		outString += string(v)
	}
	return
}

func dictZip(x [][]string, y []int32) (result map[string]int32) {
	result = make(map[string]int32)
	for i := range x {
		result[strings.Join(x[i], joinValue)] = y[i]
	}
	return
}

func bytes_to_unicode() map[int32]string {
	bs := _range(ord("!"), ord("~")+1)
	bs = append(bs, _range(ord("¡"), ord("¬")+1)...)
	bs = append(bs, _range(ord("®"), ord("ÿ")+1)...)

	cs := make([]int32, len(bs))
	copy(cs, bs)

	n := int32(0)

	var b int32
	twotoeight := int32(math.Pow(2, 8))
	for b = 0; b < int32(twotoeight); b += 1 {
		var includesB bool
		includesB = false
		for i := range bs {
			if bs[i] == b {
				includesB = true
				break
			}
		}

		if !includesB {
			bs = append(bs, b)
			cs = append(cs, twotoeight+n)
			n += 1
		}
	}

	csString := make([]string, 0)

	for i := range cs {
		csString = append(csString, string(cs[i]))
	}

	result := make(map[int32]string)

	for i := range bs {
		result[bs[i]] = csString[i]
	}

	return result
}

type Pair struct {
	First  byte
	Second byte
}

func get_pairs(word []string) map[string]struct{} {
	pairs := make(map[string]struct{})
	prev_char := word[0]

	for i := 1; i < len(word); i += 1 {
		char := word[i]
		pairs[strings.Join([]string{string(prev_char), string(char)}, joinValue)] = struct{}{}
		prev_char = char
	}

	return pairs
}

func minInt(integers []int32) (min int32) {
	min = math.MaxInt32
	for _, v := range integers {
		if v < min {
			min = v
		}
	}
	return
}

func regexp2FindAllString(re *regexp2.Regexp, s string) []string {
	var matches []string
	m, _ := re.FindStringMatch(s)
	for m != nil {
		matches = append(matches, m.String())
		m, _ = re.FindNextMatch(m)
	}
	return matches
}

func indexOf(mainArr []string, wanted string, startingPoint int) int {
	for i := startingPoint; i < len(mainArr); i += 1 {
		if mainArr[i] == wanted {
			return i
		}
	}

	return -1
}

func bpe(token string) string {
	if val, ok := cache[token]; ok {
		return val
	}

	// word := strings.Split(token, "")
	word := strings.Split(token, "")

	pairs := get_pairs(word)

	keysNum := len(pairs)

	if keysNum == 0 {
		return token
	}

	for {
		minPairs := make(map[int32]string)

		for i := range pairs {
			if rank, ok := bpe_ranks[i]; ok {
				minPairs[rank] = i
			} else {
				// can't use 10e10, too big, idk :shrug:
				minPairs[int32(10e8)] = i
			}
		}

		bigram := minPairs[minInt(func() []int32 {
			var keys []int32
			keys = make([]int32, 0)
			for i := range minPairs {
				keys = append(keys, i)
			}
			return keys
		}())]

		if _, ok := bpe_ranks[bigram]; !ok {
			break
		}

		firstAndSecond := strings.Split(bigram, joinValue)

		first := firstAndSecond[0]
		second := firstAndSecond[1]
		new_word := make([]string, 0)
		i := 0

		for i < len(word) {
			j := indexOf(word, first, i)

			fmt.Println(word)
			// in golang, each char in a string can be of different lengths
			// transform to rune for correct calculation (each char = 1 len)
			// for k := i; k < j; k += 1 {
			// 	if len(word[k]) > 1 {
			// 		j = j - len(word[k]) + 1
			// 	}
			// }

			// j := subArray(word[i:], strings.Split(first, ""))
			if j == -1 {
				new_word = append(new_word, word[i:]...)
				// new_word = strings.Join([]string{new_word, word[i:]}, "")
				break
			}
			if i < j {
				new_word = append(new_word, word[i:j]...)
			}
			// new_word = strings.Join([]string{new_word, word[i:j]}, "")
			i = j
			if word[i] == first && i < len(word)-1 && word[i+1] == second {
				new_word = append(new_word, strings.Join(firstAndSecond, ""))
				i = i + 2
			} else {
				new_word = append(new_word, word[i])
				i = i + 1
			}
		}

		word = new_word
		if len(word) == 1 {
			break
		} else {
			pairs = get_pairs(word)
		}
	}

	cache[token] = strings.Join(word, " ")
	return cache[token]
}

func Encode(text string) []int32 {
	bpe_tokens := make([]int32, 0)
	matches := regexp2FindAllString(&pat, text)
	for _, token := range matches {
		token = func() string {
			encodedStr := encodeStr(token)
			arr := make([]string, 0)
			for _, v := range encodedStr {
				arr = append(arr, byte_encoder[int32(v)])
			}
			return strings.Join(arr, "")
		}()

		new_tokens := func() []int32 {
			arr := make([]int32, 0)

			for _, v := range strings.Split(bpe(token), " ") {
				fmt.Printf("%s", v)
				arr = append(arr, encoder[v])
			}
			fmt.Printf("\n")
			return arr
		}()

		bpe_tokens = append(bpe_tokens, new_tokens...)
	}

	return bpe_tokens
}

func Decode(tokens []int32) string {
	text := func() string {
		arr := make([]string, 0)
		for _, v := range tokens {
			arr = append(arr, decoder[v])
		}
		return strings.Join(arr, "")
	}()

	text = decodeStr(func() []rune {
		arr := strings.Split(text, "")
		new_arr := make([]int32, 0)
		for _, v := range arr {
			new_arr = append(new_arr, byte_decoder[v])
		}
		return new_arr
	}())

	return text
}
