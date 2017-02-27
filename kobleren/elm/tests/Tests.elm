module Tests exposing (..)

import Test exposing (..)
import Expect
import RDF.RDF exposing (..)
import RDF.Graph as Graph


mustParse : String -> List TriplePattern
mustParse ntriples =
    case Graph.fromString ntriples of
        Err _ ->
            []

        Ok triples ->
            triples


all : Test
all =
    let
        tests =
            [ ( "<a> <b> <c> ."
              , [ TriplePattern (TermURI "a") (TermURI "b") (TermURI "c")
                ]
              )
            ]
    in
        describe "RDF module"
            [ test "Parsing N-Triples" <|
                \_ ->
                    Expect.equalLists (mustParse "<a> <b> <c> .")
                        [ TriplePattern (TermURI "a") (TermURI "b") (TermURI "c")
                        ]
            ]
