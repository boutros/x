module RDF.Ontology exposing (..)

import RDF.RDF exposing (..)


ontology : String -> String
ontology path =
    "http://data.deichman.no/ontology#" ++ path


name =
    (TermURI (ontology "name"))


birthYear =
    (TermURI (ontology "birthYear"))


deathYear =
    (TermURI (ontology "deathYear"))


number =
    (TermURI (ontology "number"))


nationality =
    (TermURI (ontology "nationality"))


subdivision =
    (TermURI (ontology "subdivision"))


specification =
    (TermURI (ontology "specification"))


altLabel =
    (TermURI (ontology "altLabel"))


place =
    (TermURI (ontology "place"))
