package utils

import (
	"bufio"
	"compress/gzip"
	"errors"
	"os"
	"strconv"
	"strings"
	"sync"
)

// ErrInvalidLine when csv line is invalid
var ErrInvalidLine = errors.New("Invalid line structure")

// ErrInvalidIPv4 when invalid ip address provided
var ErrInvalidIPv4 = errors.New("Invalid IPv4 address")

type ipRange struct {
	start       uint
	end         uint
	countryCode string
	countryName string
}

var arr []ipRange
var once sync.Once
var loadError error

// Load db-ip.com csv file
// It must be called only once
func Load(filepath string) error {
	once.Do(func() {
		loadError = load(filepath)
	})

	return loadError
}

// GetCountry returns the country which ip blongs to
func GetCountry(ip string) (string, string) {

	ipNumb, err := ipToInt(ip)
	if err != nil {
		return "", ""
	}

	index := binarySearch(arr, ipNumb, 0, len(arr)-1)
	if index == -1 {
		return "", ""
	}

	return arr[index].countryCode, arr[index].countryName
}

func load(filepath string) error {
	arr = make([]ipRange, 0)
	return loadFile(filepath)
}

func loadFile(filepath string) error {
	f, err := os.Open(filepath) // Replace with your file name
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	scanner := bufio.NewScanner(gzr)

	for scanner.Scan() {
		err = addRaw(scanner.Text())
		if err != nil {
			return err
		}
	}

	err = scanner.Err()
	return err
}

// accept input string as follows
// "{ip}","{ip}","{country}"
func addRaw(line string) error {
	//replace all double quotations
	line = strings.Replace(line, "\"", "", -1)

	startIPnumStr, endIPnumStr, countryCode, countryName, err := extract(line)
	if err != nil {
		return err
	}

	startIPnum, err := strconv.Atoi(startIPnumStr)
	if err != nil {
		return err
	}
	endIPnum, err := strconv.Atoi(endIPnumStr)
	if err != nil {
		return err
	}

	arr = append(arr, ipRange{uint(startIPnum), uint(endIPnum), countryCode, countryName})
	ensureSorted(arr)

	return nil
}

func ensureSorted(arr []ipRange) {

	i := len(arr) - 1
	temp := arr[i]
	for {

		if i == 0 || arr[i].start >= arr[i-1].start {
			break
		}

		arr[i] = arr[i-1]
		i--
	}
	arr[i] = temp
}

func ipToInt(ip string) (uint, error) {

	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return 0, ErrInvalidIPv4
	}

	var result uint
	var index uint = 3
	for i := 3; i >= 0; i-- {

		ipNumb, err := strconv.Atoi(parts[index])
		if err != nil {
			return 0, err
		}

		result |= uint(ipNumb) << ((uint(3) - index) * uint(8))
		index--
	}

	return result, nil
}

func extract(line string) (string, string, string, string, error) {
	parts := strings.Split(line, ",")
	if len(parts) != 4 {
		return "", "", "", "", ErrInvalidLine
	}

	return parts[0], parts[1], parts[2], parts[3], nil
}

func binarySearch(arr []ipRange, key uint, start, end int) int {
	for {

		if start > end {
			return -1 //not found
		}

		mid := (start + end) / 2
		if key >= arr[mid].start && key <= arr[mid].end {
			return mid
		}

		if key < arr[mid].start {
			end = mid - 1
		} else if key > arr[mid].end {
			start = mid + 1
		}

	}
}
