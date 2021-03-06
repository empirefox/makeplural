package plural

import (
	"strconv"

	"golang.org/x/text/language"
)

type PluralInfo struct {
	Cultures []Culture
	Others   []string

	culturesMap map[language.Tag]*Culture
	othersMap   map[language.Tag]bool
}

func (pi *PluralInfo) Validate(langs []string) (parseFailed, findFailed []string, ok bool) {
	parseFailed = make([]string, 0, len(langs))
	findFailed = make([]string, 0, len(langs))
	for _, item := range langs {
		lang, err := language.Parse(item)
		if err != nil {
			parseFailed = append(parseFailed, item)
			continue
		}
		if _, _, ok := pi.Find(lang); !ok {
			findFailed = append(findFailed, item)
		}
	}

	ok = len(parseFailed)+len(findFailed) == 0
	return
}

func (pi *PluralInfo) Langs() []string {
	all := make([]string, 0, 256)
	for i := range pi.Cultures {
		all = append(all, pi.Cultures[i].Langs...)
	}
	all = append(all, pi.Others...)
	return all
}

func (pi *PluralInfo) Find(lang language.Tag) (c *Culture, on language.Tag, found bool) {
	on = lang
	c, found = pi.CulturesMap()[lang]
	if found {
		return
	}

	found = pi.IsOthers(lang)
	if found {
		return
	}

	base, confidence := lang.Base()
	if confidence == language.No {
		return
	}

	lang2, err := language.Compose(base)
	if err != nil {
		return
	}

	if lang2 == lang {
		lang2 = lang.Parent()
	}

	return pi.Find(lang2)
}

func (pi *PluralInfo) CulturesMap() map[language.Tag]*Culture {
	if pi.culturesMap == nil {
		pi.culturesMap = make(map[language.Tag]*Culture, 256)
		for i := range pi.Cultures {
			for _, lang := range pi.Cultures[i].Langs {
				pi.culturesMap[language.MustParse(lang)] = &pi.Cultures[i]
			}
		}
	}
	return pi.culturesMap
}

func (pi *PluralInfo) IsOthers(cultrue language.Tag) bool {
	if pi.othersMap == nil {
		pi.othersMap = make(map[language.Tag]bool, len(pi.Others))
		for _, lang := range pi.Others {
			pi.othersMap[language.MustParse(lang)] = true
		}
	}
	return pi.othersMap[cultrue]
}

type Culture struct {
	Langs []string

	// Symbols plus P
	F, I, N, V, T, W, P Symbol

	// Cardinal defines the plural rules for numbers indicating quantities.
	Cardinal Cases

	// Ordinal defines the plural rules for numbers indicating position
	// (first, second, etc.).
	Ordinal Cases

	// Vars only come from mod
	Vars []Var

	Tests UnitTests
}

func (c Culture) HasVars() bool {
	return len(c.Vars) != 0 ||
		c.F.Use() ||
		c.I.Use() ||
		c.N.Use() ||
		c.V.Use() ||
		c.T.Use() ||
		c.W.Use() ||
		c.P.Use()
}
func (c Culture) NeedFinvtw() bool      { return c.F.Use() || c.V.Use() || c.T.Use() || c.W.Use() }
func (c Culture) HasCardinal() bool     { return len(c.Cardinal) != 0 }
func (c Culture) HasOrdinal() bool      { return len(c.Ordinal) != 0 }
func (c Culture) HasTest() bool         { return c.HasCardinalTest() || c.HasOrdinalTest() }
func (c Culture) HasCardinalTest() bool { return len(c.Tests.Cardinal) != 0 }
func (c Culture) HasOrdinalTest() bool  { return len(c.Tests.Ordinal) != 0 }

type Case struct {
	Form string
	Cond string
}

type Cases []Case

func (s Cases) ToMap() (m map[string]*Case) {
	m = make(map[string]*Case, len(s))
	for i := range s {
		m[s[i].Form] = &s[i]
	}
	return
}

type Var struct {
	Symbol Symbol
	Mod    int
}

func (v Var) Name() string { return v.Symbol.Name() + strconv.Itoa(v.Mod) }

type UnitTest struct {
	Expected string
	Integers []string
	Decimals []string
}

type UnitTests struct {
	Cardinal []UnitTest
	Ordinal  []UnitTest
}
