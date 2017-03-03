module RDF.RDF exposing (..)


type alias Literal =
    { value : String
    , language : Maybe String
    , datatype : Term
    }


type Term
    = TermURI String
    | TermLiteral Literal
    | TermBlankNode String
    | TermVar String


type alias TriplePattern =
    { subject : Term
    , predicate : Term
    , object : Term
    }


xsdString =
    TermURI "http://www.w3.org/2001/XMLSchema#string"


xsdInteger =
    TermURI "http://www.w3.org/2001/XMLSchema#integer"


rdfLangString =
    TermURI "http://www.w3.org/1999/02/22-rdf-syntax-ns#langString"


rdfType =
    TermURI "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"
