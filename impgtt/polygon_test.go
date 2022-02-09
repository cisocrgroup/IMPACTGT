package main

import (
	"fmt"
	"image"
	"testing"
)

func TestNewPolygon(t *testing.T) {
	tests := []struct {
		test, want string
		err        bool
	}{
		{"1,2 3,4 5,6", "(1,2)-(3,4)-(5,6)", false},
		{"1,2 3,4 56", "", true},
		{"1,2 3,4 a,b", "", true},
		{"1,2 3,4", "(1,2)-(3,2)-(3,4)-(1,4)", false},
		{"1,2", "", true},
		{"", "", true},
	}
	for _, tc := range tests {
		t.Run(tc.test, func(t *testing.T) {
			p, err := makePolygon(tc.test)
			if tc.err && err == nil {
				t.Fatalf("expected an error")
			}
			if !tc.err && err != nil {
				t.Fatalf("got error: %v", err)
			}
			if got := p.String(); got != tc.want {
				t.Fatalf("expected %s; got %s", tc.want, got)
			}
		})
	}
}

func TestPolygonBoundingRectangle(t *testing.T) {
	tests := []struct {
		test, want string
	}{
		{"1,2 3,4 5,6", "(1,2)-(5,6)"},
		{"1,2 5,6 3,4", "(1,2)-(5,6)"},
		{"5,6 3,4 1,2", "(1,2)-(5,6)"},
		{"5,6 1,2 3,4", "(1,2)-(5,6)"},
	}
	for _, tc := range tests {
		t.Run(tc.test, func(t *testing.T) {
			p, err := makePolygon(tc.test)
			if err != nil {
				t.Fatalf("got error: %v", err)
			}
			if got := fmt.Sprintf("%v", p.boundingRectangle()); got != tc.want {
				t.Fatalf("expected %v; got %v", tc.want, got)
			}
		})
	}
}

func TestPolygonInside(t *testing.T) {
	tests := []struct {
		test  string
		point image.Point
		want  bool
	}{
		{"1,3 2,1 3,3", image.Pt(2, 2), true},
		{"1,3 2,1 3,3", image.Pt(1, 2), false},
		{"1,3 2,1 3,3", image.Pt(1, 1), false},
	}
	for _, tc := range tests {
		t.Run(tc.test, func(t *testing.T) {
			p, err := makePolygon(tc.test)
			if err != nil {
				t.Fatalf("got error: %v", err)
			}
			if got := p.inside(tc.point); got != tc.want {
				t.Fatalf("expected %t; got %t", tc.want, got)
			}
		})
	}
}
