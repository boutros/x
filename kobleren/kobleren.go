package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
)

var (
	roles = map[string]string{
		"http://data.deichman.no/role#author":      "forfatter",
		"http://data.deichman.no/role#contributor": "bidragsyter",
		"http://data.deichman.no/role#editor":      "redaktør",
		"http://data.deichman.no/role#illustrator": "illustratør",
		"http://data.deichman.no/role#performer":   "utøver",
		"http://data.deichman.no/role#reader":      "innleser",
		"http://data.deichman.no/role#translator":  "oversetter",
	}
)

type searchResults struct {
	Took      int // ms
	TotalHits int
	Hits      []searchHit
}

type searchHit struct {
	ID       string // URI
	Type     string // work|publication|person|place|corporation|serial|workSeries|subject|genre|instrument|compositionType|event
	Label    string
	Abstract string
}

type person struct {
	Name      string `json:"name"`
	URI       string `json:"uri"`
	BirthYear string `json:"birthYear"`
	DeathYear string `json:"deathYear"`
	Work      []struct {
		URI             string `json:"uri"`
		Role            string `json:"role"`
		MainTitle       string `json:"mainTitle"`
		Subtitle        string `json:"subtitle"`
		PartTitle       string `json:"partTitle"`
		PartNumber      string `json:"partNumber"`
		PublicationYear string `json:"publicationYear"`
	} `json:"work"`
}

type stringArray []string

func (s *stringArray) MarshalJSON() ([]byte, error) {
	return json.Marshal([]string(*s))
}

func (s *stringArray) UnmarshalJSON(data []byte) error {
	if len(data) > 1 && data[0] == '[' {
		var v []string
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*s = stringArray(v)
		return nil
	}
	var v string
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*s = append(*s, v)
	return nil
}

type publication struct {
	Subject              []string    `json:"subject"`
	FirstPublicationYear string      `json:"firstPublicationYear"`
	Isbn                 stringArray `json:"isbn"`
	Bio                  string      `json:"bio"`
	MainEntryName        string      `json:"mainEntryName"`
	Language             stringArray `json:"language"`
	Audiences            []string    `json:"audiences"`
	Kd                   string      `json:"kd"`
	FictionNonfiction    string      `json:"fictionNonfiction"`
	WorkPartNumber       string      `json:"workPartNumber"`
	Languages            []string    `json:"languages"`
	Mt                   string      `json:"mt"`
	URI                  string      `json:"uri"`
	MainTitle            string      `json:"mainTitle"`
	Subtitle             string      `json:"subtitle"`
	PartTitle            string      `json:"partTitle"`
	PublicationYear      string      `json:"publicationYear"`
	Dewey                stringArray `json:"dewey"`
	PartNumber           string      `json:"partNumber"`
	Contributors         []struct {
		Agent struct {
			URI  string `json:"uri"`
			Name string `json:"name"`
		} `json:"agent"`
		Role string `json:"role"`
	} `json:"contributors"`
	WorkURI   string `json:"workUri"`
	Mediatype string `json:"mediatype"`
}

type work struct {
	MainTitle       string `json:"mainTitle"`
	Subtitle        string `json:"subtitle"`
	PartTitle       string `json:"partTitle"`
	PartNumber      string `json:"partNumber"`
	PublicationYear string `json:"publicationYear"`
	Contributors    []struct {
		Agent struct {
			URI  string `json:"uri"`
			Name string `json:"name"`
		} `json:"agent"`
		MainEntry bool   `json:"mainEntry"`
		Role      string `json:"role"`
	} `json:"contributors"`
	URI string `json:"uri"`
}

type corporation struct {
	Name          string `json:"name"`
	Specification string `json:"specification"`
	Subdivision   string `json:"subdivision"`
	URI           string `json:"uri"`
}

func personToHit(p person) searchHit {
	label := p.Name
	if p.BirthYear != "" {
		label += " ("
		label += p.BirthYear
		label += "-"
		if p.DeathYear != "" {
			label += p.DeathYear
		}
		label += ")"
	}
	var b bytes.Buffer
	sort.Slice(p.Work, func(i, j int) bool {
		// sort by role
		if p.Work[i].Role < p.Work[j].Role {
			return true
		}
		if p.Work[i].Role > p.Work[j].Role {
			return false
		}
		// then by publication year
		return p.Work[i].PublicationYear > p.Work[j].PublicationYear
	})
	var curRole string
	for _, work := range p.Work {
		if work.Role != curRole {
			curRole = work.Role
			b.WriteString(work.Role)
			b.WriteString(" av:\n")
		}
		if work.PublicationYear != "" {
			b.WriteString(work.PublicationYear)
			b.WriteString(": ")
		}
		b.WriteString(isbdTitle(work.MainTitle, work.Subtitle, work.PartTitle, work.PartNumber))
		b.WriteString("\n")
	}
	return searchHit{
		ID:       p.URI,
		Type:     "person",
		Label:    label,
		Abstract: b.String(),
	}
}

