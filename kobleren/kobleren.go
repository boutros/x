package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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
		"http://data.deichman.no/role#actor":        "skuespiller",
		"http://data.deichman.no/role#author":       "forfatter",
		"http://data.deichman.no/role#contributor":  "bidragsyter",
		"http://data.deichman.no/role#director":     "regissør",
		"http://data.deichman.no/role#editor":       "redaktør",
		"http://data.deichman.no/role#illustrator":  "illustratør",
		"http://data.deichman.no/role#performer":    "utøver",
		"http://data.deichman.no/role#producer":     "produsent",
		"http://data.deichman.no/role#reader":       "innleser",
		"http://data.deichman.no/role#translator":   "oversetter",
		"http://data.deichman.no/role#scriptWriter": "manusforfatter",
	}
	authTypes = map[string]string{
		"publication":     "utgivelse",
		"work":            "verk",
		"person":          "person",
		"corporation":     "korporasjon",
		"subject":         "emne",
		"genre":           "sjanger",
		"instrument":      "instrument",
		"compositionType": "komposisjonstype",
		"serial":          "forlagsserie",
		"place":           "sted",
		"workSeries":      "verksserie",
		"event":           "hendelse",
	}
)

const queryTemplate = `{
  "query": {
    "match": {
      "_all": {
            "query": "%s",
            "operator": "and"
      }
    }
  }
}`

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
	Subject         []string    `json:"subject"`
	Isbn            stringArray `json:"isbn"`
	MainEntryName   string      `json:"mainEntryName"`
	Language        stringArray `json:"language"`
	Audiences       []string    `json:"audiences"`
	Kd              string      `json:"kd"`
	Languages       []string    `json:"languages"`
	Mt              string      `json:"mt"`
	URI             string      `json:"uri"`
	MainTitle       string      `json:"mainTitle"`
	Subtitle        string      `json:"subtitle"`
	PartTitle       string      `json:"partTitle"`
	PublicationYear string      `json:"publicationYear"`
	Dewey           stringArray `json:"dewey"`
	PartNumber      string      `json:"partNumber"`
	Contributors    []struct {
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

type subject struct {
	PrefLabel     string `json:"prefLabel"`
	Specification string `json:"specification"`
	URI           string `json:"uri"`
}

type serial struct {
	PartTitle       string `json:"partTitle"`
	Subtitle        string `json:"subtitle"`
	PartNumber      string `json:"partNumber"`
	URI             string `json:"uri"`
	SerialMainTitle string `json:"serialMainTitle"`
	PublishedByName string `json:"publishedByName"`
}

type place struct {
	AlternativeName string `json:"alternativeName"`
	URI             string `json:"uri"`
	PrefLabel       string `json:"prefLabel"`
	Specification   string `json:"specification"`
}

type workSeries struct {
	WorkSeriesMainTitle string `json:"workSeriesMainTitle"`
	Subtitle            string `json:"subtitle"`
	PartTitle           string `json:"partTitle"`
	PartNumber          string `json:"partNumber"`
	URI                 string `json:"uri"`
}

type event struct {
	Date            string `json:"date"`
	PlacePrefLabel  string `json:"placePrefLabel"`
	AlternativeName string `json:"alternativeName"`
	URI             string `json:"uri"`
	PrefLabel       string `json:"prefLabel"`
	Ordinal         string `json:"ordinal"`
	Specification   string `json:"specification"`
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
		return p.Work[i].PublicationYear > p.Work[j].PublicationYear
	})
	for _, work := range p.Work {
		if work.PublicationYear != "" {
			b.WriteString(work.PublicationYear)
			b.WriteString(": ")
		}
		b.WriteString(isbdTitle(work.MainTitle, work.Subtitle, work.PartTitle, work.PartNumber))
		if work.Role != "Forfatter" {
			b.WriteString(" (" + strings.ToLower(work.Role) + ")")
		}
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
	sort.Slice(w.Contributors, func(i, j int) bool {
		return w.Contributors[i].Role > w.Contributors[j].Role
	})
	var curRole string
	for _, c := range w.Contributors {
		if c.MainEntry {
			title = c.Agent.Name
			title += ": "
			continue
		}
		if c.Role != curRole {
			if curRole != "" {
				abstract.WriteString("\n")
			}
			curRole = c.Role
			abstract.WriteString(strings.Title(roles[c.Role]))
			abstract.WriteString(": ")
		}
		abstract.WriteString(c.Agent.Name)
		abstract.WriteString(". ")
	}
	title += isbdTitle(w.MainTitle, w.Subtitle, w.PartTitle, w.PartNumber)
	return searchHit{
		ID:       w.URI,
		Type:     "verk",
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
		Type:     "utgivelse",
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
		Type:  "korporasjon",
		Label: title,
	}
}

func subjectToHit(s subject) searchHit {
	title := s.PrefLabel
	if s.Specification != "" {
		title += " (" + s.Specification + ")"
	}
	return searchHit{
		ID:    s.URI,
		Type:  "emne",
		Label: title,
	}
}

