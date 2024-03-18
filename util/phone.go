package util

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func NormalizeMSISDN(msisdn string) (string, error) {
	msisdn = strings.ReplaceAll(msisdn, " ", "")
	msisdn = strings.ReplaceAll(msisdn, "-", "")

	if _, err := strconv.ParseInt(msisdn, 10, 64); err != nil {
		return "", errors.New("invalid phone number")
	}

	var (
		firstChar, firstThreeChar string
	)

	if len(msisdn) > 0 {
		firstChar = msisdn[0:1]
	}
	if len(msisdn) > 3 {
		firstThreeChar = msisdn[:3]
	}

	if firstChar == "0" {
		msisdn = fmt.Sprintf("62%s", msisdn[1:])
	}
	if firstChar == "8" {
		msisdn = fmt.Sprintf("62%s", msisdn)
	}
	if firstThreeChar == "+62" {
		msisdn = msisdn[1:]
	}

	for strings.HasPrefix(msisdn, "6262") {
		msisdn = strings.Replace(msisdn, "6262", "62", 1)
	}

	return msisdn, nil
}

type MSISDNFormat string

const (
	MSISDN08     MSISDNFormat = "08"
	MSISDN8      MSISDNFormat = "8"
	MSISDN62Plus MSISDNFormat = "62"
)

func DenormalizeMSISDN(msisdn string, format MSISDNFormat) (string, error) {
	msisdn, err := NormalizeMSISDN(msisdn)
	if err != nil {
		return "", err
	}

	switch format {
	case MSISDN08:
		return fmt.Sprintf("0%s", msisdn[2:]), nil
	case MSISDN8:
		return msisdn[2:], nil
	case MSISDN62Plus:
		return fmt.Sprintf("+%s", msisdn), nil
	default:
		return msisdn, nil
	}
}
