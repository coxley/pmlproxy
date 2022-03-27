package server

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	decodeDuration = prometheus.NewSummary(prometheus.SummaryOpts{
		Namespace:  "PlantUML",
		Name:       "decode_seconds",
		MaxAge:     time.Minute * 10,
		Help:       "duration to decode short-form to digram",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	})
	encodeDuration = prometheus.NewSummary(prometheus.SummaryOpts{
		Namespace:  "PlantUML",
		Name:       "encode_seconds",
		MaxAge:     time.Minute * 10,
		Help:       "duration to encode diagram to short-form",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	})
)

// PlantUML uses it's own charater set over the standard b64 one
var charset string = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_"
var enc = base64.NewEncoding(charset)

func init() {
	prometheus.MustRegister(encodeDuration)
	prometheus.MustRegister(decodeDuration)
}

// ToShort converts diagram syntax to an encoded string
//
// The format is described here: lhttps://plantuml.com/text-encoding
func ToShort(source string) (string, error) {
	timer := prometheus.NewTimer(encodeDuration)
	defer timer.ObserveDuration()

	short, err := p64EncodeToString([]byte(source))
	if err != nil {
		return "", fmt.Errorf("failed to compress diagram: %w", err)
	}
	return short, nil
}

// FromShort converts an encoded string to the original diagram source
//
// The format is described here: lhttps://plantuml.com/text-encoding
func FromShort(short string) (string, error) {
	timer := prometheus.NewTimer(decodeDuration)
	defer timer.ObserveDuration()

	b, err := p64DecodeFromString(short)
	if err != nil {
		return "", fmt.Errorf("failed to decompress diagram: %w", err)
	}
	return string(b), nil
}

func p64EncodeToString(data []byte) (string, error) {
	var b bytes.Buffer
	w, err := flate.NewWriter(&b, -1)
	if err != nil {
		return "", err
	}
	w.Write(data)
	w.Close()
	return enc.EncodeToString(b.Bytes()), nil
}

func p64DecodeFromString(s string) ([]byte, error) {
	inflated, err := enc.DecodeString(s)
	if err != nil {
		return nil, err
	}

	var b bytes.Buffer
	r := flate.NewReader(bytes.NewReader(inflated))
	b.ReadFrom(r)
	r.Close()

	return b.Bytes(), nil
}