func serialToHit(s serial) searchHit {
	var abstract string
	if s.PublishedByName != "" {
		abstract = "Utgitt av: " + s.PublishedByName
	}
	return searchHit{
		ID:       s.URI,
		Type:     "forlagsserie",
		Label:    isbdTitle(s.SerialMainTitle, s.Subtitle, s.PartTitle, s.PartNumber),
		Abstract: abstract,
	}
}

func workSeriesToHit(s workSeries) searchHit {
	return searchHit{
		ID:    s.URI,
		Type:  "verksserie",
		Label: isbdTitle(s.WorkSeriesMainTitle, s.Subtitle, s.PartTitle, s.PartNumber),
	}
}

func placeToHit(p place) searchHit {
	title := p.PrefLabel
	if p.Specification != "" {
		title += " (" + p.Specification + ")"
	}
	abstract := ""
	if p.AlternativeName != "" {
		abstract = "Også kjent som: " + p.AlternativeName
	}
	return searchHit{
		ID:       p.URI,
		Type:     "sted",
		Label:    title,
		Abstract: abstract,
	}
}

func eventToHit(e event) searchHit {
	title := e.PrefLabel
	if e.Specification != "" {
		title += " - " + e.Specification
	}
	if e.Ordinal != "" {
		title += " (" + e.Ordinal + ")"
	}
	if e.PlacePrefLabel != "" {
		title += ". " + e.PlacePrefLabel
	}
	if e.Date != "" {
		title += ", " + e.Date
	}
	abstract := ""
	if e.AlternativeName != "" {
		abstract = "Også kjent som: " + e.AlternativeName
	}
	return searchHit{
		ID:       e.URI,
		Type:     "hendelse",
		Label:    title,
		Abstract: abstract,
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

func parseSearchResult(r io.Reader) (searchResults, error) {
	var res searchResults

	dec := json.NewDecoder(r)
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
			case "publication", "work", "person", "corporation", "subject", "genre", "instrument", "compositionType", "serial", "place", "workSeries", "event":
				nextType = s
			case "_source":
				switch nextType {
				case "person":
					var p person
					if err := dec.Decode(&p); err != nil {
						log.Printf("error parsing %q: %v", nextType, err)
						continue
						//return res, err
					}
					res.Hits = append(res.Hits, personToHit(p))
				case "publication":
					var p publication
					if err := dec.Decode(&p); err != nil {
						log.Printf("error parsing %q: %v", nextType, err)
						continue
						//return res, err
					}
					res.Hits = append(res.Hits, publicationToHit(p))
				case "work":
					var w work
					if err := dec.Decode(&w); err != nil {
						log.Printf("error parsing %q: %v", nextType, err)
						continue
						//return res, err
					}
					res.Hits = append(res.Hits, workToHit(w))
				case "corporation":
					var c corporation
					if err := dec.Decode(&c); err != nil {
						log.Printf("error parsing %q: %v", nextType, err)
						continue
						//return res, err
					}
					res.Hits = append(res.Hits, corporationToHit(c))
				case "subject":
					var s subject
					if err := dec.Decode(&s); err != nil {
						log.Printf("error parsing %q: %v", nextType, err)
						continue
						//return res, err
					}
					res.Hits = append(res.Hits, subjectToHit(s))
				case "serial":
					var s serial
					if err := dec.Decode(&s); err != nil {
						log.Printf("error parsing %q: %v", nextType, err)
						continue
						//return res, err
					}
					res.Hits = append(res.Hits, serialToHit(s))
				case "place":
					var p place
					if err := dec.Decode(&p); err != nil {
						log.Printf("error parsing %q: %v", nextType, err)
						continue
						//return res, err
					}
					res.Hits = append(res.Hits, placeToHit(p))
				case "workSeries":
					var s workSeries
					if err := dec.Decode(&s); err != nil {
						log.Printf("error parsing %q: %v", nextType, err)
						continue
						//return res, err
					}
					res.Hits = append(res.Hits, workSeriesToHit(s))
				case "event":
					var e event
					if err := dec.Decode(&e); err != nil {
						log.Printf("error parsing %q: %v", nextType, err)
						continue
						//return res, err
					}
					res.Hits = append(res.Hits, eventToHit(e))
				default:
					log.Printf("missing parser for authority type %q", nextType)
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
		f, err := os.Open(filename + ".json")
		if err != nil {
			log.Fatal(err)
		}
		res, err := parseSearchResult(f)
		f.Close()
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
		size := 100
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
		}
	})

	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")

		q := r.URL.Query().Get("q")
		if q == "" {
			http.Error(w, `missing required parameter: "q"`, http.StatusBadRequest)
			return
		}
		var b bytes.Buffer
		b.WriteString(fmt.Sprintf(queryTemplate, q))
		resp, err := http.Post(
			"http://172.19.0.2:9200/search/_search",
			"application/json",
			&b,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		res, err := parseSearchResult(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := json.NewEncoder(w).Encode(res); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	})

	log.Fatal(http.ListenAndServe(":8008", nil))
}
