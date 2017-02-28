module Tests exposing (..)

import Test exposing (..)
import Expect
import RDF.RDF exposing (..)
import RDF.Graph as Graph
import Parser


mustOk : String -> List TriplePattern
mustOk ntriples =
    case Graph.fromString ntriples of
        Err _ ->
            []

        Ok triples ->
            triples


mustFail : String -> String
mustFail ntriples =
    case Graph.fromString ntriples of
        Err err ->
            formatError err

        Ok triples ->
            ""


formatError : Parser.Error -> String
formatError err =
    let
        position =
            (toString err.row) ++ ":" ++ (toString err.col)

        context =
            case List.head err.context of
                Nothing ->
                    ""

                Just topContext ->
                    topContext.description

        problem =
            "Problem!"
    in
        position
            ++ ": parsing "
            ++ context
            ++ ": unexpected character: \""
            ++ (String.slice (err.col - 1) (err.col + 1) err.source)
            ++ "\""


all : Test
all =
    describe "Parsing N-Triples"
        [ describe "Well-formed"
            [ test "All terms are URIs" <|
                \_ ->
                    Expect.equalLists (mustOk "<a> <b> <c> .")
                        [ TriplePattern (TermURI "a") (TermURI "b") (TermURI "c")
                        ]
            ]
        , describe "Malformed"
            [ test "Missing object" <|
                \_ ->
                    Expect.equal (mustFail "<a> <b> .") "1:9: parsing object: unexpected character: \".\""
            ]
        ]
