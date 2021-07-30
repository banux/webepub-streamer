package pub

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContributorParseJSONString(t *testing.T) {
	c1 := Contributor{
		LocalizedName: NewLocalizedStringFromString("John Smith"),
	}
	var c2 Contributor
	assert.NoError(t, json.Unmarshal([]byte(`"John Smith"`), &c2))
	assert.Equal(t, c1, c2, "parsed JSON string should be equal to string")
}

func TestContributorParseMinimalJSON(t *testing.T) {
	c1 := Contributor{
		LocalizedName: NewLocalizedStringFromString("John Smith"),
	}
	var c2 Contributor
	assert.NoError(t, json.Unmarshal([]byte(`{"name": "John Smith"}`), &c2))
	assert.Equal(t, c1, c2, "parsed JSON object should be equal to contributor object")
}

func TestContributorParseFullJSON(t *testing.T) {
	sortAs := NewLocalizedStringFromString("greenwood")
	position := float64(4.0)
	c1 := Contributor{
		LocalizedName:   NewLocalizedStringFromString("Colin Greenwood"),
		LocalizedSortAs: &sortAs,
		Identifier:      "colin",
		Roles:           []string{"bassist"},
		Position:        &position,
		Links: []Link{
			{
				Href: "http://link1",
			},
			{
				Href: "http://link2",
			},
		},
	}
	var c2 Contributor
	assert.NoError(t, json.Unmarshal([]byte(`{
		"name": "Colin Greenwood",
		"identifier": "colin",
		"sortAs": "greenwood",
		"role": "bassist",
		"position": 4,
		"links": [
			{"href": "http://link1"},
			{"href": "http://link2"}
		]
	}`), &c2))
	assert.Equal(t, c1, c2, "parsed JSON object should be equal to contributor object")
}

func TestContributorParseJSONWithDuplicateRoles(t *testing.T) {
	c1 := Contributor{
		LocalizedName: NewLocalizedStringFromString("Thom Yorke"),
		Roles:         []string{"singer", "guitarist"},
	}
	var c2 Contributor
	assert.NoError(t, json.Unmarshal([]byte(`{
		"name": "Thom Yorke",
		"role": ["singer", "guitarist", "guitarist"]
	}`), &c2))
	assert.Equal(t, c1, c2, "parsed JSON object should be equal to contributor object")
}

func TestContributorParseRequiresName(t *testing.T) {
	var c Contributor
	assert.NoError(t, json.Unmarshal([]byte(`{"identifier": "loremipsonium"}`), &c))
	assert.Equal(t, c, Contributor{}, "parsed JSON object should be empty Contributor")
}

func TestContributorNameFromDefaultTranslation(t *testing.T) {
	c := Contributor{
		LocalizedName: LocalizedString{
			translations: map[string]string{
				"en": "Jonny Greenwood",
				"fr": "Jean Boisvert",
			},
		},
	}
	assert.Equal(t, c.Name(), "Jonny Greenwood", "Contributor's name should be equal to \"Jonny Greenwood\"")
}

func TestContributorMinimalJSON(t *testing.T) {
	l := Contributor{LocalizedName: NewLocalizedStringFromString("Colin Greenwood")}
	s, err := json.Marshal(l)
	assert.NoError(t, err)
	assert.JSONEq(t, `"Colin Greenwood"`, string(s), "JSON of Contributor with one name should be equal to JSON representation")
}

func TestContributorMinimalJSONWithLocalizedName(t *testing.T) {
	n := LocalizedString{}
	n.SetTranslation("en", "Colin Greenwood")
	n.SetTranslation("fr", "Colain Grinwoud")
	l := Contributor{LocalizedName: n}
	s, err := json.Marshal(l)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"name": {"fr": "Colain Grinwoud", "en": "Colin Greenwood"}}`, string(s), "JSON of Contributor with one name should be equal to JSON representation")
}

func TestContributorFullJSON(t *testing.T) {
	pos := float64(4.0)
	sortAs := NewLocalizedStringFromString("greenwood")
	l := Contributor{
		LocalizedName:   NewLocalizedStringFromString("Colin Greenwood"),
		LocalizedSortAs: &sortAs,
		Identifier:      "colin",
		Roles:           []string{"bassist"},
		Position:        &pos,
		Links: []Link{
			{
				Href:      "http://link1",
				Templated: true,
			},
			{
				Href:      "http://link2",
				Templated: false,
			},
		},
	}
	s, err := json.Marshal(l)
	assert.NoError(t, err)
	assert.JSONEq(t, `{
		"name": "Colin Greenwood",
		"identifier": "colin",
		"sortAs": "greenwood",
		"role": ["bassist"],
		"position": 4.0,
		"links": [
			{"href": "http://link1", "templated": true},
			{"href": "http://link2"}
		]
	}`, string(s), "JSON of Contributor with all fields filled should be equal to JSON representation")
}