func workToHit(w work) searchHit {
	var title string
	var abstract bytes.Buffer
	if w.PublicationYear != "" {
		abstract.WriteString("Første gang utgitt: ")
		abstract.WriteString(w.PublicationYear)
		abstract.WriteString("\n")
	}
	for _, c := range w.Contributors {
		if c.MainEntry {
			title = c.Agent.Name
			title += ": "
			continue
		}
		abstract.WriteString(strings.Title(roles[c.Role]))
		abstract.WriteString(": ")
		abstract.WriteString(c.Agent.Name)
		abstract.WriteString(". ")
	}
	title += isbdTitle(w.MainTitle, w.Subtitle, w.PartTitle, w.PartNumber)
	return searchHit{
		ID:       w.URI,
		Type:     "work",
		Label:    title,
		Abstract: abstract.String(),
	}
}

func publicationToHit(p publication) searchHit {
	var title string
	if p.MainEntryName != "" {
		title = p.MainEntryName + " - "
	}
	title += isbdTitle(p.MainTitle, p.Subtitle, p.PartTitle, p.PartNumber)
	if p.PublicationYear != "" {
		title += " (" + p.PublicationYear + ")"
	}
	var abstract bytes.Buffer
	contribs := make(map[string][]string)
	for _, c := range p.Contributors {
		if p.MainEntryName == c.Agent.Name {
			continue
		}
		contribs[roles[c.Role]] = append(contribs[roles[c.Role]], c.Agent.Name)
	}
	for role, agents := range contribs {
		abstract.WriteString(strings.Title(role))
		abstract.WriteString(": ")
		for i, agent := range agents {
			abstract.WriteString(agent)
			if i == len(agents)-1 {
				abstract.WriteString(".\n")
			} else {
				abstract.WriteString(", ")
			}
		}
	}
	return searchHit{
		ID:       p.URI,
		Type:     "publication",
		Label:    title,
		Abstract: abstract.String(),
	}
}

func corporationToHit(c corporation) searchHit {
	title := c.Name
	if c.Subdivision != "" {
		title += " - "
		title += c.Subdivision
	}
	if c.Specification != "" {
		title += " (" + c.Specification + ")"
	}
	return searchHit{
		ID:    c.URI,
		Type:  "corporation",
		Label: title,
	}
}
func isbdTitle(mainTitle, subtitle, partTitle, partNumber string) string {
	s := mainTitle
	if subtitle != "" {
		s += " : "
		s += subtitle
	}
	if partNumber != "" {
		s += ". "
		s += partNumber
	}
	if partTitle != "" {
		s += ". "
		s += partTitle
	}
	return s
}

func parseSearchResult(filename string) (searchResults, error) {
	var res searchResults

	f, err := os.Open(filename)
	if err != nil {
		return res, err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	var nextType string
	for {
		t, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return res, err
		}

		if s, ok := t.(string); ok {
			switch s {
			case "took":
				if err := dec.Decode(&res.Took); err != nil {
					return res, err
				}
			case "total":
				if err := dec.Decode(&res.TotalHits); err != nil {
					return res, err
				}
			case "publication", "work", "person", "corporation":
				nextType = s
			case "_source":
				switch nextType {
				case "person":
					var p person
					if err := dec.Decode(&p); err != nil {
						return res, err
					}
					res.Hits = append(res.Hits, personToHit(p))
				case "publication":
					var p publication
					if err := dec.Decode(&p); err != nil {
						return res, err
					}
					res.Hits = append(res.Hits, publicationToHit(p))
				case "work":
					var w work
					if err := dec.Decode(&w); err != nil {
						return res, err
					}
					res.Hits = append(res.Hits, workToHit(w))
				case "corporation":
					var c corporation
					if err := dec.Decode(&c); err != nil {
						return res, err
					}
					res.Hits = append(res.Hits, corporationToHit(c))
				}
			}
		}
	}
	return res, nil
}

func main() {
	log.SetFlags(0)

	dummyDB := make(map[string]searchResults)
	for _, filename := range []string{"bach", "oslo", "åsen", "hamsun"} {
		res, err := parseSearchResult(filename + ".json")
		if err != nil {
			log.Fatal(err)
		}
		dummyDB[filename] = res
	}

	http.HandleFunc("/dummysearch", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")

		q := r.URL.Query().Get("q")
		if q == "" {
			http.Error(w, `missing required parameter: "q"`, http.StatusBadRequest)
			return
		}
		res := dummyDB[q]
		from := 0
		size := 10
		if fromQ := r.URL.Query().Get("from"); fromQ != "" {
			f, err := strconv.Atoi(fromQ)
			if err != nil {
				http.Error(w, `parameter "from" must be an integer`, http.StatusBadRequest)
				return
			}
			from = f
		}
		if from < len(res.Hits) {
			res.Hits = res.Hits[from:]
		}
		if sizeQ := r.URL.Query().Get("size"); sizeQ != "" {
			s, err := strconv.Atoi(sizeQ)
			if err != nil {
				http.Error(w, `parameter "size" must be an integer`, http.StatusBadRequest)
				return
			}
			size = s
		}
		if size < len(res.Hits) {
			res.Hits = res.Hits[:size]
		}

		if err := json.NewEncoder(w).Encode(res); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		panic("TODO") // proxy to elasticsearch
	})

	log.Fatal(http.ListenAndServe(":8008", nil))
}
