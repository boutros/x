module RDF.Graph exposing (fromString, Graph)

import RDF.RDF exposing (TriplePattern)
import RDF.Parser exposing (parseTriples)
import Parser


fromString : String -> Result String (List TriplePattern)
fromString ntriples =
    case Parser.run (parseTriples ntriples) ntriples of
        Err err ->
            Result.Err (formatError err)

        Ok triples ->
            Result.Ok triples


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


type alias Graph =
    List TriplePattern



-- TODO fromString : String -> Result (List TriplePattern)
-- with serialized error message
