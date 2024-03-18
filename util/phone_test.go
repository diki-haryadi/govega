package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeMSISDN(t *testing.T) {
	t.Parallel()

	var (
		normalizedMSISDN = "6285641234567"
	)

	tests := []struct {
		msisdn     string
		normalized string
	}{
		{
			msisdn:     "085641234567",
			normalized: normalizedMSISDN,
		},
		{
			msisdn:     "0856 4123 4567",
			normalized: normalizedMSISDN,
		},
		{
			msisdn:     "085-641-234-567",
			normalized: normalizedMSISDN,
		},
		{
			msisdn:     "85641234567",
			normalized: normalizedMSISDN,
		},
		{
			msisdn:     "85 641 234 567",
			normalized: normalizedMSISDN,
		},
		{
			msisdn:     "85-641-234-567",
			normalized: normalizedMSISDN,
		},
		{
			msisdn:     normalizedMSISDN,
			normalized: normalizedMSISDN,
		},
		{
			msisdn:     "62856 4123 4567",
			normalized: normalizedMSISDN,
		},
		{
			msisdn:     "6285-641-234-567",
			normalized: normalizedMSISDN,
		},
		{
			msisdn:     "+6285641234567",
			normalized: normalizedMSISDN,
		},
		{
			msisdn:     "+62 85641234567",
			normalized: normalizedMSISDN,
		},
		{
			msisdn:     "+62856-4123-4567",
			normalized: normalizedMSISDN,
		},
		{
			msisdn:     "6285641234567a",
			normalized: "",
		},
		{
			msisdn:     "",
			normalized: "",
		},
		{
			msisdn:     "626285641234567",
			normalized: normalizedMSISDN,
		},
		{
			msisdn:     "62626285641234567",
			normalized: normalizedMSISDN,
		},
		{
			msisdn:     "6262626285641234567",
			normalized: normalizedMSISDN,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.msisdn, func(t *testing.T) {
			t.Parallel()
			res, _ := NormalizeMSISDN(tt.msisdn)
			assert.Equal(t, tt.normalized, res)
		})
	}
}

func TestDenormalizeMSISDN(t *testing.T) {
	t.Parallel()

	var (
		normalizedMSISDN = "6285641234567"

		denormalizedMSISDNs = map[MSISDNFormat]string{
			MSISDN8:      "85641234567",
			MSISDN08:     "085641234567",
			MSISDN62Plus: "+6285641234567",
		}
	)

	tests := []struct {
		msisdn string
		error  bool
	}{
		{
			msisdn: "085641234567",
		},
		{
			msisdn: "0856 4123 4567",
		},
		{
			msisdn: "085-641-234-567",
		},
		{
			msisdn: "85641234567",
		},
		{
			msisdn: "85 641 234 567",
		},
		{
			msisdn: "85-641-234-567",
		},
		{
			msisdn: normalizedMSISDN,
		},
		{
			msisdn: "62856 4123 4567",
		},
		{
			msisdn: "6285-641-234-567",
		},
		{
			msisdn: "+6285641234567",
		},
		{
			msisdn: "+62 85641234567",
		},
		{
			msisdn: "+62856-4123-4567",
		},
		{
			msisdn: "626285641234567",
		},
		{
			msisdn: "62626285641234567",
		},
		{
			msisdn: "6262626285641234567",
		},
		{
			msisdn: "6285641234567a",
			error:  true,
		},
		{
			msisdn: "",
			error:  true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.msisdn, func(t *testing.T) {
			t.Parallel()
			for msisdnFormat, denormalized := range denormalizedMSISDNs {
				res, err := DenormalizeMSISDN(tt.msisdn, msisdnFormat)
				if err != nil {
					if !tt.error {
						t.Errorf("unexpected error: %s", err)
					}
					continue
				}
				if tt.error {
					t.Errorf("expected error not happened")
					continue
				}
				assert.Equal(t, denormalized, res)
			}
		})
	}
}
