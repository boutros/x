module A_Model exposing (Model, valueFromGraph)

import B_Message exposing (Route)
import C_Data as Data
import RDF.RDF as RDF
import RDF.Graph as Graph


type alias Model =
    { route : Route
    , error : String
    , query : String
    , results : Maybe Data.SearchResults
    , graph : Graph.Graph
    }


valueFromGraph : Model -> List RDF.TriplePattern -> String
valueFromGraph model patterns =
    let
        triples =
            Graph.query patterns model.graph
    in
        case triples of
            [] ->
                ""

            [ triple ] ->
                valueFromObject triple.object

            triple :: rest ->
                Debug.crash "expected one value, found several"


valueFromObject : RDF.Term -> String
valueFromObject obj =
    case obj of
        RDF.TermLiteral l ->
            l.value

        _ ->
            ""



-- valuesFromGraph -> List RDF.TriplePattern -> Model -> List String
