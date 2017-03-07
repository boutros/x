module RDF.Graph exposing (Graph, empty, toNTriples, fromNTriples)

import RDF.RDF exposing (TriplePattern, Term, serialize)
import RDF.Parser exposing (parseNTriples)
import Dict
import Set
import List


type alias Graph =
    { bnodeId : Int
    , term2id : Dict.Dict String Int
    , id2term : Dict.Dict Int String
    , spo : Dict.Dict ( Int, Int ) (Set.Set Int)
    }


empty : Graph
empty =
    Graph 0 Dict.empty Dict.empty Dict.empty


insert : Graph -> TriplePattern -> Graph
insert graph triple =
    let
        ( s, g1 ) =
            insertTerm triple.subject graph

        ( p, g2 ) =
            insertTerm triple.predicate g1

        ( o, g3 ) =
            insertTerm triple.predicate g2
    in
        indexTriple ( s, p, o ) g3


insertTerm : Term -> Graph -> ( Int, Graph )
insertTerm term graph =
    let
        t =
            serialize term
    in
        case Dict.get t graph.term2id of
            Just id ->
                ( id, graph )

            Nothing ->
                let
                    id =
                        Dict.size graph.term2id + 1

                    term2id =
                        Dict.insert t id graph.term2id

                    id2term =
                        Dict.insert id t graph.id2term
                in
                    ( id, { graph | id2term = id2term, term2id = term2id } )


indexTriple : ( Int, Int, Int ) -> Graph -> Graph
indexTriple ( s, p, o ) graph =
    let
        oldO =
            maybeToSet (Dict.get ( s, p ) graph.spo)

        newO =
            Set.insert o oldO
    in
        { graph | spo = Dict.insert ( s, p ) newO }


fromNTriples : String -> Result String Graph
fromNTriples nt =
    case parseNTriples nt of
        Err err ->
            Result.Err err

        Ok ntriples ->
            Result.Ok (fromTriples ntriples)


fromTriples : List TriplePattern -> Graph
fromTriples triples =
    empty


toNTriples : Graph -> String
toNTriples graph =
    "TODO"



--maybeToList : Maybe a -> List b


maybeToSet m =
    case m of
        Nothing ->
            Set.empty

        Just x ->
            x



--addIfNotPresent : a -> List a -> List a


addIfNotPresent a list =
    if (List.member a list) then
        list
    else
        a :: list



--
-- insert1 -> TriplePattern -> Graph
-- delete1 -> TriplePattern -> Graph
-- insert -> List TriplePattern -> Graph
-- delete -> List TriplePattern -> Graph
-- set/replace -> TriplePattern -> Graph
