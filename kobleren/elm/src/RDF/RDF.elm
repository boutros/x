module RDF.RDF exposing (..)

-- TYPES


type Term
    = TermURI String
    | TermLiteral Literal
    | TermBlankNode String
    | TermVar String


type alias Literal =
    { value : String
    , language : Maybe String
    , datatype : Term
    }


type alias TriplePattern =
    { subject : Term
    , predicate : Term
    , object : Term
    }



-- FUNCTIONS


serialize : Term -> String
serialize term =
    case term of
        TermURI uri ->
            "<" ++ uri ++ ">"

        TermBlankNode id ->
            "_:" ++ id

        TermLiteral lit ->
            case lit.language of
                Just lang ->
                    "\"" ++ lit.value ++ "\"@" ++ lang

                Nothing ->
                    if lit.datatype == xsdString then
                        "\"" ++ lit.value ++ "\""
                    else
                        "\"" ++ lit.value ++ "\"^^" ++ serialize lit.datatype

        TermVar var ->
            "?" ++ var



-- USEFULL URI CONSTANTS


xsdString =
    TermURI "http://www.w3.org/2001/XMLSchema#string"


xsdInteger =
    TermURI "http://www.w3.org/2001/XMLSchema#integer"


rdfLangString =
    TermURI "http://www.w3.org/1999/02/22-rdf-syntax-ns#langString"


rdfType =
    TermURI "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"
