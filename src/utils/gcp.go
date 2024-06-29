package utils

import (
	"io"
	"net/http"
	"strings"
)

const (
	unknownRegion = "Unknown Server Region"
)

var regionCodeToName = map[string]string{
	"us-central1":             "Iowa, US",
	"us-west1":                "Oregon, US",
	"us-west2":                "Los Angeles, US",
	"us-west3":                "Salt Lake City, US",
	"us-west4":                "Las Vegas, US",
	"us-east1":                "South Carolina, US",
	"us-east4":                "Northern Virginia, US",
	"northamerica-northeast1": "Montreal, Canada",
	"southamerica-east1":      "São Paulo, Brazil",
	"europe-north1":           "Finland",
	"europe-west1":            "Belgium",
	"europe-west2":            "London, UK",
	"europe-west3":            "Frankfurt, Germany",
	"europe-west4":            "Netherlands",
	"europe-west6":            "Zurich, Switzerland",
	"asia-east1":              "Taiwan",
	"asia-east2":              "Hong Kong",
	"asia-northeast1":         "Tokyo, Japan",
	"asia-northeast2":         "Osaka, Japan",
	"asia-northeast3":         "Seoul, South Korea",
	"asia-south1":             "Mumbai, India",
	"asia-southeast1":         "Singapore",
	"asia-southeast2":         "Jakarta, Indonesia",
	"australia-southeast1":    "Sydney, Australia",
	"australia-southeast2":    "Melbourne, Australia",
	"southasia-east1":         "Delhi, India",
	"us-east5":                "Columbus, US",
}

func GetServerRegion() string {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://metadata.google.internal/computeMetadata/v1/instance/zone", nil)
	if err != nil {
		return unknownRegion
	}
	req.Header.Set("Metadata-Flavor", "Google")
	resp, err := client.Do(req)
	if err != nil {
		return unknownRegion
	}

	if resp.StatusCode != 200 {
		return unknownRegion
	}

	msgBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return unknownRegion
	}

	msg := string(msgBytes)
	msgSplitted := strings.Split(msg, "/")
	if len(msgSplitted) < 4 {
		return unknownRegion
	}

	zone := msgSplitted[3]
	if len(zone) < 2 {
		return unknownRegion
	}

	region, ok := regionCodeToName[zone[:len(zone)-2]]
	if !ok {
		return unknownRegion
	} else {
		return region
	}
}
