module B_Message exposing (..)

import C_Data as Data
import Http
import Navigation exposing (Location)


type Msg
    = Search String
    | GetResults (Result Http.Error Data.SearchResults)
    | Paginate Int
    | LocationChange Location
    | NavigateTo String


type Route
    = HomeRoute
    | EditAuthorityRoute String String
    | NotFoundRoute
