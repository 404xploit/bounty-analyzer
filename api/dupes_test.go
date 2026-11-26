package api

import "testing"

func TestFindDuplicateMatchesNormalizesAndSorts(t *testing.T) {
	cache := Cache{Programs: map[string]Program{
		"zeta": {Name: "zeta", Reports: []Report{{ID: "2", Title: "GraphQL authorization"}}},
		"alpha": {Name: "alpha", Reports: []Report{
			{ID: "3", Title: "IDOR in profile"},
			{ID: "1", Title: "GraphQL query leak"},
		}},
	}}

	matches := FindDuplicateMatches(" IDOR,idor,, GraphQL ", cache)
	if len(matches) != 3 {
		t.Fatalf("expected 3 unique matches, got %#v", matches)
	}
	got := []string{
		matches[0].ProgramName + ":" + matches[0].ReportTitle + ":" + matches[0].Keyword,
		matches[1].ProgramName + ":" + matches[1].ReportTitle + ":" + matches[1].Keyword,
		matches[2].ProgramName + ":" + matches[2].ReportTitle + ":" + matches[2].Keyword,
	}
	want := []string{
		"alpha:GraphQL query leak:GraphQL",
		"alpha:IDOR in profile:IDOR",
		"zeta:GraphQL authorization:GraphQL",
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected match order/content: got %v, want %v", got, want)
		}
	}
}

func TestFindDuplicateMatchesIgnoresEmptyKeywords(t *testing.T) {
	cache := Cache{Programs: map[string]Program{
		"one": {Name: "one", Reports: []Report{{Title: "Anything"}}},
	}}
	if matches := FindDuplicateMatches(", ,", cache); len(matches) != 0 {
		t.Fatalf("empty keywords must not match every report: %#v", matches)
	}
}
