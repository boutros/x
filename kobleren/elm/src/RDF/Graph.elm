module RDF.Graph exposing (Graph, empty, toNTriples, fromNTriples, query)

import RDF.RDF exposing (..)
import RDF.Parser exposing (parseNTriples)
import Dict
import Set
import List


-- TYPES


type alias Graph =
    { bnodeId : Int
    , triples : List TriplePattern
    }



-- EXPORTED FUNCTIONS


{-| Create an empty graph.
-}
empty : Graph
empty =
    Graph 0 []


{-| Parse NTriples into a graph.
-}
fromNTriples : String -> Result String Graph
fromNTriples nt =
    case parseNTriples nt of
        Err err ->
            Result.Err err

        Ok ntriples ->
            Result.Ok (fromTriples ntriples)


{-| Serialize graph as NTriples.
-}
toNTriples : Graph -> String
toNTriples graph =
    List.map
        (\triple ->
            serialize triple.subject
                ++ " "
                ++ serialize triple.predicate
                ++ " "
                ++ serialize triple.object
                ++ " .\n"
        )
        graph.triples
        |> List.foldr (++) ""


{-| Query the graph for matching triple-patterns.
-}
query : List TriplePattern -> Graph -> List TriplePattern
query patterns graph =
    List.concatMap (\pattern -> match pattern graph) patterns



-- QUERY HELPER FUNCTIONS


match : TriplePattern -> Graph -> List TriplePattern
match pattern graph =
    List.filter (tripleMatch pattern) graph.triples


tripleMatch : TriplePattern -> TriplePattern -> Bool
tripleMatch hasTriple qTriple =
    (termMatch hasTriple.subject qTriple.subject)
        && (termMatch hasTriple.predicate qTriple.predicate)
        && (termMatch hasTriple.object qTriple.object)


termMatch : Term -> Term -> Bool
termMatch a b =
    case a of
        TermVar _ ->
            True

        TermURI uri ->
            (TermURI uri) == b

        TermLiteral lit ->
            (TermLiteral lit) == b

        --TODO
        TermBlankNode _ ->
            True



-- MUTATING HELPER FUNCTIONS


insert1 : TriplePattern -> Graph -> Graph
insert1 triple graph =
    if (List.member triple graph.triples) then
        graph
    else
        { graph | triples = triple :: graph.triples }



-- DECODING HELPER FUNCTIONS


fromTriples : List TriplePattern -> Graph
fromTriples triples =
    List.foldr insert1 empty triples



-- insert1 -> TriplePattern -> Graph
-- delete1 -> TriplePattern -> Graph
-- insert -> List TriplePattern -> Graph
-- delete -> List TriplePattern -> Graph
-- set/replace -> TriplePattern -> Graph
