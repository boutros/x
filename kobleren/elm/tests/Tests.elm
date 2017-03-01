module Tests exposing (..)

import Test exposing (..)
import Expect
import RDF.RDF exposing (..)
import RDF.Graph as Graph
import Parser


mustOk : String -> List TriplePattern
mustOk ntriples =
    case Graph.fromString ntriples of
        Err err ->
            let
                _ =
                    Debug.crash (toString err)
            in
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
            [ test "All terms as URIs" <|
                \_ ->
                    Expect.equalLists (mustOk "<a> <b> <c> .")
                        [ TriplePattern (TermURI "a") (TermURI "b") (TermURI "c")
                        ]
            , test "Blank node as subject" <|
                \_ ->
                    Expect.equalLists (mustOk "_:xyz1 <b> <c> .")
                        [ TriplePattern (TermBlankNode "xyz1") (TermURI "b") (TermURI "c")
                        ]
            , test "Blank node as object" <|
                \_ ->
                    Expect.equalLists (mustOk "<a> <b> _:xyz1 .")
                        [ TriplePattern (TermURI "a") (TermURI "b") (TermBlankNode "xyz1")
                        ]
            , test "String literal" <|
                \_ ->
                    Expect.equalLists (mustOk "<a> <b> \"xyz\" .")
                        [ TriplePattern
                            (TermURI "a")
                            (TermURI "b")
                            (TermLiteral <| Literal "xyz" Nothing xsdString)
                        ]
            , test "Language-tagged literal" <|
                \_ ->
                    Expect.equalLists (mustOk "<a> <b> \"hei\"@no .")
                        [ TriplePattern
                            (TermURI "a")
                            (TermURI "b")
                            (TermLiteral <| Literal "hei" (Just "no") rdfLangString)
                        ]
            , test "Typed literal" <|
                \_ ->
                    Expect.equalLists (mustOk "<a> <b> \"99\"^^<http://www.w3.org/2001/XMLSchema#integer> .")
                        [ TriplePattern
                            (TermURI "a")
                            (TermURI "b")
                            (TermLiteral <| Literal "99" Nothing xsdInteger)
                        ]
            ]
        , describe "Malformed"
            [ test "Missing object" <|
                \_ ->
                    Expect.equal (mustFail "<a> <b> .") "1:9: parsing object: unexpected character: \".\""
            ]
        ]
