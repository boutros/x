module B_Message exposing (..)

import C_Data as Data
import RDF.RDF as RDF
import Http
import Navigation exposing (Location)


type Msg
    = Search String
    | GetResults (Result Http.Error Data.SearchResults)
    | Paginate Int
    | LocationChange Location
    | NavigateTo String
    | LoadResource String
    | GetResource (Result Http.Error String)


type Route
    = HomeRoute
    | EditAuthorityRoute String String
    | NotFoundRoute
