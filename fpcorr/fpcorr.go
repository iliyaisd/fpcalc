package fpcorr

import (
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const (
	// seconds to sample audio file for
	sampleTime = 500
	// number of points to scan cross correlation over
	span = 150
	// step size (in points) of cross correlation
	step = 1
	// minimum number of points that must overlap in cross correlation
	// exception is raised if this cannot be met
	minOverlap = 20
	// report match when cross correlation has a peak exceeding threshold
	threshold = 0.5
)

var removeNonDigits = regexp.MustCompile("[^0-9]+")

func calculateFingerprints(filename string) ([]int, error) {
	out, err := exec.Command("fpcalc", "-raw", "-length", strconv.Itoa(sampleTime), filename).Output()
	if err != nil {
		return nil, err
	}

	fpcalcOut := string(out)
	fingerprintIndex := strings.Index(fpcalcOut, "FINGERPRINT=") + 12

	var fingerprints []int
	for _, s := range strings.Split(fpcalcOut[fingerprintIndex:], ",") {
		s = removeNonDigits.ReplaceAllString(s, "")
		i, err := strconv.Atoi(s)
		if err != nil {
			return nil, err
		}
		fingerprints = append(fingerprints, i)
	}
	return fingerprints, nil
}

func correlation(listx, listy []int) float64 {
	if len(listx) == 0 || len(listy) == 0 {
		// Error checking in main program should prevent us from ever being
		// able to get here.
		panic("Empty lists cannot be correlated.")
	}
	if len(listx) > len(listy) {
		listx = listx[:len(listy)]
	} else if len(listx) < len(listy) {
		listy = listy[:len(listx)]
	}

	covariance := 0
	for i := 0; i < len(listx); i++ {
		covariance += 32 - countOnes(listx[i]^listy[i])
	}

	return (float64(covariance) / float64(len(listx))) / float64(32)
}

func countOnes(num int) int {
	count := 0
	for num > 0 {
		if num&1 == 1 {
			count++
		}
		num >>= 1
	}
	return count
}

func crossCorrelation(listx, listy []int, offset int) float64 {
	if offset > 0 {
		listx = listx[offset:]
		listy = listy[:len(listx)]
	} else if offset < 0 {
		offset = -offset
		listy = listy[offset:]
		listx = listx[:len(listy)]
	}
	if min(len(listx), len(listy)) < minOverlap {
		return 0
	}
	return correlation(listx, listy)
}

func compare(listx, listy []int, span, step int) []float64 {
	if span > min(len(listx), len(listy)) {
		return nil
	}
	var corrXy []float64
	for offset := -span; offset < span+1; offset += step {
		corrXy = append(corrXy, crossCorrelation(listx, listy, offset))
	}
	return corrXy
}

func maxIndex(listx []float64) int {
	curMaxIndex := 0
	maxValue := listx[0]
	for i, value := range listx {
		if value > maxValue {
			maxValue = value
			curMaxIndex = i
		}
	}
	return curMaxIndex
}

func getMaxCorr(corr []float64, source, target string) float64 {
	maxCorrIndex := maxIndex(corr)
	maxCorrOffset := -span + maxCorrIndex*step
	log.Printf("max_corr_index = %d, max_corr_offset = %d", maxCorrIndex, maxCorrOffset)
	if corr[maxCorrIndex] > threshold {
		log.Printf("%s and %s match with correlation of %.4f at offset %d", source, target, corr[maxCorrIndex], maxCorrOffset)
	}
	return corr[maxCorrIndex]
}

func min(a int, b int) int {
	if a > b {
		return b
	}
	return a
}

func AudioCorrelate(source, target string) (float64, error) {
	fingerprintSource, err := calculateFingerprints(source)
	if err != nil {
		return 0, err
	}

	fingerprintTarget, err := calculateFingerprints(target)
	if err != nil {
		return 0, err
	}

	corr := compare(fingerprintSource, fingerprintTarget, span, step)
	maxCorr := getMaxCorr(corr, source, target)

	return maxCorr, nil
}
