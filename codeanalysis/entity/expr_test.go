package entity

import (
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/packages"
	"reflect"
	"testing"
)

func TestExpressionVisit(t *testing.T) {
	pkg := &packages.Package{
		ID: "example.com/mypackage",
	}

	tests := []struct {
		expr ast.Expr
		want []exprInfo
	}{
		{
			expr: &ast.Ident{Name: "User"},
			want: []exprInfo{
				{
					sel:  "example.com/mypackage",
					val:  "User",
					ship: OneOne,
				},
			},
		},
	}

	for _, tt := range tests {
		expr := &expression{pkg: pkg, expr: tt.expr}
		var got []exprInfo
		cb := func(path, name string, ship RelationShip) {
			got = append(got, exprInfo{sel: path, val: name, ship: ship})
		}
		expr.visit(cb)

		if len(got) != len(tt.want) {
			t.Errorf("expr.visit(%#v): got %d exprInfos, want %d", tt.expr, len(got), len(tt.want))
			continue
		}

		for i := range tt.want {
			if got[i] != tt.want[i] {
				t.Errorf("expr.visit(%#v)[%d]: got %v, want %v", tt.expr, i, got[i], tt.want[i])
			}
		}
	}
}

func TestGetExprsInfo(t *testing.T) {
	type testStruct struct {
		input ast.Expr
		want  []*exprInfo
	}

	tests := []testStruct{
		{
			input: &ast.Ident{Name: "foo"},
			want: []*exprInfo{
				{sel: "", val: "foo", ship: OneOne},
			},
		},
		{
			input: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "pkg"},
				Sel: &ast.Ident{Name: "foo"},
			},
			want: []*exprInfo{
				{sel: "pkg", val: "foo", ship: OneOne},
			},
		},
		{
			input: &ast.StarExpr{
				X: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "pkg"},
					Sel: &ast.Ident{Name: "foo"},
				},
			},
			want: []*exprInfo{
				{sel: "pkg", val: "foo", ship: OneOne},
			},
		},
		{
			input: &ast.ArrayType{
				Elt: &ast.Ident{Name: "foo"},
			},
			want: []*exprInfo{
				{sel: "", val: "foo", ship: OneMany},
			},
		},
		{
			input: &ast.MapType{
				Key:   &ast.Ident{Name: "foo"},
				Value: &ast.Ident{Name: "bar"},
			},
			want: []*exprInfo{
				{sel: "", val: "foo", ship: OneOne},
				{sel: "", val: "bar", ship: OneOne},
			},
		},
	}

	for _, test := range tests {
		got, err := getExprsInfo(test.input)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
			continue
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("Input: %s\nGot: %v\nWant: %v", test.input, got, test.want)
		}
	}
}

func TestGetExprInfo(t *testing.T) {
	type test struct {
		name     string
		expr     ast.Expr
		expected *exprInfo
		err      bool
	}
	tests := []test{
		{
			name: "StarExpr",
			expr: &ast.StarExpr{
				X: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "pkg"},
					Sel: &ast.Ident{Name: "Type"},
				},
			},
			expected: &exprInfo{
				sel:  "pkg",
				val:  "Type",
				ship: OneOne,
			},
			err: false,
		},
		{
			name: "StarExpr-Ident",
			expr: &ast.StarExpr{
				X: &ast.Ident{
					Name: "Type",
				},
			},
			expected: &exprInfo{
				val:  "Type",
				ship: OneOne,
			},
			err: false,
		},
		{
			name: "SelectorExpr",
			expr: &ast.SelectorExpr{
				X: &ast.Ident{Name: "pkg"},
				Sel: &ast.Ident{
					Name: "Type",
				},
			},
			expected: &exprInfo{
				sel:  "pkg",
				val:  "Type",
				ship: OneOne,
			},
			err: false,
		},
		{
			name: "Ident",
			expr: &ast.Ident{
				Name: "Type",
			},
			expected: &exprInfo{
				val:  "Type",
				ship: OneOne,
			},
			err: false,
		},
		{
			name: "ArrayType",
			expr: &ast.ArrayType{
				Elt: &ast.Ident{
					Name: "Type",
				},
			},
			expected: &exprInfo{
				val:  "Type",
				ship: OneMany,
			},
			err: false,
		},
		{
			name: "InvalidMapType",
			expr: &ast.MapType{
				Key: &ast.Ident{
					Name: "KeyType",
				},
				Value: &ast.MapType{
					Key: &ast.Ident{
						Name: "KeyType",
					},
					Value: &ast.Ident{
						Name: "ValueType",
					},
				},
			},
			expected: nil,
			err:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exprInfo, err := getExprInfo(tt.expr)
			if tt.err && err == nil {
				t.Errorf("Expected an error, but none occurred")
			}
			if !tt.err && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tt.expected != nil {
				if exprInfo.sel != tt.expected.sel {
					t.Errorf("Selector not equal. expected %q but got %q", tt.expected.sel, exprInfo.sel)
				}
				if exprInfo.val != tt.expected.val {
					t.Errorf("Value not equal. expected %q but got %q", tt.expected.val, exprInfo.val)
				}
				if exprInfo.ship != tt.expected.ship {
					t.Errorf("RelationShip not equal. expected %v but got %v", tt.expected.ship, exprInfo.ship)
				}
			}
		})
	}
}

