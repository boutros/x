module Graph exposing (graph)

import Test exposing (..)
import Expect
import RDF.RDF exposing (..)
import RDF.Graph as Graph
import RDF.Parser exposing (parseNTriples)


mustParseTriples : String -> List TriplePattern
mustParseTriples nt =
    case parseNTriples nt of
        Err _ ->
            []

        Ok triples ->
            triples


mustParseGraph : String -> Graph.Graph
mustParseGraph ntriples =
    case Graph.fromNTriples ntriples of
        Err err ->
            let
                _ =
                    Debug.crash err
            in
                Graph.empty

        Ok graph ->
            graph


sortLines : String -> List String
sortLines s =
    s
        |> String.split "\n"
        |> List.map String.trim
        |> List.filter (not << String.isEmpty)
        |> List.sort


graph : Test
graph =
    describe "Graph"
        [ describe "decoding/encoding"
            [ test "encode-decode-encode roundtrip" <|
                \_ ->
                    let
                        nt =
                            """
                            <a> <b> <c> .
                            <s> <p> "o" .
                            <s2> <p> "xyz"@en .
                            <s3> <p2> "123"^^<http://data/mytype> .
                            """
                    in
                        Expect.equalLists
                            (sortLines nt)
                            (sortLines (Graph.toNTriples (mustParseGraph nt)))
            ]
        , describe "create and manipulate"
            []
        , describe "querying"
            [ test "one triple pattern" <|
                \_ ->
                    let
                        nt =
                            """
                                <s> <p> <o> .
                                <s> <p> <o2> .
                                <s> <p2> <o3> .
                            """

                        graph =
                            (mustParseGraph nt)
                    in
                        Expect.equalLists
                            (Graph.query
                                -- TODO: (mustParseTriples "<s> <p> ?o .")
                                [ (TriplePattern
                                    (TermURI "s")
                                    (TermURI "p")
                                    (TermVar "o")
                                  )
                                ]
                                graph
                            )
                            (mustParseTriples "<s> <p> <o> .\n <s> <p> <o2> .")
            ]
        ]
