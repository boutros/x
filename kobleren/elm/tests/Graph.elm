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


graph : Test
graph =
    describe "Graph"
        [ describe "decoding/encoding"
            [ test "encode-decode-encode roundtrip" <|
                \_ ->
                    let
                        nt =
                            "<a> <b> <c> .\n"
                    in
                        Expect.equal nt (Graph.toNTriples (mustParse nt))
            ]
        , describe "create and manipulate"
            []
        ]
