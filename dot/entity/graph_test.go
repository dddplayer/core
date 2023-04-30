package entity

import (
	"reflect"
	"testing"
)

func TestPortStr(t *testing.T) {
	s := "192.168.0.1:8080"
	expected := "192_168_0_1:8080"
	if res := portStr(s); res != expected {
		t.Errorf("portStr(%q) = %q, expected %q", s, res, expected)
	}

	s = "10.0.0.1-8080"
	expected = "10_0_0_1_8080"
	if res := portStr(s); res != expected {
		t.Errorf("portStr(%q) = %q, expected %q", s, res, expected)
	}

	s = "foo.bar_baz"
	expected = "foo_bar_baz"
	if res := portStr(s); res != expected {
		t.Errorf("portStr(%q) = %q, expected %q", s, res, expected)
	}
}

func TestBlankColumn(t *testing.T) {
	expected := &Data{
		Text:    "",
		Port:    "",
		BgColor: "white",
		RowSpan: 1,
		ColSpan: 1,
	}
	if res := blankColumn(); !reflect.DeepEqual(res, expected) {
		t.Errorf("blankColumn() = %+v, expected %+v", res, expected)
	}
}

func TestColumn(t *testing.T) {
	name := "test"
	id := "test-id"
	bgColor := "black"

	expected := &Data{
		Text:    name,
		Port:    portStr(id),
		BgColor: bgColor,
		RowSpan: 1,
		ColSpan: 1,
	}
	if res := column(name, id, bgColor); !reflect.DeepEqual(res, expected) {
		t.Errorf("column(%q, %q, %q) = %+v, expected %+v", name, id, bgColor, res, expected)
	}
}

func TestNameRow(t *testing.T) {
	colSpan := 3
	name := "Test Name"
	row := nameRow(name, colSpan)

	if len(row.Data) != 1 {
		t.Errorf("Expected row to have 1 data element, but got %d", len(row.Data))
	}

	if row.Data[0].Text != name {
		t.Errorf("Expected row data to have text '%s', but got '%s'", name, row.Data[0].Text)
	}

	if row.Data[0].ColSpan != colSpan {
		t.Errorf("Expected row data to have colspan %d, but got %d", colSpan, row.Data[0].ColSpan)
	}
}

func TestBlankRow(t *testing.T) {
	row := blankRow()

	if len(row.Data) != rowStartEmptyBlockNum {
		t.Errorf("Expected %v empty blocks, got %v", rowStartEmptyBlockNum, len(row.Data))
	}

	for i, d := range row.Data {
		if d.Text != "" {
			t.Errorf("Expected empty block, but found non-empty block at index %v", i)
		}

		if d.Port != "" {
			t.Errorf("Expected empty port, but found non-empty port at index %v", i)
		}

		if d.BgColor != "white" {
			t.Errorf("Expected background color white, but found %v at index %v", d.BgColor, i)
		}

		if d.RowSpan != 1 {
			t.Errorf("Expected row span of 1, but found %v at index %v", d.RowSpan, i)
		}

		if d.ColSpan != 1 {
			t.Errorf("Expected column span of 1, but found %v at index %v", d.ColSpan, i)
		}
	}
}
