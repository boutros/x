module RDF.Graph exposing (Graph)

import RDF.RDF exposing (TriplePattern, Term)
import RDF.Parser exposing (parseNTriples)
import Dict


type alias Graph =
    { bnodeId : Int
    , nodes : Dict.Dict String (Dict.Dict String (List Term))
    }


empty : Graph
empty =
    Graph 0 Dict.empty



--
-- insert1 -> TriplePattern -> Graph
-- delete1 -> TriplePattern -> Graph
-- insert -> List TriplePattern -> Graph
-- delete -> List TriplePattern -> Graph
-- set/replace -> TriplePattern -> Graph
-- fromTriples -> List TriplePattern -> Graph
-- fromNTriples -> String -> Graph
-- toNTriples -> Graph -> String
