module RDF.Graph exposing (Graph, empty, toNTriples, fromNTriples, query)

import RDF.RDF exposing (..)
import RDF.Parser exposing (parseNTriples)
import Dict
import Set
import List


-- TYPES


type alias Graph =
    { bnodeId : Int
    , term2id : Dict.Dict String Int
    , id2term : Dict.Dict Int String
    , spo : Dict.Dict ( Int, Int ) (Set.Set Int)
    }



-- EXPORTED FUNCTIONS


{-| Create an empty graph.
-}
empty : Graph
empty =
    Graph 0 Dict.empty Dict.empty Dict.empty


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
    Dict.toList graph.spo
        |> List.map
            (\( ( s, p ), o ) ->
                let
                    sp =
                        mustSerialize s graph.id2term
                            ++ " "
                            ++ mustSerialize p graph.id2term
                            ++ " "

                    olist =
                        Set.toList o

                    -- TODO refactor this mess
                    all =
                        \l ->
                            case l of
                                [] ->
                                    Debug.crash "BUG: empty object list in SPO index"

                                [ o1 ] ->
                                    sp ++ mustSerialize o1 graph.id2term ++ " .\n"

                                o1 :: rest ->
                                    sp ++ mustSerialize o1 graph.id2term ++ " .\n" ++ all rest
                in
                    all
                        olist
            )
        |> List.foldr (++) ""


{-| Query the graph for matching triple-patterns.
-}
query : List TriplePattern -> Graph -> List TriplePattern
query patterns graph =
    List.concatMap (\pattern -> match pattern graph) patterns



-- QUERY HELPER FUNCTIONS


match : TriplePattern -> Graph -> List TriplePattern
match pattern graph =
    List.map (\ids -> tripleFromIds ids graph) (matchAll graph)


matchAll : Graph -> List ( Int, Int, Int )
matchAll graph =
    Dict.toList graph.spo
        |> List.concatMap
            (\( ( s, p ), objs ) ->
                List.map (\o -> ( s, p, o )) (Set.toList objs)
            )


tripleFromIds : ( Int, Int, Int ) -> Graph -> TriplePattern
tripleFromIds ( s, p, o ) graph =
    TriplePattern (termFromId s graph) (termFromId p graph) (termFromId o graph)


termFromId : Int -> Graph -> Term
termFromId id graph =
    case Dict.get id graph.id2term of
        Nothing ->
            Debug.crash "BUG: term ID in index but not in dictionary"

        Just s ->
            termFromString s


termFromString : String -> Term
termFromString s =
    case (String.slice 0 1 s) of
        "<" ->
            TermURI (String.slice 1 -1 s)

        "_" ->
            TermBlankNode (String.dropLeft 2 s)

        _ ->
            TermLiteral (literalFromString s)


literalFromString : String -> Literal
literalFromString s =
    Literal s Nothing xsdString



-- MUTATING HELPER FUNCTIONS


insert1 : TriplePattern -> Graph -> Graph
insert1 triple g0 =
    let
        ( s, g1 ) =
            insertTerm triple.subject g0

        ( p, g2 ) =
            insertTerm triple.predicate g1

        ( o, g3 ) =
            insertTerm triple.object g2
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
        { graph | spo = Dict.insert ( s, p ) newO graph.spo }



-- DECODING HELPER FUNCTIONS


fromTriples : List TriplePattern -> Graph
fromTriples triples =
    List.foldr insert1 empty triples



-- ENCODING HELPER FUNCTIONS


mustSerialize : Int -> Dict.Dict Int String -> String
mustSerialize id terms =
    case Dict.get id terms of
        Nothing ->
            Debug.crash "BUG: term ID in index but not in dictionary"

        Just s ->
            s



-- OTHER HELPER FUNCTIONS


maybeToSet : Maybe (Set.Set Int) -> Set.Set Int
maybeToSet m =
    case m of
        Nothing ->
            Set.empty

        Just x ->
            x



-- insert1 -> TriplePattern -> Graph
-- delete1 -> TriplePattern -> Graph
-- insert -> List TriplePattern -> Graph
-- delete -> List TriplePattern -> Graph
-- set/replace -> TriplePattern -> Graph
