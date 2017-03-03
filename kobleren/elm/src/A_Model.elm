module A_Model exposing (Model)

import B_Message exposing (Route)
import C_Data as Data
import RDF.RDF as RDF


type alias Model =
    { route : Route
    , error : String
    , query : String
    , results : Maybe Data.SearchResults
    , graph : List RDF.TriplePattern
    }
