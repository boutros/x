module A_Model exposing (Model)

import B_Message exposing (Route)
import C_Data as Data
import RDF.Graph exposing (Graph)


type alias Model =
    { route : Route
    , error : String
    , query : String
    , results : Maybe Data.SearchResults
    , graph : Graph
    }