func TestExtractExpr(t *testing.T) {
	cases := []struct {
		name  string
		input ast.Expr
		want  struct {
			sel  string
			val  string
			ship RelationShip
		}
	}{
		{
			name: "Test StarExpr",
			input: &ast.StarExpr{
				X: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "foo"},
					Sel: &ast.Ident{Name: "bar"},
				},
			},
			want: struct {
				sel  string
				val  string
				ship RelationShip
			}{
				sel:  "foo",
				val:  "bar",
				ship: OneOne,
			},
		},
		{
			name: "Test Ident",
			input: &ast.Ident{
				Name: "foo",
			},
			want: struct {
				sel  string
				val  string
				ship RelationShip
			}{
				val:  "foo",
				ship: OneOne,
			},
		},
		{
			name: "Test SelectorExpr",
			input: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "foo"},
				Sel: &ast.Ident{Name: "bar"},
			},
			want: struct {
				sel  string
				val  string
				ship RelationShip
			}{
				sel:  "foo",
				val:  "bar",
				ship: OneOne,
			},
		},
		{
			name: "Test ArrayType",
			input: &ast.ArrayType{
				Elt: &ast.Ident{Name: "foo"},
			},
			want: struct {
				sel  string
				val  string
				ship RelationShip
			}{
				val:  "foo",
				ship: OneMany,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sel, val, ship, err := extractExpr(c.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if sel != c.want.sel {
				t.Errorf("got sel=%q, want sel=%q", sel, c.want.sel)
			}
			if val != c.want.val {
				t.Errorf("got val=%q, want val=%q", val, c.want.val)
			}
			if ship != c.want.ship {
				t.Errorf("got ship=%v, want ship=%v", ship, c.want.ship)
			}
		})
	}
}

func TestTrimDoubleQuote(t *testing.T) {
	input := "\"hello world\""
	expected := "hello world"
	result := trimDoubleQuote(input)
	if result != expected {
		t.Errorf("trimDoubleQuote(%q) = %q; want %q", input, result, expected)
	}
}

func TestGetPath(t *testing.T) {
	imports := []*ast.ImportSpec{
		{
			Path: &ast.BasicLit{
				Kind:  token.STRING,
				Value: "\"github.com/dddplayer/core\"",
			},
		},
	}
	name := "core"
	expected := "github.com/dddplayer/core"
	result := getPath(imports, name)
	if result != expected {
		t.Errorf("getPath(%v, %q) = %q; want %q", imports, name, result, expected)
	}
}

func TestIsBasicTypes(t *testing.T) {
	testCases := []struct {
		name     string
		expected bool
	}{
		{"string", true},
		{"int", true},
		{"bool", false},
		{"any", true},
		{"unknown", false},
	}

	for _, tc := range testCases {
		result := isBasicTypes(tc.name)
		if result != tc.expected {
			t.Errorf("isBasicTypes(%q) = %t; want %t", tc.name, result, tc.expected)
		}
	}
}
