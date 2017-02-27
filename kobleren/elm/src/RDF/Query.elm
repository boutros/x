module RDF.Query exposing (..)

import RDF.RDF exposing (..)


matchOne : Term -> Term -> List TriplePattern
matchOne subj pred =
    [ (TriplePattern subj pred (TermVar "x")) ]
