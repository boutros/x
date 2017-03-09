module Graph exposing (graph)

import Test exposing (..)
import Expect
import RDF.RDF exposing (..)
import RDF.Graph as Graph


mustParse : String -> Graph.Graph
mustParse ntriples =
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
                            (sortLines (Graph.toNTriples (mustParse nt)))
            ]
        , describe "create and manipulate"
            []
        ]
